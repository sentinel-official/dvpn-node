package types

import (
	"fmt"
	"net"
	"sync"
)

type IPv4Pool struct {
	ip       IPv4
	ipNet    *net.IPNet
	reserved map[IPv4]bool
	mutex    *sync.Mutex
}

func NewIPv4Pool(ip net.IP, ipNet *net.IPNet) *IPv4Pool {
	return &IPv4Pool{
		ipNet:    ipNet,
		ip:       NewIPv4FromIP(ip),
		reserved: make(map[IPv4]bool),
		mutex:    &sync.Mutex{},
	}
}

func (p *IPv4Pool) Get() (IPv4, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for ip := p.ip; p.ipNet.Contains(ip.IP()); {
		if _, ok := p.reserved[ip]; !ok {
			p.reserved[ip] = true
			return ip, nil
		}

		ip = ip.Next()
	}

	return IPv4{}, fmt.Errorf("ipv4 pool is full")
}

func (p *IPv4Pool) Release(ip IPv4) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.reserved, ip)
}

func NewIPv4PoolFromCIDR(s string) (*IPv4Pool, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}

	return NewIPv4Pool(ip, ipNet), nil
}

type IPv6Pool struct {
	ip       IPv6
	ipNet    *net.IPNet
	reserved map[IPv6]bool
	mutex    *sync.Mutex
}

func NewIPv6Pool(ip net.IP, ipNet *net.IPNet) *IPv6Pool {
	return &IPv6Pool{
		ipNet:    ipNet,
		ip:       NewIPv6FromIP(ip),
		reserved: make(map[IPv6]bool),
		mutex:    &sync.Mutex{},
	}
}

func (p *IPv6Pool) Get() (IPv6, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for ip := p.ip; p.ipNet.Contains(ip.IP()); {
		if _, ok := p.reserved[ip]; !ok {
			p.reserved[ip] = true
			return ip, nil
		}

		ip = ip.Next()
	}

	return IPv6{}, fmt.Errorf("ipv6 pool is full")
}

func (p *IPv6Pool) Release(ip IPv6) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.reserved, ip)
}

func NewIPv6PoolFromCIDR(s string) (*IPv6Pool, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}

	return NewIPv6Pool(ip, ipNet), nil
}

type IPPool struct {
	ipv4 *IPv4Pool
	ipv6 *IPv6Pool
}

func NewIPPool(v4 *IPv4Pool, v6 *IPv6Pool) *IPPool {
	return &IPPool{
		ipv4: v4,
		ipv6: v6,
	}
}

func (p *IPPool) Get() (IPv4, IPv6, error) {
	v4, err := p.ipv4.Get()
	if err != nil {
		return IPv4{}, IPv6{}, err
	}

	v6, err := p.ipv6.Get()
	if err != nil {
		return IPv4{}, IPv6{}, err
	}

	return v4, v6, nil
}

func (p *IPPool) Put(v4 IPv4, v6 IPv6) {
	p.ipv4.Release(v4)
	p.ipv6.Release(v6)
}
