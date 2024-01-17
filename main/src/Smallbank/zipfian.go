package Smallbank

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Zipfian struct {
	//Number

	//lock util.SpinLock

	items int64
	base  int64

	zipfianConstant float64

	alpha      float64
	zetan      float64
	theta      float64
	eta        float64
	zeta2Theta float64

	countForZeta int64

	allowItemCountDecrease bool
}

func NewZipfianWithItems(items int64, zipfianConstant float64) *Zipfian {
	return NewZipfianWithRange(0, items-1, zipfianConstant)
}

func NewZipfianWithRange(min int64, max int64, zipfianConstant float64) *Zipfian {
	return NewZipfian(min, max, zipfianConstant, zetaStatic(0, max-min+1, zipfianConstant, 0))
}

// NewZipfian creates the Zipfian generator.
func NewZipfian(min int64, max int64, zipfianConstant float64, zetan float64) *Zipfian {
	items := max - min + 1
	z := new(Zipfian)
	z.items = items
	z.base = min

	z.zipfianConstant = zipfianConstant
	theta := z.zipfianConstant
	z.theta = theta

	z.zeta2Theta = z.zeta(0, 2, theta, 0)

	z.alpha = 1.0 / (1.0 - theta)
	z.zetan = zetan
	z.countForZeta = items
	z.eta = (1 - math.Pow(2.0/float64(items), 1-theta)) / (1 - z.zeta2Theta/z.zetan)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	z.Next(r)
	return z
}

func (z *Zipfian) zeta(st int64, n int64, thetaVal float64, initialSum float64) float64 {
	z.countForZeta = n
	return zetaStatic(st, n, thetaVal, initialSum)
}

func zetaStatic(st int64, n int64, theta float64, initialSum float64) float64 {
	sum := initialSum

	for i := st; i < n; i++ {
		sum += 1 / math.Pow(float64(i+1), theta)
	}

	return sum
}

func (z *Zipfian) next(r *rand.Rand, itemCount int64) int64 {
	if itemCount != z.countForZeta {
		//z.lock.Lock()
		if itemCount > z.countForZeta {
			//we have added more items. can compute zetan incrementally, which is cheaper
			z.zetan = z.zeta(z.countForZeta, itemCount, z.theta, z.zetan)
			z.eta = (1 - math.Pow(2.0/float64(z.items), 1-z.theta)) / (1 - z.zeta2Theta/z.zetan)
		} else if itemCount < z.countForZeta && z.allowItemCountDecrease {
			//note : for large itemsets, this is very slow. so don't do it!
			fmt.Printf("recomputing Zipfian distribution, should be avoided,item count %v, count for zeta %v\n", itemCount, z.countForZeta)
			z.zetan = z.zeta(0, itemCount, z.theta, 0)
			z.eta = (1 - math.Pow(2.0/float64(z.items), 1-z.theta)) / (1 - z.zeta2Theta/z.zetan)
		}
		//z.lock.Unlock()
	}

	u := r.Float64()
	uz := u * z.zetan

	if uz < 1.0 {
		return z.base
	}

	if uz < 1.0+math.Pow(0.5, z.theta) {
		return z.base + 1
	}

	ret := z.base + int64(float64(itemCount)*math.Pow(z.eta*u-z.eta+1, z.alpha))
	//z.SetLastValue(ret)
	return ret
}

// Next implements the Generator Next interface.
func (z *Zipfian) Next(r *rand.Rand) int64 {
	return z.next(r, z.items)
}
