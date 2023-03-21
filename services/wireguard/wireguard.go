package wireguard

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	wgtypes "github.com/sentinel-official/dvpn-node/services/wireguard/types"
	"github.com/sentinel-official/dvpn-node/types"
)

const (
	InfoLen = 2 + 32
)

var (
	_ types.Service = (*WireGuard)(nil)
)

type WireGuard struct {
	info   []byte
	config *wgtypes.Config
	peers  *wgtypes.Peers
	pool   *wgtypes.IPPool
}

func NewWireGuard(pool *wgtypes.IPPool) types.Service {
	return &WireGuard{
		pool:   pool,
		config: wgtypes.NewConfig(),
		info:   make([]byte, InfoLen),
		peers:  wgtypes.NewPeers(),
	}
}

func (s *WireGuard) Type() uint64 {
	return wgtypes.Type
}

func (s *WireGuard) Init(home string) (err error) {
	v := viper.New()
	v.SetConfigFile(filepath.Join(home, wgtypes.ConfigFileName))

	s.config, err = wgtypes.ReadInConfig(v)
	if err != nil {
		return err
	}
	if err = s.config.Validate(); err != nil {
		return err
	}

	t, err := template.New("wireguard_conf").Parse(configTemplate)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err = t.Execute(&buffer, s.config); err != nil {
		return err
	}

	path := fmt.Sprintf("/etc/wireguard/%s.conf", s.config.Interface)
	if err = os.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		return err
	}

	key, err := wgtypes.KeyFromString(s.config.PrivateKey)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint16(s.info[:2], s.config.ListenPort)
	copy(s.info[2:], key.Public().Bytes())

	return nil
}

func (s *WireGuard) Info() []byte {
	return s.info
}

func (s *WireGuard) Start() error {
	cmd := exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("up %s", s.config.Interface), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *WireGuard) Stop() error {
	cmd := exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("down %s", s.config.Interface), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *WireGuard) AddPeer(data []byte) (result []byte, err error) {
	identity := base64.StdEncoding.EncodeToString(data)

	v4, v6, err := s.pool.Get()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			s.pool.Release(v4, v6)
		}
	}()

	cmd := exec.Command("wg", strings.Split(
		fmt.Sprintf(`set %s peer %s allowed-ips %s/32,%s/128`,
			s.config.Interface, identity, v4.IP(), v6.IP()), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return nil, err
	}

	s.peers.Put(
		wgtypes.Peer{
			Identity: identity,
			IPv4:     v4,
			IPv6:     v6,
		},
	)

	result = append(result, v4.Bytes()...)
	result = append(result, v6.Bytes()...)
	return result, nil
}

func (s *WireGuard) HasPeer(data []byte) bool {
	var (
		identity = base64.StdEncoding.EncodeToString(data)
		peer     = s.peers.Get(identity)
	)

	return !peer.Empty()
}

func (s *WireGuard) RemovePeer(data []byte) error {
	identity := base64.StdEncoding.EncodeToString(data)

	cmd := exec.Command("wg", strings.Split(
		fmt.Sprintf(`set %s peer %s remove`,
			s.config.Interface, identity), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if v := s.peers.Get(identity); !v.Empty() {
		s.peers.Delete(v.Identity)
		s.pool.Release(v.IPv4, v.IPv6)
	}

	return nil
}

func (s *WireGuard) Peers() (items []types.Peer, err error) {
	output, err := exec.Command("wg", strings.Split(
		fmt.Sprintf("show %s transfer", s.config.Interface), " ")...).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		columns := strings.Split(line, "\t")
		if len(columns) != 3 {
			continue
		}

		upload, err := strconv.ParseInt(columns[1], 10, 64)
		if err != nil {
			return nil, err
		}

		download, err := strconv.ParseInt(columns[2], 10, 64)
		if err != nil {
			return nil, err
		}

		items = append(items,
			types.Peer{
				Key:      columns[0],
				Upload:   upload,
				Download: download,
			},
		)
	}

	return items, nil
}

func (s *WireGuard) PeerCount() int {
	return s.peers.Len()
}
