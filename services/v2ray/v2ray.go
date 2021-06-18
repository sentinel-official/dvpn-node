package v2ray

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/itchyny/gojq"
	"github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	statsCommand "github.com/v2fly/v2ray-core/v4/app/stats/command"
	"google.golang.org/grpc"

	vrtypes "github.com/sentinel-official/dvpn-node/services/v2ray/types"
	"github.com/sentinel-official/dvpn-node/types"
)

const (
	InfoLen = 2 + 2 // {vmess-port}{trojan-port}
)

type V2Ray struct {
	cmd   *exec.Cmd
	info  []byte
	cfg   *vrtypes.Config
	peers map[string]bool
}

func NewV2Ray() types.Service {
	return &V2Ray{
		info:  make([]byte, InfoLen),
		cfg:   vrtypes.NewConfig(),
		peers: make(map[string]bool),
	}
}

func (v *V2Ray) Type() uint64 {
	return vrtypes.Type
}

func (v *V2Ray) Init(homeDir string) error {
	configInputPath := filepath.Join(homeDir, vrtypes.ConfigInputFileName)
	if err := v.cfg.LoadFromPath(configInputPath); err != nil {
		return err
	}
	v.cfg.HomeDir = homeDir

	t, err := template.New("").Parse(configTemplate)
	if err != nil {
		return err
	}

	var configBuffer bytes.Buffer
	if err := t.Execute(&configBuffer, v.cfg); err != nil {
		return err
	}

	var configString = configBuffer.String()

	for i, protocol := range vrtypes.SupportedProtocols {
		var mask uint16 = 1 << i
		if v.cfg.ProtocolSelection&mask != 0 {
			continue
		}
		deleteInboundQuery := fmt.Sprintf("del(.inbounds[] | select(.tag == \"%s\"))", protocol.Tag())
		configString, err = runJQ(configString, deleteInboundQuery)
		if err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filepath.Join(v.cfg.HomeDir, vrtypes.ConfigFileName), []byte(configString), 0600); err != nil {
		return err
	}

	binary.BigEndian.PutUint16(v.info[:2], v.cfg.VMessPort)
	binary.BigEndian.PutUint16(v.info[2:], v.cfg.TrojanPort)

	return nil
}

func (v *V2Ray) Info() []byte {
	return v.info
}

func (v *V2Ray) Start() error {
	v.cmd = exec.Command("v2ray", strings.Split(
		fmt.Sprintf("--config %s", filepath.Join(v.cfg.HomeDir, vrtypes.ConfigFileName)), " ")...)
	v.cmd.Stdout = os.Stdout
	v.cmd.Stderr = os.Stderr

	return v.cmd.Start()
}

func (v *V2Ray) Stop() error {
	return v.cmd.Process.Kill()
}

func (v *V2Ray) AddPeer(data []byte) ([]byte, error) {
	// {uuid}@sentinel.co
	email := string(data)

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", v.cfg.APIPort), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	hsClient := command.NewHandlerServiceClient(cmdConn)

	for i, protocol := range vrtypes.SupportedProtocols {
		var mask uint16 = 1 << i
		if v.cfg.ProtocolSelection&mask == 0 {
			continue
		}
		err := protocol.AddPeer(hsClient, data)
		if err != nil {
			return nil, err
		}
	}

	v.peers[email] = true

	return nil, nil
}

func (v *V2Ray) RemovePeer(data []byte) error {
	email := string(data)

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", v.cfg.APIPort), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}

	hsClient := command.NewHandlerServiceClient(cmdConn)
	for i, protocol := range vrtypes.SupportedProtocols {
		var mask uint16 = 1 << i
		if v.cfg.ProtocolSelection&mask == 0 {
			continue
		}
		err := protocol.RemovePeer(hsClient, data)
		if err != nil {
			return nil
		}
	}

	if _, ok := v.peers[email]; ok {
		delete(v.peers, email)
	}

	return nil
}

func (v *V2Ray) Peers() ([]types.Peer, error) {
	var items []types.Peer

	cmdConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", v.cfg.APIPort), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	ssClient := statsCommand.NewStatsServiceClient(cmdConn)

	for email, _ := range v.peers {
		downloadStat, err := ssClient.GetStats(context.Background(), &statsCommand.GetStatsRequest{
			Name: fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email),
		})
		if err != nil {
			return nil, err
		}
		download := downloadStat.Stat.Value

		uploadStat, err := ssClient.GetStats(context.Background(), &statsCommand.GetStatsRequest{
			Name: fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email),
		})
		if err != nil {
			return nil, err
		}
		upload := uploadStat.Stat.Value

		items = append(items, types.Peer{
			Key:      email,
			Download: download,
			Upload:   upload,
		})
	}
	return items, nil
}

func (v *V2Ray) PeersCount() int {
	return len(v.peers)
}

func runJQ(input string, query string) (string, error) {
	_query, err := gojq.Parse(query)
	if err != nil {
		return "", err
	}
	var jsonObject interface{}
	err = json.Unmarshal([]byte(input), &jsonObject)
	if err != nil {
		return "", err
	}
	iter := _query.Run(jsonObject)
	var output = ""
	for {
		result, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := result.(error); ok {
			return "", err
		}
		s, _ := json.Marshal(result)
		output += string(s)
	}
	return output, nil
}
