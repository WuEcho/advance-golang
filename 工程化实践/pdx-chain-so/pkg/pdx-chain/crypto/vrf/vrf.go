/*************************************************************************
 * Copyright (C) 2016-2019 PDX Technologies, Inc. All Rights Reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * @Time   : 2020/10/15 9:58 上午
 * @Author : liangc
 *************************************************************************/

package vrf

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"math/big"
	"pdx-chain-so/pkg/pdx-chain/crypto/secp256k1"
)

/*
# Abstract:

We efficiently combine unpredictability and verifiability by extending the Goldreich-Goldwasser-Micali (1986)
construction of pseudorandom functions f/sub s/ from a secret seed s, so that knowledge of s not only enables
one to evaluate f/sub s/ at any point x, but also to provide an NP-proof that the value f/sub s/(x) is indeed
correct without compromising the unpredictability of f/sub s/ at any other point for which no such a proof
was provided.

http://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=814584
*/

// PrivateKey supports evaluating the VRF function.
type IPrivateKey interface {
	// Evaluate returns the output of H(f_k(m)) and its proof.
	Evaluate(m []byte) (index [32]byte, proof []byte)
	// Public returns the corresponding public key.
	Public() crypto.PublicKey
}

// PublicKey supports verifying output from the VRF function.
type IPublicKey interface {
	// ProofToHash verifies the NP-proof supplied by Proof and outputs Index.
	ProofToHash(m, proof []byte) (index [32]byte, err error)
}

var (
	curve  = secp256k1.S256()
	params = curve.Params()

	// ErrPointNotOnCurve occurs when a public key is not on the curve.
	ErrPointNotOnCurve = errors.New("point is not on the S256 curve")
	// ErrWrongKeyType occurs when a key is not an ECDSA key.
	ErrWrongKeyType = errors.New("not an ECDSA key")
	// ErrNoPEMFound occurs when attempting to parse a non PEM data structure.
	ErrNoPEMFound = errors.New("no PEM block found")
	// ErrInvalidVRF occurs when the VRF does not validate.
	ErrInvalidVRF = errors.New("invalid VRF proof")
)

// Unmarshal a compressed point in the form specified in section 4.3.6 of ANSI X9.62.
func Unmarshal(curve elliptic.Curve, data []byte) (x, y *big.Int) {
	byteLen := (curve.Params().BitSize + 7) >> 3
	if (data[0] &^ 1) != 2 {
		return // unrecognized point encoding
	}
	if len(data) != 1+byteLen {
		return
	}

	// Based on Routine 2.2.4 in NIST Mathematical routines paper
	params := curve.Params()
	tx := new(big.Int).SetBytes(data[1 : 1+byteLen])
	y2 := y2(params, tx)
	sqrt := defaultSqrt
	ty := sqrt(y2, params.P)
	if ty == nil {
		return // "y^2" is not a square: invalid point
	}
	var y2c big.Int
	y2c.Mul(ty, ty).Mod(&y2c, params.P)
	if y2c.Cmp(y2) != 0 {
		return // sqrt(y2)^2 != y2: invalid point
	}
	if ty.Bit(0) != uint(data[0]&1) {
		ty.Sub(params.P, ty)
	}

	x, y = tx, ty // valid point: return it
	return
}

// Use the curve equation to calculate y² given x.
// only applies to curves of the form y² = x³ - 3x + b.
func y2(curve *elliptic.CurveParams, x *big.Int) *big.Int {

	// y² = x³ - 3x + b
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)

	//threeX := new(big.Int).Lsh(x, 1)
	//threeX.Add(threeX, x)
	//
	//x3.Sub(x3, threeX)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)
	return x3
}

func defaultSqrt(x, p *big.Int) *big.Int {
	var r big.Int
	if nil == r.ModSqrt(x, p) {
		return nil // x is not a square
	}
	return &r
}

// PublicKey holds a public VRF key.
type PublicKey struct {
	*ecdsa.PublicKey
}

// PrivateKey holds a private VRF key.
type PrivateKey struct {
	*ecdsa.PrivateKey
}

// GenerateKey generates a fresh keypair for this VRF
func GenerateKey() (IPrivateKey, IPublicKey) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil
	}

	return &PrivateKey{PrivateKey: key}, &PublicKey{PublicKey: &key.PublicKey}
}

// h1 hashes m to a curve point
func h1(m []byte) (x, y *big.Int) {
	h := sha512.New()
	var i uint32
	byteLen := (params.BitSize + 7) >> 3
	for x == nil && i < 100 {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		r := []byte{2} // Set point encoding to "compressed", y=0.
		r = h.Sum(r)
		x, y = Unmarshal(curve, r[:byteLen+1])
		i++
	}
	return
}

var one = big.NewInt(1)

// h2 hashes to an integer [1,N-1]
func h2(m []byte) *big.Int {
	// NIST SP 800-90A § A.5.1: Simple discard method.
	byteLen := (params.BitSize + 7) >> 3
	h := sha512.New()
	for i := uint32(0); ; i++ {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		b := h.Sum(nil)
		k := new(big.Int).SetBytes(b[:byteLen])
		if k.Cmp(new(big.Int).Sub(params.N, one)) == -1 {
			return k.Add(k, one)
		}
	}
}

// Evaluate returns the verifiable unpredictable function evaluated at m
func (k PrivateKey) Evaluate(m []byte) (index [32]byte, proof []byte) {
	nilIndex := [32]byte{}
	// Prover chooses r <-- [1,N-1]
	r, _, _, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nilIndex, nil
	}
	ri := new(big.Int).SetBytes(r)

	// H = h1(m)
	Hx, Hy := h1(m)
	if !curve.IsOnCurve(Hx, Hy) {
		panic("not on curve")
	}
	// VRF_k(m) = [k]H
	sHx, sHy := curve.ScalarMult(Hx, Hy, k.D.Bytes())
	if !curve.IsOnCurve(sHx, sHy) {
		panic("not on curve2")
	}
	vrf := elliptic.Marshal(curve, sHx, sHy) // 65 bytes.
	uHx, _ := elliptic.Unmarshal(curve, vrf)
	if uHx == nil {
		panic("333")
	}
	// G is the base point
	// s = h2(G, H, [k]G, VRF, [r]G, [r]H)
	rGx, rGy := curve.ScalarBaseMult(r)
	rHx, rHy := curve.ScalarMult(Hx, Hy, r)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(curve, params.Gx, params.Gy))
	b.Write(elliptic.Marshal(curve, Hx, Hy))
	b.Write(elliptic.Marshal(curve, k.PublicKey.X, k.PublicKey.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(curve, rGx, rGy))
	b.Write(elliptic.Marshal(curve, rHx, rHy))
	s := h2(b.Bytes())

	// t = r−s*k mod N
	t := new(big.Int).Sub(ri, new(big.Int).Mul(s, k.D))
	t.Mod(t, params.N)

	// Index = H(vrf)
	index = sha256.Sum256(vrf)

	// Write s, t, and vrf to a proof blob. Also write leading zeros before s and t
	// if needed.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(s.Bytes())))
	buf.Write(s.Bytes())
	buf.Write(make([]byte, 32-len(t.Bytes())))
	buf.Write(t.Bytes())
	buf.Write(vrf)

	return index, buf.Bytes()
}

// ProofToHash asserts that proof is correct for m and outputs index.
func (pk *PublicKey) ProofToHash(m, proof []byte) (index [32]byte, err error) {
	nilIndex := [32]byte{}
	// verifier checks that s == h2(m, [t]G + [s]([k]G), [t]h1(m) + [s]VRF_k(m))
	if got, want := len(proof), 64+65; got != want {
		return nilIndex, ErrInvalidVRF
	}

	// Parse proof into s, t, and vrf.
	s := proof[0:32]
	t := proof[32:64]
	vrf := proof[64 : 64+65]

	uHx, uHy := elliptic.Unmarshal(curve, vrf)
	if uHx == nil {
		return nilIndex, ErrInvalidVRF
	}

	// [t]G + [s]([k]G) = [t+ks]G
	tGx, tGy := curve.ScalarBaseMult(t)
	ksGx, ksGy := curve.ScalarMult(pk.X, pk.Y, s)
	tksGx, tksGy := params.Add(tGx, tGy, ksGx, ksGy)

	// H = h1(m)
	// [t]H + [s]VRF = [t+ks]H
	Hx, Hy := h1(m)
	tHx, tHy := curve.ScalarMult(Hx, Hy, t)
	sHx, sHy := curve.ScalarMult(uHx, uHy, s)
	tksHx, tksHy := params.Add(tHx, tHy, sHx, sHy)

	//   h2(G, H, [k]G, VRF, [t]G + [s]([k]G), [t]H + [s]VRF)
	// = h2(G, H, [k]G, VRF, [t+ks]G, [t+ks]H)
	// = h2(G, H, [k]G, VRF, [r]G, [r]H)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(curve, params.Gx, params.Gy))
	b.Write(elliptic.Marshal(curve, Hx, Hy))
	b.Write(elliptic.Marshal(curve, pk.X, pk.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(curve, tksGx, tksGy))
	b.Write(elliptic.Marshal(curve, tksHx, tksHy))
	h2 := h2(b.Bytes())

	// Left pad h2 with zeros if needed. This will ensure that h2 is padded
	// the same way s is.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(h2.Bytes())))
	buf.Write(h2.Bytes())

	if !hmac.Equal(s, buf.Bytes()) {
		return nilIndex, ErrInvalidVRF
	}
	return sha256.Sum256(vrf), nil
}

// NewVRFSigner creates a signer object from a private key.
func NewVRFSigner(key *ecdsa.PrivateKey) (IPrivateKey, error) {
	if *(key.Params()) != *curve.Params() {
		return nil, ErrPointNotOnCurve
	}
	if !curve.IsOnCurve(key.X, key.Y) {
		return nil, ErrPointNotOnCurve
	}
	return &PrivateKey{PrivateKey: key}, nil
}

// Public returns the corresponding public key as bytes.
func (k PrivateKey) Public() crypto.PublicKey {
	return &k.PublicKey
}

// NewVRFVerifier creates a verifier object from a public key.
func NewVRFVerifier(pubkey *ecdsa.PublicKey) (IPublicKey, error) {
	if *(pubkey.Params()) != *curve.Params() {
		return nil, ErrPointNotOnCurve
	}
	if !curve.IsOnCurve(pubkey.X, pubkey.Y) {
		return nil, ErrPointNotOnCurve
	}
	return &PublicKey{PublicKey: pubkey}, nil
}

func VerifyRandom(pub *ecdsa.PublicKey, msg []byte, n [32]byte, proof []byte) error {
	if vrfPubkey, err := NewVRFVerifier(pub); err != nil {
		return err
	} else if _n, err := vrfPubkey.ProofToHash(msg, proof); err != nil {
		return err
	} else if _n == n {
		return nil
	}
	return errors.New("vrf verify fail")
}

func GenRandom(k *ecdsa.PrivateKey, msg []byte) (n [32]byte, proof []byte) {
	if vrfPrvkey, err := NewVRFSigner(k); err != nil {
		panic(err)
	} else {
		return vrfPrvkey.Evaluate(msg)
	}
}
