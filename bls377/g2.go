// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gurvy DO NOT EDIT

package bls377

import (
	"math/big"

	"github.com/consensys/gurvy/bls377/fr"
	"github.com/consensys/gurvy/utils/debug"
)

// G2Jac is a point with E2 coordinates
type G2Jac struct {
	X, Y, Z E2
}

// G2Proj point in projective coordinates
type G2Proj struct {
	X, Y, Z E2
}

// G2Affine point in affine coordinates
type G2Affine struct {
	X, Y E2
}

//  g2JacExtended parameterized jacobian coordinates (x=X/ZZ, y=Y/ZZZ, ZZ**3=ZZZ**2)
type g2JacExtended struct {
	X, Y, ZZ, ZZZ E2
}

// SetInfinity sets p to O
func (p *g2JacExtended) SetInfinity() *g2JacExtended {
	p.X.SetOne()
	p.Y.SetOne()
	p.ZZ.SetZero()
	p.ZZZ.SetZero()
	return p
}

// ToAffine sets p in affine coords
func (p *g2JacExtended) ToAffine(Q *G2Affine) *G2Affine {
	var zero E2
	if p.ZZ.Equal(&zero) {
		Q.X.Set(&zero)
		Q.Y.Set(&zero)
		return Q
	}
	Q.X.Inverse(&p.ZZ).Mul(&Q.X, &p.X)
	Q.Y.Inverse(&p.ZZZ).Mul(&Q.Y, &p.Y)
	return Q
}

// ToJac sets p in affine coords
func (p *g2JacExtended) ToJac(Q *G2Jac) *G2Jac {
	var zero E2
	if p.ZZ.Equal(&zero) {
		Q.Set(&g2Infinity)
		return Q
	}
	Q.X.Mul(&p.ZZ, &p.X).Mul(&Q.X, &p.ZZ)
	Q.Y.Mul(&p.ZZZ, &p.Y).Mul(&Q.Y, &p.ZZZ)
	Q.Z.Set(&p.ZZZ)
	return Q
}

// unsafeToJac sets p in affine coords, but don't check for infinity
func (p *g2JacExtended) unsafeToJac(Q *G2Jac) *G2Jac {
	Q.X.Mul(&p.ZZ, &p.X).Mul(&Q.X, &p.ZZ)
	Q.Y.Mul(&p.ZZZ, &p.Y).Mul(&Q.Y, &p.ZZZ)
	Q.Z.Set(&p.ZZZ)
	return Q
}

// mAdd
// http://www.hyperelliptic.org/EFD/ g2p/auto-shortw-xyzz.html#addition-madd-2008-s
func (p *g2JacExtended) mAdd(a *G2Affine) *g2JacExtended {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.ZZ.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.ZZ.SetOne()
		p.ZZZ.SetOne()
		return p
	}

	var U2, S2, P, R, PP, PPP, Q, Q2, RR, X3, Y3 E2

	// p2: a, p1: p
	U2.Mul(&a.X, &p.ZZ)
	S2.Mul(&a.Y, &p.ZZZ)
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.double(a)
	}
	P.Sub(&U2, &p.X)
	R.Sub(&S2, &p.Y)
	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&p.X, &PP)
	RR.Square(&R)
	X3.Sub(&RR, &PPP)
	Q2.Double(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).Mul(&Y3, &R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.Mul(&p.ZZ, &PP)
	p.ZZZ.Mul(&p.ZZZ, &PPP)

	return p
}

// double point in ZZ coords
// http://www.hyperelliptic.org/EFD/ g2p/auto-shortw-xyzz.html#doubling-dbl-2008-s-1
func (p *g2JacExtended) double(q *G2Affine) *g2JacExtended {

	var U, S, M, _M, Y3 E2

	U.Double(&q.Y)
	p.ZZ.Square(&U)
	p.ZZZ.Mul(&U, &p.ZZ)
	S.Mul(&q.X, &p.ZZ)
	_M.Square(&q.X)
	M.Double(&_M).
		Add(&M, &_M) // -> + a, but a=0 here
	p.X.Square(&M).
		Sub(&p.X, &S).
		Sub(&p.X, &S)
	Y3.Sub(&S, &p.X).Mul(&Y3, &M)
	U.Mul(&p.ZZZ, &q.Y)
	p.Y.Sub(&Y3, &U)

	return p
}

// Set set p to the provided point
func (p *G2Jac) Set(a *G2Jac) *G2Jac {
	p.X.Set(&a.X)
	p.Y.Set(&a.Y)
	p.Z.Set(&a.Z)
	return p
}

// Equal tests if two points (in Jacobian coordinates) are equal
func (p *G2Jac) Equal(a *G2Jac) bool {

	if p.Z.IsZero() && a.Z.IsZero() {
		return true
	}
	_p := G2Affine{}
	_p.FromJacobian(p)

	_a := G2Affine{}
	_a.FromJacobian(a)

	return _p.X.Equal(&_a.X) && _p.Y.Equal(&_a.Y)
}

// Equal tests if two points (in Affine coordinates) are equal
func (p *G2Affine) Equal(a *G2Affine) bool {
	return p.X.Equal(&a.X) && p.Y.Equal(&a.Y)
}

// Clone returns a copy of self
func (p *G2Jac) Clone() *G2Jac {
	return &G2Jac{
		p.X, p.Y, p.Z,
	}
}

// Neg computes -G
func (p *G2Jac) Neg(a *G2Jac) *G2Jac {
	p.Set(a)
	p.Y.Neg(&a.Y)
	return p
}

// Neg computes -G
func (p *G2Affine) Neg(a *G2Affine) *G2Affine {
	p.X.Set(&a.X)
	p.Y.Neg(&a.Y)
	return p
}

// SubAssign substracts two points on the curve
func (p *G2Jac) SubAssign(a G2Jac) *G2Jac {
	a.Y.Neg(&a.Y)
	p.AddAssign(&a)
	return p
}

// FromJacobian rescale a point in Jacobian coord in z=1 plane
func (p *G2Affine) FromJacobian(p1 *G2Jac) *G2Affine {

	var a, b E2

	if p1.Z.IsZero() {
		p.X.SetZero()
		p.Y.SetZero()
		return p
	}

	a.Inverse(&p1.Z)
	b.Square(&a)
	p.X.Mul(&p1.X, &b)
	p.Y.Mul(&p1.Y, &b).Mul(&p.Y, &a)

	return p
}

// FromJacobian converts a point from Jacobian to projective coordinates
func (p *G2Proj) FromJacobian(Q *G2Jac) *G2Proj {
	// memalloc
	var buf E2
	buf.Square(&Q.Z)

	p.X.Mul(&Q.X, &Q.Z)
	p.Y.Set(&Q.Y)
	p.Z.Mul(&Q.Z, &buf)

	return p
}

func (p *G2Jac) String() string {
	if p.Z.IsZero() {
		return "O"
	}
	_p := G2Affine{}
	_p.FromJacobian(p)
	return "E([" + _p.X.String() + "," + _p.Y.String() + "]),"
}

// FromAffine sets p = Q, p in Jacboian, Q in affine
func (p *G2Jac) FromAffine(Q *G2Affine) *G2Jac {
	if Q.X.IsZero() && Q.Y.IsZero() {
		p.Z.SetZero()
		p.X.SetOne()
		p.Y.SetOne()
		return p
	}
	p.Z.SetOne()
	p.X.Set(&Q.X)
	p.Y.Set(&Q.Y)
	return p
}

func (p *G2Affine) String() string {
	var x, y E2
	x.Set(&p.X)
	y.Set(&p.Y)
	return "E([" + x.String() + "," + y.String() + "]),"
}

// IsInfinity checks if the point is infinity (in affine, it's encoded as (0,0))
func (p *G2Affine) IsInfinity() bool {
	return p.X.IsZero() && p.Y.IsZero()
}

// AddAssign point addition in montgomery form
// https://hyperelliptic.org/EFD/g2p/auto-shortw-jacobian-3.html#addition-add-2007-bl
func (p *G2Jac) AddAssign(a *G2Jac) *G2Jac {

	// p is infinity, return a
	if p.Z.IsZero() {
		p.Set(a)
		return p
	}

	// a is infinity, return p
	if a.Z.IsZero() {
		return p
	}

	var Z1Z1, Z2Z2, U1, U2, S1, S2, H, I, J, r, V E2
	Z1Z1.Square(&a.Z)
	Z2Z2.Square(&p.Z)
	U1.Mul(&a.X, &Z2Z2)
	U2.Mul(&p.X, &Z1Z1)
	S1.Mul(&a.Y, &p.Z).
		Mul(&S1, &Z2Z2)
	S2.Mul(&p.Y, &a.Z).
		Mul(&S2, &Z1Z1)

	// if p == a, we double instead
	if U1.Equal(&U2) && S1.Equal(&S2) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &U1)
	I.Double(&H).
		Square(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &S1).Double(&r)
	V.Mul(&U1, &I)
	p.X.Square(&r).
		Sub(&p.X, &J).
		Sub(&p.X, &V).
		Sub(&p.X, &V)
	p.Y.Sub(&V, &p.X).
		Mul(&p.Y, &r)
	S1.Mul(&S1, &J).Double(&S1)
	p.Y.Sub(&p.Y, &S1)
	p.Z.Add(&p.Z, &a.Z)
	p.Z.Square(&p.Z).
		Sub(&p.Z, &Z1Z1).
		Sub(&p.Z, &Z2Z2).
		Mul(&p.Z, &H)

	return p
}

// AddMixed point addition
// http://www.hyperelliptic.org/EFD/g2p/auto-shortw-jacobian-0.html#addition-madd-2007-bl
func (p *G2Jac) AddMixed(a *G2Affine) *G2Jac {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.Z.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.Z.SetOne()
		return p
	}

	// get some Element from our pool
	var Z1Z1, U2, S2, H, HH, I, J, r, V E2
	Z1Z1.Square(&p.Z)
	U2.Mul(&a.X, &Z1Z1)
	S2.Mul(&a.Y, &p.Z).
		Mul(&S2, &Z1Z1)

	// if p == a, we double instead
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &p.X)
	HH.Square(&H)
	I.Double(&HH).Double(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &p.Y).Double(&r)
	V.Mul(&p.X, &I)
	p.X.Square(&r).
		Sub(&p.X, &J).
		Sub(&p.X, &V).
		Sub(&p.X, &V)
	J.Mul(&J, &p.Y).Double(&J)
	p.Y.Sub(&V, &p.X).
		Mul(&p.Y, &r)
	p.Y.Sub(&p.Y, &J)
	p.Z.Add(&p.Z, &H)
	p.Z.Square(&p.Z).
		Sub(&p.Z, &Z1Z1).
		Sub(&p.Z, &HH)

	return p
}

// Double doubles a point in Jacobian coordinates
// https://hyperelliptic.org/EFD/g2p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *G2Jac) Double(q *G2Jac) *G2Jac {
	p.Set(q)
	p.DoubleAssign()
	return p
}

// DoubleAssign doubles a point in Jacobian coordinates
// https://hyperelliptic.org/EFD/g2p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *G2Jac) DoubleAssign() *G2Jac {

	// get some Element from our pool
	var XX, YY, YYYY, ZZ, S, M, T E2

	XX.Square(&p.X)
	YY.Square(&p.Y)
	YYYY.Square(&YY)
	ZZ.Square(&p.Z)
	S.Add(&p.X, &YY)
	S.Square(&S).
		Sub(&S, &XX).
		Sub(&S, &YYYY).
		Double(&S)
	M.Double(&XX).Add(&M, &XX)
	p.Z.Add(&p.Z, &p.Y).
		Square(&p.Z).
		Sub(&p.Z, &YY).
		Sub(&p.Z, &ZZ)
	T.Square(&M)
	p.X = T
	T.Double(&S)
	p.X.Sub(&p.X, &T)
	p.Y.Sub(&S, &p.X).
		Mul(&p.Y, &M)
	YYYY.Double(&YYYY).Double(&YYYY).Double(&YYYY)
	p.Y.Sub(&p.Y, &YYYY)

	return p
}

// ScalarMulByGen multiplies given scalar by generator
func (p *G2Jac) ScalarMulByGen(s *big.Int) *G2Jac {
	return p.ScalarMulGLV(&g2GenAff, s)
}

// ScalarMultiplication algo for exponentiation
func (p *G2Jac) ScalarMultiplication(a *G2Affine, s *big.Int) *G2Jac {

	var res G2Jac
	res.Set(&g2Infinity)
	b := s.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0x80)
		for j := 0; j < 8; j++ {
			res.DoubleAssign()
			if (w&mask)>>(7-j) != 0 {
				res.AddMixed(a)
			}
			mask = mask >> 1
		}
	}
	p.Set(&res)

	return p
}

// ScalarMulGLV performs scalar multiplication using GLV (without the lattice reduction)
func (p *G2Jac) ScalarMulGLV(a *G2Affine, s *big.Int) *G2Jac {

	var g2, phig2, res G2Jac
	var phig2Affine G2Affine
	res.Set(&g2Infinity)
	g2.FromAffine(a)
	phig2.Set(&g2)
	phig2.X.MulByElement(&phig2.X, &thirdRootOneG2)

	phig2Affine.FromJacobian(&phig2)

	// s = s1*lambda+s2
	var s1, s2 big.Int
	s1.DivMod(s, &lambdaGLV, &s2)

	// s1 part (on phi(g2)=lambda*g2)
	phig2.ScalarMultiplication(&phig2Affine, &s1)

	// s2 part (on g2)
	g2.ScalarMultiplication(a, &s2)

	res.AddAssign(&phig2)
	res.AddAssign(&g2)

	p.Set(&res)

	return p
}

// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf
func (p *G2Jac) MultiExp(points []G2Affine, scalars []fr.Element) chan G2Jac {
	// note:
	// each of the multiExpcX method is the same, except for the c constant it declares
	// duplicating (through template generation) these methods allows to declare the buckets on the stack
	// the choice of c needs to be improved:
	// there is a theoritical value that gives optimal asymptotics
	// but in practice, other factors come into play, including:
	// * if c doesn't divide 64, the word size, then we're bound to select bits over 2 words of our scalars, instead of 1
	// * number of CPUs
	// * cache friendliness (which depends on the host, G1 or G2... )
	//	--> for example, on BN256, a G1 point fits into one cache line of 64bytes, but a G2 point don't.

	nbPoints := len(points)
	if nbPoints <= (1 << 5) {
		return p.multiExpc4(points, scalars)
	} else if nbPoints <= 200000 {
		return p.multiExpc8(points, scalars)
	} else {
		return p.multiExpc16(points, scalars)
	}
}

func (p *G2Jac) multiExpc4(points []G2Affine, scalars []fr.Element) chan G2Jac {

	const c = 4                              // scalars partitioned into c-bit radixes
	const t = fr.Bits / c                    // number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1 // low c bits are 1
	const nbChunks = t + 1                   // note: if c doesn't divide fr.Bits, nbChunks != t)

	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan G2Jac
	for i := 0; i < nbChunks; i++ {
		chTotals[i] = make(chan G2Jac, 1)
	}

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [(1 << c) - 1]g2JacExtended
			bucketAccumulateG2(chunk, c, selectorMask, points, scalars, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduceG2(p, c, chTotals[:])

}

func (p *G2Jac) multiExpc8(points []G2Affine, scalars []fr.Element) chan G2Jac {

	const c = 8                              // scalars partitioned into c-bit radixes
	const t = fr.Bits / c                    // number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1 // low c bits are 1
	const nbChunks = t + 1                   // note: if c doesn't divide fr.Bits, nbChunks != t)

	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan G2Jac
	for i := 0; i < nbChunks; i++ {
		chTotals[i] = make(chan G2Jac, 1)
	}

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [(1 << c) - 1]g2JacExtended
			bucketAccumulateG2(chunk, c, selectorMask, points, scalars, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduceG2(p, c, chTotals[:])

}

func (p *G2Jac) multiExpc10(points []G2Affine, scalars []fr.Element) chan G2Jac {

	const c = 10                             // scalars partitioned into c-bit radixes
	const t = fr.Bits / c                    // number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1 // low c bits are 1
	const nbChunks = t + 1                   // note: if c doesn't divide fr.Bits, nbChunks != t)

	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan G2Jac
	for i := 0; i < nbChunks; i++ {
		chTotals[i] = make(chan G2Jac, 1)
	}

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [(1 << c) - 1]g2JacExtended
			bucketAccumulateG2(chunk, c, selectorMask, points, scalars, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduceG2(p, c, chTotals[:])

}

func (p *G2Jac) multiExpc14(points []G2Affine, scalars []fr.Element) chan G2Jac {

	const c = 14                             // scalars partitioned into c-bit radixes
	const t = fr.Bits / c                    // number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1 // low c bits are 1
	const nbChunks = t + 1                   // note: if c doesn't divide fr.Bits, nbChunks != t)

	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan G2Jac
	for i := 0; i < nbChunks; i++ {
		chTotals[i] = make(chan G2Jac, 1)
	}

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [(1 << c) - 1]g2JacExtended
			bucketAccumulateG2(chunk, c, selectorMask, points, scalars, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduceG2(p, c, chTotals[:])

}

func (p *G2Jac) multiExpc16(points []G2Affine, scalars []fr.Element) chan G2Jac {

	const c = 16                             // scalars partitioned into c-bit radixes
	const t = fr.Bits / c                    // number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1 // low c bits are 1
	const nbChunks = t + 1                   // note: if c doesn't divide fr.Bits, nbChunks != t)

	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan G2Jac
	for i := 0; i < nbChunks; i++ {
		chTotals[i] = make(chan G2Jac, 1)
	}

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [(1 << c) - 1]g2JacExtended
			bucketAccumulateG2(chunk, c, selectorMask, points, scalars, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduceG2(p, c, chTotals[:])

}

func (p *G2Jac) multiExpc18(points []G2Affine, scalars []fr.Element) chan G2Jac {

	const c = 18                             // scalars partitioned into c-bit radixes
	const t = fr.Bits / c                    // number of c-bit radixes in a scalar
	const selectorMask uint64 = (1 << c) - 1 // low c bits are 1
	const nbChunks = t + 1                   // note: if c doesn't divide fr.Bits, nbChunks != t)

	// 1 channel per chunk, which will contain the weighted sum of the its buckets
	var chTotals [nbChunks]chan G2Jac
	for i := 0; i < nbChunks; i++ {
		chTotals[i] = make(chan G2Jac, 1)
	}

	// for each chunk, add points to the buckets, then do the weighted sum of the buckets
	// TODO we don't take into account the number of available CPUs here, and we should. WIP on parralelism strategy.
	for j := nbChunks - 1; j >= 0; j-- {
		go func(chunk int) {
			var buckets [(1 << c) - 1]g2JacExtended
			bucketAccumulateG2(chunk, c, selectorMask, points, scalars, buckets[:], chTotals[chunk])
		}(j)
	}

	return chunkReduceG2(p, c, chTotals[:])

}

// bucketAccumulate places points into buckets base on their selector and return the weighted bucket sum in given channel
func bucketAccumulateG2(chunk, c int, selectorMask uint64, points []G2Affine, scalars []fr.Element, buckets []g2JacExtended, chRes chan<- G2Jac) {

	for i := 0; i < len(buckets); i++ {
		buckets[i].SetInfinity()
	}

	// place points into buckets based on their selector
	jc := uint64(chunk * c)
	selectorIndex := jc / 64
	selectorShift := jc - (selectorIndex * 64)
	selectedBits := selectorMask << selectorShift

	multiWordSelect := int(selectorShift) > (64-c) && selectorIndex < (fr.Limbs-1)

	if !multiWordSelect {
		for i := 0; i < len(scalars); i++ {
			selector := (scalars[i][selectorIndex] & selectedBits) >> selectorShift
			if selector == 0 {
				continue
			}
			buckets[selector-1].mAdd(&points[i])
		}
	} else {
		// we are selecting bits over 2 words
		selectorIndexNext := selectorIndex + 1
		nbBitsHigh := selectorShift - uint64(64-c)
		highShift := 64 - nbBitsHigh
		highShiftRight := highShift - (64 - selectorShift)

		for i := 0; i < len(scalars); i++ {
			selector := (scalars[i][selectorIndex] & selectedBits) >> selectorShift
			selectorNext := (scalars[i][selectorIndexNext] << highShift) >> highShiftRight
			selector |= selectorNext
			if selector == 0 {
				continue
			}
			buckets[selector-1].mAdd(&points[i])
		}
	}

	// reduce buckets into total
	// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]

	var runningSum, tj, total G2Jac
	runningSum.Set(&g2Infinity)
	total.Set(&g2Infinity)
	for k := len(buckets) - 1; k >= 0; k-- {
		if !buckets[k].ZZ.IsZero() {
			runningSum.AddAssign(buckets[k].unsafeToJac(&tj))
		}
		total.AddAssign(&runningSum)
	}

	chRes <- total
	close(chRes)
}

func chunkReduceG2(p *G2Jac, c int, chTotals []chan G2Jac) chan G2Jac {
	chRes := make(chan G2Jac, 1)
	debug.Assert(len(chTotals) >= 2)
	go func() {
		totalj := <-chTotals[len(chTotals)-1]
		p.Set(&totalj)
		for j := len(chTotals) - 2; j >= 0; j-- {
			for l := 0; l < c; l++ {
				p.DoubleAssign()
			}
			totalj := <-chTotals[j]
			p.AddAssign(&totalj)
		}

		chRes <- *p
		close(chRes)
	}()

	return chRes
}
