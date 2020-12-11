package wireguard

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	node "github.com/sentinel-official/hub/x/node/types"

	"github.com/sentinel-official/dvpn-node/types"
)

var _ types.Service = (*WireGuard)(nil)

type Peer struct {
	IPv4 types.IPv4
	IPv6 types.IPv6
}

type WireGuard struct {
	cfg   *Config
	ipp   *types.IPPool
	peers map[string]Peer
}

func NewWireGuard(ipp *types.IPPool) types.Service {
	return &WireGuard{
		ipp:   ipp,
		peers: make(map[string]Peer),
	}
}

func (w *WireGuard) Type() node.Category {
	return node.CategoryWireGuard
}

func (w *WireGuard) Initialize(home string) error {
	w.cfg = NewConfig()
	if err := w.cfg.LoadFromPath(filepath.Join(home, "wireguard.toml")); err != nil {
		return err
	}

	t, err := template.New("config").Parse(serverConfigTemplate)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, w.cfg); err != nil {
		return err
	}

	configFilePath := fmt.Sprintf("/etc/wireguard/%s.conf", w.cfg.Device)
	return ioutil.WriteFile(configFilePath, buffer.Bytes(), 0600)
}

func (w *WireGuard) Start() error {
	err := exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("up %s", w.cfg.Device), " ")...).Run()
	if err != nil {
		return err
	}

	return nil
}

func (w *WireGuard) Stop() error {
	return exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("down %s", w.cfg.Device), " ")...).Run()
}

func (w *WireGuard) AddPeer(key []byte) (result []byte, err error) {
	publicKey := base64.StdEncoding.EncodeToString(key)

	v4, v6, err := w.ipp.Get()
	if err != nil {
		return nil, err
	}

	err = exec.Command("wg", strings.Split(
		fmt.Sprintf(`set %s peer %s allowed-ips %s/32,%s/128`,
			w.cfg.Device, publicKey, v4.IP(), v6.IP()), " ")...).Run()
	if err != nil {
		w.ipp.Release(v4, v6)
		return nil, err
	}

	w.peers[publicKey] = Peer{v4, v6}

	result = append(result, v4.Bytes()...)
	result = append(result, v6.Bytes()...)
	return result, nil
}

func (w *WireGuard) RemovePeer(key []byte) error {
	publicKey := base64.StdEncoding.EncodeToString(key)

	err := exec.Command("wg", strings.Split(
		fmt.Sprintf(`set %s peer %s remove`,
			w.cfg.Device, publicKey), " ")...).Run()
	if err != nil {
		return err
	}

	peer, ok := w.peers[publicKey]
	if ok {
		delete(w.peers, publicKey)
		w.ipp.Release(peer.IPv4, peer.IPv6)
	}

	return nil
}

func (w *WireGuard) Peers() ([]types.Peer, error) {
	output, err := exec.Command("wg", strings.Split(
		fmt.Sprintf("show %s transfer", w.cfg.Device), " ")...).Output()
	if err != nil {
		return nil, err
	}

	// nolint: prealloc
	var (
		lines = strings.Split(string(output), "\n")
		items []types.Peer
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

		items = append(items, types.Peer{
			Identity: columns[0],
			Upload:   upload,
			Download: download,
		})
	}

	return items, nil
}

func (w *WireGuard) PeersCount() int {
	return len(w.peers)
}
