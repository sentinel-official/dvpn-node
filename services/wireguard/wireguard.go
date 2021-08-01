package wireguard

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	tmlog "github.com/tendermint/tendermint/libs/log"

	wgtypes "github.com/sentinel-official/dvpn-node/services/wireguard/types"
	nodetypes "github.com/sentinel-official/dvpn-node/types"
)

const (
	InfoLen = 2 + 32
)

var (
	_ nodetypes.Service = (*WireGuard)(nil)
)

type WireGuard struct {
	info  []byte
	cfg   *wgtypes.Config
	log   tmlog.Logger
	peers *wgtypes.Peers
	pool  *wgtypes.IPPool
}

func NewWireGuard() nodetypes.Service {
	return &WireGuard{
		cfg:   wgtypes.NewConfig(),
		info:  make([]byte, InfoLen),
		peers: wgtypes.NewPeers(),
	}
}

func (w *WireGuard) Type() uint64                                { return wgtypes.Type }
func (w *WireGuard) Info() []byte                                { return w.info }
func (w *WireGuard) WithLogger(v tmlog.Logger) nodetypes.Service { w.log = v; return w }

func (w *WireGuard) PreInit(home string) (err error) {
	v := viper.New()
	v.SetConfigFile(filepath.Join(home, wgtypes.ConfigFileName))

	w.cfg, err = wgtypes.ReadInConfig(v)
	if err != nil {
		return err
	}

	if err := w.cfg.Validate(); err != nil {
		return err
	}

	w.log.Info("Creating IPv4 pool", "CIDR", w.cfg.IPv4CIDR, "type", w.Type())
	v4, err := wgtypes.NewIPv4PoolFromCIDR(w.cfg.IPv4CIDR)
	if err != nil {
		return err
	}

	w.log.Info("Creating IPv6 pool", "CIDR", w.cfg.IPv6CIDR, "type", w.Type())
	v6, err := wgtypes.NewIPv6PoolFromCIDR(w.cfg.IPv6CIDR)
	if err != nil {
		return err
	}

	w.pool = wgtypes.NewIPPool(v4, v6)
	return nil
}

func (w *WireGuard) PostInit(_ string) error {
	key, err := wgtypes.KeyFromString(w.cfg.PrivateKey)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint16(w.info[:2], w.cfg.ListenPort)
	copy(w.info[2:], key.Public().Bytes())

	return nil
}

func (w *WireGuard) PostUp() error {
	commands := [][]string{
		{"iptables", fmt.Sprintf(`-A FORWARD -i %s -j ACCEPT`, w.cfg.IFace)},
		{"iptables", fmt.Sprintf(`-A POSTROUTING -t nat -o %s -j MASQUERADE`, w.cfg.IFaceWAN)},
		{"ip6tables", fmt.Sprintf(`-A FORWARD -i %s -j ACCEPT`, w.cfg.IFace)},
		{"ip6tables", fmt.Sprintf(`-A POSTROUTING -t nat -o %s -j MASQUERADE`, w.cfg.IFaceWAN)},
	}

	for _, item := range commands {
		cmd := exec.Command(item[0], strings.Split(item[1], " ")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (w *WireGuard) PostDown() error {
	commands := [][]string{
		{"iptables", fmt.Sprintf(`-D FORWARD -i %s -j ACCEPT`, w.cfg.IFace)},
		{"iptables", fmt.Sprintf(`-D POSTROUTING -t nat -o %s -j MASQUERADE`, w.cfg.IFaceWAN)},
		{"ip6tables", fmt.Sprintf(`-D FORWARD -i %s -j ACCEPT`, w.cfg.IFace)},
		{"ip6tables", fmt.Sprintf(`-D POSTROUTING -t nat -o %s -j MASQUERADE`, w.cfg.IFaceWAN)},
	}

	for _, item := range commands {
		cmd := exec.Command(item[0], strings.Split(item[1], " ")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (w *WireGuard) Start() error {
	if err := w.PreUp(); err != nil {
		return err
	}
	if err := w.Up(); err != nil {
		return err
	}
	if err := w.PostUp(); err != nil {
		return err
	}

	return nil
}

func (w *WireGuard) Stop() error {
	if err := w.PreDown(); err != nil {
		return err
	}
	if err := w.Down(); err != nil {
		return err
	}
	if err := w.PostDown(); err != nil {
		return err
	}

	return nil
}

func (w *WireGuard) AddPeer(data []byte) (result []byte, err error) {
	identity := base64.StdEncoding.EncodeToString(data)

	v4, v6, err := w.pool.Get()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("wg", strings.Split(
		fmt.Sprintf(`set %s peer %s allowed-ips %s/32,%s/128`,
			w.cfg.IFace, identity, v4.IP(), v6.IP()), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		w.pool.Release(v4, v6)
		return nil, err
	}

	if v := w.peers.Get(identity); !v.Empty() {
		w.peers.Delete(v.Identity)
		w.pool.Release(v.IPv4, v.IPv6)
	}

	w.peers.Put(
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

func (w *WireGuard) RemovePeer(data []byte) error {
	identity := base64.StdEncoding.EncodeToString(data)

	cmd := exec.Command("wg", strings.Split(
		fmt.Sprintf(`set %s peer %s remove`,
			w.cfg.IFace, identity), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if v := w.peers.Get(identity); !v.Empty() {
		w.peers.Delete(v.Identity)
		w.pool.Release(v.IPv4, v.IPv6)
	}

	return nil
}

func (w *WireGuard) Peers() ([]nodetypes.Peer, error) {
	output, err := exec.Command("wg", strings.Split(
		fmt.Sprintf("show %s transfer", w.cfg.IFace), " ")...).Output()
	if err != nil {
		return nil, err
	}

	// nolint: prealloc
	var (
		items []nodetypes.Peer
		lines = strings.Split(string(output), "\n")
	)

	for _, line := range lines {
		columns := strings.Split(line, "\t")
		if len(columns) != 3 {
			continue
		}

		download, err := strconv.ParseInt(columns[1], 10, 64)
		if err != nil {
			return nil, err
		}

		upload, err := strconv.ParseInt(columns[2], 10, 64)
		if err != nil {
			return nil, err
		}

		items = append(items,
			nodetypes.Peer{
				Key:      columns[0],
				Upload:   upload,
				Download: download,
			},
		)
	}

	return items, nil
}

func (w *WireGuard) PeersLen() int {
	return w.peers.Len()
}
