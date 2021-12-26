package jwt

import (
	"crypto"
	"crypto/rand"
	"errors"
	"math/big"
	"pdx-chain-so/pkg/pdx-chain/crypto/gmsm/sm2"
)

var (
	ErrSM2Verification = errors.New("sm2: verification error")
)

// Implements the SM2 signing methods
// Expects *sm2.PrivateKey for signing and *sm2.PublicKey for verification
type SigningMethodSM2 struct {
	Name      string
	Hash      crypto.Hash
}

// Specific instances for SM2
var (
	SM2Signing *SigningMethodSM2
)

func init() {
	// SM2
	SM2Signing = &SigningMethodSM2{"SM2", crypto.SHA256}
	RegisterSigningMethod(SM2Signing.Alg(), func() SigningMethod {
		return SM2Signing
	})
}

func (m *SigningMethodSM2) Alg() string {
	return m.Name
}

// Implements the Verify method from SigningMethod
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *SigningMethodSM2) Verify(signingString, signature string, key interface{}) error {
	var err error

	// Decode the signature
	var sig []byte
	if sig, err = DecodeSegment(signature); err != nil {
		return err
	}

	// Get the key
	var sm2Key *sm2.PublicKey
	switch k := key.(type) {
	case *sm2.PublicKey:
		sm2Key = k
	default:
		return ErrInvalidKeyType
	}

	if len(sig) != 2*32 {
		return ErrSM2Verification
	}

	r := big.NewInt(0).SetBytes(sig[:32])
	s := big.NewInt(0).SetBytes(sig[32:])

	// Create hasher
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}
	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Verify the signature
	if verifystatus := sm2.Sm2Verify(sm2Key, hasher.Sum(nil), nil, r, s); verifystatus == true {
		return nil
	} else {
		return ErrSM2Verification
	}
}

// Implements the Sign method from SigningMethod
// For this signing method, key must be an ecdsa.PrivateKey struct
func (m *SigningMethodSM2) Sign(signingString string, key interface{}) (string, error) {
	// Get the key
	var ecdsaKey *sm2.PrivateKey
	switch k := key.(type) {
	case *sm2.PrivateKey:
		ecdsaKey = k
	default:
		return "", ErrInvalidKeyType
	}

	// Create the hasher
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := m.Hash.New()
	hasher.Write([]byte(signingString))

	// Sign the string and return r, s
	if r, s, err := sm2.Sm2Sign(ecdsaKey, hasher.Sum(nil), nil, rand.Reader); err == nil {
		curveBits := ecdsaKey.Curve.Params().BitSize

		keyBytes := curveBits / 8
		if curveBits%8 > 0 {
			keyBytes += 1
		}

		// We serialize the outpus (r and s) into big-endian byte arrays and pad
		// them with zeros on the left to make sure the sizes work out. Both arrays
		// must be keyBytes long, and the output must be 2*keyBytes long.
		rBytes := r.Bytes()
		rBytesPadded := make([]byte, keyBytes)
		copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

		sBytes := s.Bytes()
		sBytesPadded := make([]byte, keyBytes)
		copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

		out := append(rBytesPadded, sBytesPadded...)

		return EncodeSegment(out), nil
	} else {
		return "", err
	}

}
