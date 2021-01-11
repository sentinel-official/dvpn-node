package types

import (
	"sync"
)

type Peer struct {
	Identity string
	IPv4     IPv4
	IPv6     IPv6
}

type Peers struct {
	m     map[string]Peer
	mutex *sync.RWMutex
}

func NewPeers() *Peers {
	return &Peers{
		m:     make(map[string]Peer),
		mutex: &sync.RWMutex{},
	}
}

func (p *Peers) Get(key string) (Peer, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	peer, ok := p.m[key]
	return peer, ok
}

func (p *Peers) Put(item Peer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.m[item.Identity]; !ok {
		p.m[item.Identity] = item
	}
}

func (p *Peers) Delete(key string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.m, key)
}

func (p *Peers) Len() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return len(p.m)
}
