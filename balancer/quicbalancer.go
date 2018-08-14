package quicbalancer

import (
	"sync"

	qnet "github.com/gfanton/grpc-quic/net"
	ma "github.com/multiformats/go-multiaddr"

	"golang.org/x/net/context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

// Name is the name of round_robin balancer.
const Name = "quic_balancer"

// newBuilder creates a new roundrobin balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &rrPickerBuilder{})
}

func init() {
	balancer.Register(newBuilder())
}

type rrPickerBuilder struct{}

func (*rrPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	var scsTCP, scsUDP []balancer.SubConn

	for a, sc := range readySCs {
		m, err := ma.NewMultiaddr(a.Addr)
		if err != nil {
			// @TODO: LOG THIS
			continue
		}

		_, protocol, err := qnet.ParseMultiaddr(m)
		if err != nil {
			// @TODO: LOG THIS
			continue
		}

		switch protocol {
		case ma.P_UDP:
			scsUDP = append(scsUDP, sc)
		case ma.P_TCP:
			scsTCP = append(scsTCP, sc)
		default: // @TODO: LOG THIS
		}
	}

	return &rrPicker{
		subConnsUDP: scsUDP,
		subConnsTCP: scsTCP,
	}
}

type rrPicker struct {
	// subConnsTCP and subConnsUDP are the snapshot of the roundrobin balancer when this picker was
	// created. The slice is immutable. Each Get() will do a round robin
	// selection from it and return the selected SubConn.
	subConnsTCP []balancer.SubConn
	subConnsUDP []balancer.SubConn

	mu   sync.Mutex
	next int
}

func (p *rrPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	// Chain TCP subConn after UDP subconn
	scs := append(p.subConnsUDP, p.subConnsTCP...)
	if len(scs) <= 0 {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}

	p.mu.Lock()
	sc := scs[p.next]
	p.next = (p.next + 1) % len(scs)
	p.mu.Unlock()
	return sc, nil, nil
}
