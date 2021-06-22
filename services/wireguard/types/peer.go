package types

import (
	"sync"
)

type Peer struct {
	Identity string
	IPv4     IPv4
	IPv6     IPv6
}

func (p Peer) Empty() bool {
	return p.Identity == ""
}

type Peers struct {
	sync.RWMutex
	m map[string]Peer
}

func NewPeers() *Peers {
	return &Peers{
		m: make(map[string]Peer),
	}
}

func (p *Peers) Get(key string) Peer {
	p.RLock()
	defer p.RUnlock()

	v, ok := p.m[key]
	if !ok {
		return Peer{}
	}

	return v
}

func (p *Peers) Put(v Peer) {
	p.Lock()
	defer p.Unlock()

	_, ok := p.m[v.Identity]
	if ok {
		return
	}

	p.m[v.Identity] = v
}

func (p *Peers) Delete(v string) {
	p.Lock()
	defer p.Unlock()

	delete(p.m, v)
}

func (p *Peers) Len() int {
	p.RLock()
	defer p.RUnlock()

	return len(p.m)
}
