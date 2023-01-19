package proposedLB

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/benchmark/stats"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"math/rand"
	"sync/atomic"
)

// Name taken from rr example
const Name = "proposedLB"



// initialize variables before calling functions which will change the values

var numConnections = 0

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
	 extraParams keepalive.ServerParameters
	 extraParams2 keepalive.EnforcementPolicy
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
		
	bucket_count := make([]int, intervals.Buckets.Count)	
	
	for k:= 0; k <= intervals.Buckets.Count; k++{
		bucket_count = append(bucket_count, intervals.Buckets[k].Count)
	}
	
	
	return &newlbPicker{
		subConns: scs,
		next:     uint32(rand.Intn(len(scs))),
		bucketCount: bucket_count,
	}
}

type newlbPicker struct {
	// subConns is the snapshot of the new load balancer when this picker was
	// created. The slice is immutable. Each Get() will
	// select from it and return the selected SubConn.
	// use grpc methods and avoid using third party libraries
	subConns []balancer.SubConn
	
	next uint32
	
	bucketCount []int
}

func (n newlbPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	//Taken from rr algorithm
	
	for k:= 0; k < 32; k++{
		subConnsLen := uint32(len(info.ReadySCs))
		nextIndex := atomic.AddUint32(&n.next, &n.bucketCount[k])
		sc := n.subConns[nextIndex%subConnsLen]
		return balancer.PickResult{SubConn: sc}, nil
	}
	
	return "done"

}
