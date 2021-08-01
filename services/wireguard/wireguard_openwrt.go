// +build linux
// +build openwrt

package wireguard

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (w *WireGuard) Init(_ string) (err error) {
	return nil
}

func (w *WireGuard) PreUp() error {
	commands := [][]string{
		{"uci", fmt.Sprintf(`set network.%s=interface`, w.cfg.IFace)},
		{"uci", fmt.Sprintf(`set network.%s.proto=wireguard`, w.cfg.IFace)},
		{"uci", fmt.Sprintf(`set network.%s.private_key=%s`, w.cfg.IFace, w.cfg.PrivateKey)},
		{"uci", fmt.Sprintf(`set network.%s.listen_port=%d`, w.cfg.IFace, w.cfg.ListenPort)},
		{"uci", fmt.Sprintf(`add_list network.%s.addresses=%s`, w.cfg.IFace, w.cfg.IPv4CIDR)},
		{"uci", fmt.Sprintf(`add_list network.%s.addresses=%s`, w.cfg.IFace, w.cfg.IPv6CIDR)},

		{"uci", fmt.Sprintf(`add firewall zone`)},
		{"uci", fmt.Sprintf(`set firewall.@zone[-1].name=wireguard`)},
		{"uci", fmt.Sprintf(`set firewall.@zone[-1].input=ACCEPT`)},
		{"uci", fmt.Sprintf(`set firewall.@zone[-1].forward=ACCEPT`)},
		{"uci", fmt.Sprintf(`set firewall.@zone[-1].output=ACCEPT`)},
		{"uci", fmt.Sprintf(`add_list firewall.@zone[-1].network=%s`, w.cfg.IFace)},

		{"uci", fmt.Sprintf(`add firewall forwarding`)},
		{"uci", fmt.Sprintf(`set firewall.@forwarding[-1].src=wireguard`)},
		{"uci", fmt.Sprintf(`set firewall.@forwarding[-1].dest=%s`, w.cfg.IFaceWAN)},

		{"uci", fmt.Sprintf(`add firewall forwarding`)},
		{"uci", fmt.Sprintf(`set firewall.@forwarding[-1].src=%s`, w.cfg.IFaceWAN)},
		{"uci", fmt.Sprintf(`set firewall.@forwarding[-1].dest=wireguard`)},
	}

	for _, item := range commands {
		fmt.Println(item[0], item[1])

		cmd := exec.Command(item[0], strings.Split(item[1], " ")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (w *WireGuard) Up() error {
	commands := [][]string{
		{"/etc/init.d/network", fmt.Sprintf(`restart`)},
		{"sleep", fmt.Sprintf(`%ds`, 5)},
		{"/etc/init.d/firewall", fmt.Sprintf(`restart`)},
	}

	for _, item := range commands {
		fmt.Println(item[0], item[1])

		cmd := exec.Command(item[0], strings.Split(item[1], " ")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (w *WireGuard) PreDown() error {
	return nil
}

func (w *WireGuard) Down() error {
	return nil
}
