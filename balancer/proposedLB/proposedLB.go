package proposedLB

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/benchmark/stats"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"math/rand"
	"sync/atomic"
	"time"
)

// Name taken from rr example
const Name = "proposedLB"

const smallestCPUsize = 8
const totalCPUsize = 32
const capacity = float64(60 / 100)

const growthFactor = float64(12 / 100)

// initialize variables before calling functions which will change the values
var numItemsInBucket = 0

var numBalancersNeeded = 0

// taken from rr example
var logger = grpclog.Component("proposedLB")

// taken from rr example
func newProposedBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &newLBPickerBuilder{extraParams: keepalive.ServerParameters{
		MaxConnectionAgeGrace: time.Duration(5 / growthFactor),
	}}, base.Config{HealthCheck: true})
}

// taken from rr example
func init() {
	balancer.Register(newProposedBuilder())
}

type newLBPickerBuilder struct {
	extraParams keepalive.ServerParameters
}

func (*newLBPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	// specify method used to choose the load balancer here
	// lines below are taken from rr example
	logger.Infof("proposedLB: Build called with info: %v", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	// get sub connections

	// get count of all sub connections and turn it into a histogram
	intervals := stats.Histogram{
		Count: int64(len(info.ReadySCs)),
	}

	for k := 1; k <= len(info.ReadySCs); k++ {
		// for each bucket in the histogram, obtain the number of items needed
		numItemsInBucket = int(intervals.Buckets[k].Count)

		// make it divisible and not infinite.. original RAM capacity must be divisible by the newer and smaller RAM capacity
		if capacity > 0 && totalCPUsize > 0 && totalCPUsize%smallestCPUsize == 0 {
			// most important line of code
			if numItemsInBucket > int(capacity*float64(len(info.ReadySCs))) {
				numBalancersNeeded = numItemsInBucket * smallestCPUsize / int(growthFactor*capacity*float64(totalCPUsize))
			} else {
				numBalancersNeeded = numItemsInBucket * (totalCPUsize / smallestCPUsize)
			}

		} else {
			numBalancersNeeded = 0
		}

		if len(info.ReadySCs) == 0 {
			return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
		}
	}
	var scs = make([]balancer.SubConn, 0, numBalancersNeeded)

	return &newlbPicker{
		subConns: scs,
		next:     uint32(rand.Intn(numBalancersNeeded)),
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

func (n newlbPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	//Taken from rr algorithm
	subConnsLen := uint32(len(n.subConns))
	nextIndex := atomic.AddUint32(&n.next, 1)

	sc := n.subConns[nextIndex%subConnsLen]
	return balancer.PickResult{SubConn: sc}, nil
}
