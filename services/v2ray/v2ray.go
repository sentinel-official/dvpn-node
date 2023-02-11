package v2ray

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	proxymancommand "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	statscommand "github.com/v2fly/v2ray-core/v5/app/stats/command"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v2raytypes "github.com/sentinel-official/dvpn-node/services/v2ray/types"
	"github.com/sentinel-official/dvpn-node/types"
)

const (
	InfoLen = 2 + 1
)

var (
	_ types.Service = (*V2Ray)(nil)
)

type V2Ray struct {
	info   []byte
	cmd    *exec.Cmd
	config *v2raytypes.Config
	peers  *v2raytypes.Peers
}

func NewV2Ray() *V2Ray {
	return &V2Ray{
		info:   make([]byte, InfoLen),
		cmd:    nil,
		config: v2raytypes.NewConfig(),
		peers:  v2raytypes.NewPeers(),
	}
}

func (s *V2Ray) configFilePath() string {
	return filepath.Join(os.TempDir(), "v2ray_config.json")
}

func (s *V2Ray) Type() uint64 {
	return v2raytypes.Type
}

func (s *V2Ray) Info() []byte {
	return s.info
}

func (s *V2Ray) Init(home string) (err error) {
	v := viper.New()
	v.SetConfigFile(filepath.Join(home, v2raytypes.ConfigFileName))

	s.config, err = v2raytypes.ReadInConfig(v)
	if err != nil {
		return err
	}

	t, err := template.New("config_v2ray_json").Parse(configTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, s.config); err != nil {
		return err
	}
	if err = os.WriteFile(s.configFilePath(), buf.Bytes(), 0600); err != nil {
		return err
	}

	binary.BigEndian.PutUint16(s.info[0:], s.config.VMess.ListenPort)
	transport := v2raytypes.NewTransportFromString(s.config.VMess.Transport)
	s.info[2] = transport.Byte()

	return nil
}

func (s *V2Ray) Start() error {
	s.cmd = exec.Command("v2ray", strings.Split(
		fmt.Sprintf("run --config %s", s.configFilePath()), " ")...)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	return s.cmd.Start()
}

func (s *V2Ray) Stop() error {
	if s.cmd == nil {
		return errors.New("command is nil")
	}

	return s.cmd.Process.Kill()
}

func (s *V2Ray) clientConn() (*grpc.ClientConn, error) {
	target := "127.0.0.1:23"
	return grpc.Dial(
		target,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func (s *V2Ray) handlerServiceClient() (proxymancommand.HandlerServiceClient, error) {
	conn, err := s.clientConn()
	if err != nil {
		return nil, err
	}

	client := proxymancommand.NewHandlerServiceClient(conn)
	return client, nil
}

func (s *V2Ray) statsServiceClient() (statscommand.StatsServiceClient, error) {
	conn, err := s.clientConn()
	if err != nil {
		return nil, err
	}

	client := statscommand.NewStatsServiceClient(conn)
	return client, nil
}

func (s *V2Ray) AddPeer(data []byte) (result []byte, err error) {
	if len(data) != 1+16 {
		return nil, errors.New("data length must be 17 bytes")
	}

	client, err := s.handlerServiceClient()
	if err != nil {
		return nil, err
	}

	var (
		proxy  = v2raytypes.Proxy(data[0])
		uid, _ = uuid.ParseBytes(data[1:])
	)

	req := &proxymancommand.AlterInboundRequest{
		Tag: proxy.Tag(),
		Operation: serial.ToTypedMessage(
			&proxymancommand.AddUserOperation{
				User: &protocol.User{
					Level:   0,
					Email:   uid.String(),
					Account: proxy.Account(uid),
				},
			},
		),
	}

	_, err = client.AlterInbound(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	s.peers.Put(
		v2raytypes.Peer{
			Identity: uid.String(),
		},
	)

	return result, nil
}

func (s *V2Ray) RemovePeer(data []byte) error {
	if len(data) != 1+16 {
		return errors.New("data length must be 17 bytes")
	}

	client, err := s.handlerServiceClient()
	if err != nil {
		return err
	}

	var (
		proxy  = v2raytypes.Proxy(data[0])
		uid, _ = uuid.ParseBytes(data[1:])
	)

	req := &proxymancommand.AlterInboundRequest{
		Tag: proxy.Tag(),
		Operation: serial.ToTypedMessage(
			&proxymancommand.RemoveUserOperation{
				Email: uid.String(),
			},
		),
	}

	_, err = client.AlterInbound(context.TODO(), req)
	if err != nil {
		return err
	}

	s.peers.Delete(uid.String())

	return nil
}

func (s *V2Ray) Peers() ([]types.Peer, error) {
	client, err := s.statsServiceClient()
	if err != nil {
		return nil, err
	}

	req := &statscommand.QueryStatsRequest{
		Reset_: false,
		Patterns: []string{
			"user>>>",
		},
		Regexp: false,
	}

	res, err := client.QueryStats(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	var (
		upLink   = make(map[string]int64)
		downLink = make(map[string]int64)
	)

	for _, stat := range res.GetStat() {
		name := strings.Split(stat.GetName(), ">>>")
		if len(name) != 4 {
			continue
		}

		var (
			link = name[3]
			uid  = name[1]
		)

		if _, ok := upLink[uid]; !ok {
			upLink[uid] = 0
		}
		if _, ok := downLink[uid]; !ok {
			downLink[uid] = 0
		}

		value := stat.GetValue()
		if link == "uplink" {
			upLink[uid] = value
		} else if link == "downlink" {
			downLink[uid] = value
		}
	}

	var items []types.Peer
	for key := range upLink {
		items = append(
			items,
			types.Peer{
				Key:      key,
				Upload:   upLink[key],
				Download: downLink[key],
			},
		)
	}

	return items, nil
}

func (s *V2Ray) PeersCount() int {
	return s.peers.Len()
}
