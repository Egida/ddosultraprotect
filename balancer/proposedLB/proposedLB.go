package proposedLB

import (
	"gonum.org/v1/gonum/optimize"
	"gonum.org/v1/gonum/stat"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
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

//var cg = optimize.CG{
//	Variant:                &optimize.HestenesStiefel{},
//	Linesearcher:           &optimize.MoreThuente{},
//	IterationRestartFactor: 1,
//	GradStopThreshold:      0,
//}

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

func objective(x []float64) float64 {
	// Compute the loss function
	x1 := x0(len(x))

	loss := stat.ChiSquare(x1, x)

	// Compute the L2 regularization term
	reg := 0.0
	for i := 0; i < len(x); i++ {
		reg += x[i] * x[i]
	}
	reg *= rand.Float64() / 2.0

	// Return the sum of the loss function and the regularization term
	return loss + reg
}

func x0(l int) []float64 {
	var arrN = make([]float64, 0, 0)
	for i := 0; i < l; i++ {
		arrN = append(arrN, rand.Float64())
	}
	return arrN
}

func (*newLBPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	// specify method used to choose the load balancer here
	// lines below are taken from rr example
	logger.Infof("proposedLB: Build called with info: %v", info)
	allSCs := rand.Intn(len(info.ReadySCs))

	if allSCs == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make([]balancer.SubConn, allSCs)

	result, _ := optimize.Minimize(optimize.Problem{Func: objective}, x0(allSCs), nil, &optimize.CG{})

	return &newlbPicker{
		subConns: scs,
		next:     uint32(result.F),

		//cg.NextDirection(&optimize.Location{}, make([]float64, float64(rand.Intn(allSCs))))),
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
