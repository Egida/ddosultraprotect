package proposedLB

import (
	"gonum.org/v1/gonum/mat"
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

var growthFactor = 8

// taken from rr example
var logger = grpclog.Component("proposedLB")

// taken from rr example
func newProposedBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &newLBPickerBuilder{extraParams: keepalive.ServerParameters{
		MaxConnectionAgeGrace: time.Duration(2),
		MaxConnectionAge:      time.Duration(2 / growthFactor),
		Time:                  time.Duration(1000000), // 1 second
	},
		extraParams2: keepalive.EnforcementPolicy{
			MinTime:             time.Duration(1 / growthFactor),
			PermitWithoutStream: false,
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
	randomNum := rand.Intn(len(info.ReadySCs))
	diagMat := mat.NewDiagonalRect(randomNum, randomNum, make([]float64, 0, float64(randomNum)))

	var lu mat.LU

	lu.Factorize(diagMat)

	det, sign := lu.LogDet()

	if sign < 0 {
		logger.Infof("negative sign for determinant")
	}

	return &newlbPicker{
		subConns: scs,
		next:     uint32(det),
	}
}

type newlbPicker struct {
	// subConns is the snapshot of the new load balancer when this picker was
	// created. The slice is immutable. Each Get() will
	// select from it and return the selected SubConn.
	// use grpc methods and avoid using third party libraries
	subConns []balancer.SubConn
	next     uint32
}

func (n *newlbPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	//Taken from rr algorithm
	subConnsLen := uint32(len(n.subConns))

	nextIndex := atomic.AddUint32(&n.next, 1)

	sc := n.subConns[nextIndex%subConnsLen]
	return balancer.PickResult{SubConn: sc}, nil
}
