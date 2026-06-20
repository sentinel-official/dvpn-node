package core

import (
	"fmt"
	"net/netip"
	"sort"

	"github.com/sentinel-official/sentinel-go-sdk/amneziawg"
	"github.com/sentinel-official/sentinel-go-sdk/hysteria2"
	sdknetip "github.com/sentinel-official/sentinel-go-sdk/libs/netip"
	"github.com/sentinel-official/sentinel-go-sdk/openvpn"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/v2ray"
	"github.com/sentinel-official/sentinel-go-sdk/wireguard"
	"github.com/sentinel-official/sentinel-go-sdk/xray"

	"github.com/sentinel-official/sentinel-dvpnx/config"
)

// servicePorts returns the host-bound ports for the given service type extracted
// from the service's persisted config.  Port strings are parsed via the SDK
// netip helper; the bound (outbound) port value is used for each service.
func servicePorts(t types.ServiceType, cfg *config.Config) ([]uint16, error) {
	switch t {
	case types.ServiceTypeWireGuard:
		sc := cfg.Services[t].(*wireguard.ServerConfig)
		p := sc.OutPort()
		if p == 0 {
			return nil, fmt.Errorf("wireguard: invalid port %q", sc.Port)
		}
		return []uint16{p}, nil

	case types.ServiceTypeAmneziaWG:
		sc := cfg.Services[t].(*amneziawg.ServerConfig)
		p := sc.OutPort()
		if p == 0 {
			return nil, fmt.Errorf("amneziawg: invalid port %q", sc.Port)
		}
		return []uint16{p}, nil

	case types.ServiceTypeOpenVPN:
		sc := cfg.Services[t].(*openvpn.ServerConfig)
		p := sc.OutPort()
		if p == 0 {
			return nil, fmt.Errorf("openvpn: invalid port %q", sc.Port)
		}
		return []uint16{p}, nil

	case types.ServiceTypeHysteria2:
		sc := cfg.Services[t].(*hysteria2.ServerConfig)
		var ports []uint16
		for _, p := range []uint16{sc.Port, sc.AuthPort, sc.StatsPort} {
			if p == 0 {
				continue
			}
			ports = append(ports, p)
		}
		return ports, nil

	case types.ServiceTypeV2Ray:
		sc := cfg.Services[t].(*v2ray.ServerConfig)
		var ports []uint16
		for _, inbound := range sc.Inbounds {
			port, err := sdknetip.NewPortFromString(inbound.Port)
			if err != nil {
				return nil, fmt.Errorf("v2ray: parsing inbound port %q: %w", inbound.Port, err)
			}
			if port == nil {
				return nil, fmt.Errorf("v2ray: empty inbound port")
			}
			ports = append(ports, port.OutFrom)
		}
		return ports, nil

	case types.ServiceTypeXray:
		sc := cfg.Services[t].(*xray.ServerConfig)
		var ports []uint16
		for _, inbound := range sc.Inbounds {
			port, err := sdknetip.NewPortFromString(inbound.Port)
			if err != nil {
				return nil, fmt.Errorf("xray: parsing inbound port %q: %w", inbound.Port, err)
			}
			if port == nil {
				return nil, fmt.Errorf("xray: empty inbound port")
			}
			ports = append(ports, port.OutFrom)
		}
		return ports, nil

	default:
		return nil, fmt.Errorf("unsupported service type %q", t)
	}
}

// serviceSubnet returns the IPv4 subnet for the given service type when one
// exists (WireGuard, AmneziaWG, OpenVPN).  Returns (zero, false, nil) for
// service types that do not bind a subnet (Hysteria2, V2Ray, Xray).
func serviceSubnet(t types.ServiceType, cfg *config.Config) (netip.Prefix, bool, error) {
	var ipv4Addr string

	switch t {
	case types.ServiceTypeWireGuard:
		ipv4Addr = cfg.Services[t].(*wireguard.ServerConfig).IPv4Addr
	case types.ServiceTypeAmneziaWG:
		ipv4Addr = cfg.Services[t].(*amneziawg.ServerConfig).IPv4Addr
	case types.ServiceTypeOpenVPN:
		ipv4Addr = cfg.Services[t].(*openvpn.ServerConfig).IPv4Addr
	default:
		return netip.Prefix{}, false, nil
	}

	if ipv4Addr == "" {
		return netip.Prefix{}, false, nil
	}

	prefix, err := netip.ParsePrefix(ipv4Addr)
	if err != nil {
		return netip.Prefix{}, false, fmt.Errorf("parsing ipv4_addr %q for service %q: %w", ipv4Addr, t, err)
	}

	return prefix.Masked(), true, nil
}

// conflictingServices iterates the enabled set in deterministic ascending order
// (by ServiceType byte value), accumulates used ports and subnets, and returns
// a map of service types whose ports or subnet would collide with an earlier
// (already-accepted) service.  The map is empty when there are no conflicts.
func conflictingServices(set []types.ServiceType, cfg *config.Config) map[types.ServiceType]error {
	// Sort ascending by byte value for deterministic resolution order.
	sorted := make([]types.ServiceType, len(set))
	copy(sorted, set)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	usedPorts := make(map[uint16]types.ServiceType)
	usedSubnets := make(map[netip.Prefix]types.ServiceType)
	result := make(map[types.ServiceType]error)

	for _, t := range sorted {
		ports, err := servicePorts(t, cfg)
		if err != nil {
			result[t] = fmt.Errorf("extracting ports for service %q: %w", t, err)
			continue
		}

		subnet, hasSubnet, err := serviceSubnet(t, cfg)
		if err != nil {
			result[t] = fmt.Errorf("extracting subnet for service %q: %w", t, err)
			continue
		}

		// Check for port collisions against already-accepted services.
		var collision error
		for _, p := range ports {
			if other, occupied := usedPorts[p]; occupied {
				collision = fmt.Errorf("port %d collides with %q", p, other)
				break
			}
		}

		// Check for subnet collision if no port collision found yet.
		if collision == nil && hasSubnet {
			if other, occupied := usedSubnets[subnet]; occupied {
				collision = fmt.Errorf("subnet %s collides with %q", subnet, other)
			}
		}

		if collision != nil {
			result[t] = collision
			continue
		}

		// No collision — register this service's ports and subnet.
		for _, p := range ports {
			usedPorts[p] = t
		}
		if hasSubnet {
			usedSubnets[subnet] = t
		}
	}

	return result
}
