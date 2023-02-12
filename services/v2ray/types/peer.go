package types

import (
	"sync"
)

type Peer struct {
	Identity string
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

func (p *Peers) Iterate(f func(key string, value Peer) (bool, error)) error {
	p.RLock()
	defer p.RUnlock()

	for key, value := range p.m {
		stop, err := f(key, value)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	return nil
}
