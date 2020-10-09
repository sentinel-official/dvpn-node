package wireguard

import (
	"bytes"
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

type WireGuard struct {
	cfg   *Config
	peers int64
}

func NewWireGuard() types.Service {
	return &WireGuard{}
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
	err := exec.Command("wg-quick",
		strings.Split(fmt.Sprintf("up %s", w.cfg.Device), " ")...).Run()
	if err != nil {
		return err
	}

	w.peers = 0
	return nil
}

func (w *WireGuard) Stop() error {
	return exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("down %s", w.cfg.Device), " ")...).Run()
}

func (w *WireGuard) AddPeer(keys ...string) error {
	err := exec.Command("wg", strings.Split(
		fmt.Sprintf("set %s peer %s allowed-ips %s", w.cfg.Device, keys[0], keys[1]), " ")...).Run()
	if err != nil {
		return err
	}

	w.peers = w.peers + 1
	return nil
}

func (w *WireGuard) RemovePeer(keys ...string) error {
	err := exec.Command("wg", strings.Split(
		fmt.Sprintf("set %s peer %s remove", w.cfg.Device, keys[0]), " ")...).Run()
	if err != nil {
		return err
	}

	w.peers = w.peers - 1
	return nil
}

func (w *WireGuard) Peers() ([]types.Peer, error) {
	output, err := exec.Command("wg",
		strings.Split(fmt.Sprintf("show %s transfer", w.cfg.Device), " ")...).Output()
	if err != nil {
		return nil, err
	}

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

func (w *WireGuard) PeersCount() int64 {
	return w.peers
}
