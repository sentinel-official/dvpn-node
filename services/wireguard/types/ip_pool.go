package types

import (
	"fmt"
	"net"
	"sync"
)

type IPv4Pool struct {
	Net       *net.IPNet
	current   IPv4
	available []IPv4
	reserved  map[IPv4]bool
	mutex     *sync.Mutex
}

func NewIPv4Pool(ip net.IP, ipNet *net.IPNet) *IPv4Pool {
	return &IPv4Pool{
		Net:      ipNet,
		current:  NewIPv4FromIP(ip),
		reserved: make(map[IPv4]bool),
		mutex:    &sync.Mutex{},
	}
}

func (p *IPv4Pool) Get() (ip IPv4, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.available) == 0 {
		if p.current[3] == 0x00 || p.current[3] == 0xff {
			p.current = p.current.Next()
		}
		if !p.Net.Contains(p.current.IP()) {
			return ip, fmt.Errorf("ipv4 pool is pull")
		}

		ip, p.current = p.current, p.current.Next()
	} else {
		ip, p.available = p.available[0], p.available[1:]
	}

	p.reserved[ip] = true
	return ip, nil
}

func (p *IPv4Pool) Release(ip IPv4) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.reserved[ip] {
		delete(p.reserved, ip)
		p.available = append(p.available, ip)
	}
}

func NewIPv4PoolFromCIDR(s string) (*IPv4Pool, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}

	return NewIPv4Pool(ip, ipNet), nil
}

type IPv6Pool struct {
	Net       *net.IPNet
	current   IPv6
	available []IPv6
	reserved  map[IPv6]bool
	mutex     *sync.Mutex
}

func NewIPv6Pool(ip net.IP, ipNet *net.IPNet) *IPv6Pool {
	return &IPv6Pool{
		Net:      ipNet,
		current:  NewIPv6FromIP(ip),
		reserved: make(map[IPv6]bool),
		mutex:    &sync.Mutex{},
	}
}

func (p *IPv6Pool) Get() (ip IPv6, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.available) == 0 {
		if !p.Net.Contains(p.current.IP()) {
			return ip, fmt.Errorf("ipv6 pool is pull")
		}

		ip, p.current = p.current, p.current.Next()
	} else {
		ip, p.available = p.available[0], p.available[1:]
	}

	p.reserved[ip] = true
	return ip, nil
}

func (p *IPv6Pool) Release(ip IPv6) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.reserved[ip] {
		delete(p.reserved, ip)
		p.available = append(p.available, ip)
	}
}

func NewIPv6PoolFromCIDR(s string) (*IPv6Pool, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}

	return NewIPv6Pool(ip, ipNet), nil
}

type IPPool struct {
	V4 *IPv4Pool
	V6 *IPv6Pool
}

func NewIPPool(v4 *IPv4Pool, v6 *IPv6Pool) *IPPool {
	return &IPPool{
		V4: v4,
		V6: v6,
	}
}

func (p *IPPool) Get() (IPv4, IPv6, error) {
	v4, err := p.V4.Get()
	if err != nil {
		return IPv4{}, IPv6{}, err
	}

	v6, err := p.V6.Get()
	if err != nil {
		p.V4.Release(v4)
		return IPv4{}, IPv6{}, err
	}

	return v4, v6, nil
}

func (p *IPPool) Release(v4 IPv4, v6 IPv6) {
	p.V4.Release(v4)
	p.V6.Release(v6)
}
