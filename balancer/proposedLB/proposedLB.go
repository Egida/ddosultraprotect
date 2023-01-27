package proposedLB

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/benchmark/stats"
	"google.golang.org/grpc/grpclog"
	"math/rand"
	"sync/atomic"
)

// Name taken from rr example
const Name = "proposedLB"

// initialize variables before calling functions which will change the values

//var growthFactor = 8

// taken from rr example
var logger = grpclog.Component("proposedLB")

// taken from rr example
func newProposedBuilder() balancer.Builder {

	// gcp keepalive params
	return base.NewBalancerBuilder(Name, &newLBPickerBuilder{}, base.Config{HealthCheck: true})
}

// taken from rr example
func init() {
	balancer.Register(newProposedBuilder())
}

type newLBPickerBuilder struct {
}

func (*newLBPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	// specify method used to choose the load balancer here
	// lines below are taken from rr example
	logger.Infof("proposedLB: Build called with info: %v", info)
	allSCs := len(info.ReadySCs)
	const bucketsCount = 32
	if allSCs == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make([]balancer.SubConn, allSCs)

	return &newlbPicker{
		subConns: scs,
		next: int64(stats.Histogram{
			Count: int64(rand.Intn(allSCs)),
		}.Buckets[rand.Intn(bucketsCount)].LowBound),
	}
}

type newlbPicker struct {
	// subConns is the snapshot of the new load balancer when this picker was
	// created. The slice is immutable. Each Get() will
	// select from it and return the selected SubConn.
	// use grpc methods and avoid using third party libraries
	subConns []balancer.SubConn
	next     int64
}

func (n *newlbPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	//Taken from rr algorithm
	subConnsLen := int64(len(n.subConns))

	nextIndex := atomic.AddInt64(&n.next, 1)

	sc := n.subConns[nextIndex%subConnsLen]
	return balancer.PickResult{SubConn: sc}, nil
}
