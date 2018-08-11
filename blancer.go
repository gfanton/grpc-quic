package grpcquic

import (
	"context"
	"math/rand"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
)

// Balancer implement blancer.Balancer
var _ blancer.Balancer = (*Balancer)(nil)

type Balancer struct {
	cc          balancer.ClientConn
	serviceName string
}

// HandleSubConnStateChange is called by gRPC when the connectivity state
// of sc has changed.
// Balancer is expected to aggregate all the state of SubConn and report
// that back to gRPC.
// Balancer should also generate and update Pickers when its internal state has
// been changed by the new state.
func (b *Balancer) HandleSubConnStateChange(sc SubConn, state connectivity.State) {
}

// HandleResolvedAddrs is called by gRPC to send updated resolved addresses to
// balancers.
// Balancer can create new SubConn or remove SubConn with the addresses.
// An empty address slice and a non-nil error will be passed if the resolver returns
// non-nil error to gRPC.

func (b *Balancer) HandleResolvedAddrs([]resolver.Address, error) {

}

func (b *Balancer) Close() {}

type balancerBuilder struct {
	name string
}

// NewBalancerBuilder returns a balancer builder. The balancers
// built by this builder will use the picker builder to build pickers.
func NewBalancerBuilder(name string) balancer.Builder {
	return &balancerBuilder{
		name: name,
	}
}

func (bb *balancerBuilder) Build(cc balancer.ClientConn, opt balancer.BuildOptions) balancer.Balancer {
	return &Balancer{
		cc:          cc,
		serviceName: bb.name,
	}
}

func (bb *balancerBuilder) Name() string {
	return bb.name
}

type (
	randomBuilder struct{}
	randomPicker  struct {
		conns []balancer.SubConn
		seed  int64
	}
)

// OLD

const (
	// Name is the balancer name.
	Name = "random"
)

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &randomBuilder{})
}

func (*randomBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	var conns []balancer.SubConn
	for _, conn := range readySCs {
		conns = append(conns, conn)
	}
	return &randomPicker{
		conns: conns,
		seed:  rand.Int63(),
	}
}

func (p *randomPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	if len(p.conns) == 0 {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return p.conns[rand.Intn(len(p.conns))], nil, nil
}

func init() {
	balancer.Register(newBuilder())
}
