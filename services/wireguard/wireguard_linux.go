// +build linux
// +build !openwrt

package wireguard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	wgtypes "github.com/sentinel-official/dvpn-node/services/wireguard/types"
)

func (w *WireGuard) Init(_ string) (err error) {
	t, err := template.New("").Parse(wgtypes.ConfigTemplate)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, w.cfg); err != nil {
		return err
	}

	path := fmt.Sprintf("/etc/wireguard/%s.conf", w.cfg.IFace)
	if err := ioutil.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}

func (w *WireGuard) PreUp() error {
	return nil
}

func (w *WireGuard) Up() error {
	cmd := exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("up %s", w.cfg.IFace), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (w *WireGuard) PreDown() error {
	return nil
}

func (w *WireGuard) Down() error {
	cmd := exec.Command("wg-quick", strings.Split(
		fmt.Sprintf("down %s", w.cfg.IFace), " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
