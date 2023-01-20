package proposedLB

import (
	"gonum.org/v1/gonum/optimize"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"math/rand"
	"sync/atomic"
	"time"
)

// Name taken from rr example
const Name = "proposedLB"

// initialize variables before calling functions which will change the values

var growthFactor = 0.6

// taken from rr example
var logger = grpclog.Component("proposedLB")

// taken from rr example
func newProposedBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &newLBPickerBuilder{extraParams: keepalive.ServerParameters{
		MaxConnectionAgeGrace: time.Duration(10 / growthFactor)},
		extraParams2: keepalive.EnforcementPolicy{
			MinTime: time.Duration(1),
		},
	}, base.Config{HealthCheck: true})
}

// taken from rr example
func init() {
	balancer.Register(newProposedBuilder())
}

type newLBPickerBuilder struct {
	extraParams  keepalive.ServerParameters
	extraParams2 keepalive.EnforcementPolicy
}

func (*newLBPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	// specify method used to choose the load balancer here
	// lines below are taken from rr example
	logger.Infof("proposedLB: Build called with info: %v", info)

	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))

	arrNums := make([]float64, rand.Intn(len(info.ReadySCs)), len(info.ReadySCs))

	cg := optimize.CG{
		Variant:      &optimize.HestenesStiefel{},
		Linesearcher: &optimize.MoreThuente{},
	}
	//cg.Init(1, len(info.ReadySCs))

	return &newlbPicker{
		subConns: scs,
		next:     uint32(cg.NextDirection(&optimize.Location{}, arrNums)),
	}
}

type newlbPicker struct {
	// subConns is the snapshot of the new load balancer when this picker was
	// created. The slice is immutable. Each Get() will
	// select from it and return the selected SubConn.
	// use grpc methods and avoid using third party libraries
	subConns []balancer.SubConn

	next uint32
}

func (n *newlbPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	//Taken from rr algorithm
	subConnsLen := uint32(len(n.subConns))

	nextIndex := atomic.AddUint32(&n.next, 1)

	sc := n.subConns[nextIndex%subConnsLen]
	return balancer.PickResult{SubConn: sc}, nil
}
