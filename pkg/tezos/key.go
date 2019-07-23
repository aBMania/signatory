package tezos

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/crypto/blake2b"
	"github.com/ecadlabs/signatory/pkg/crypto"
)

const (
	publicKeyHashLenght = 20
)

func newUnkownPrefixErr(prefix string) error {
	return fmt.Errorf("Unkown prefix: %s", prefix)
}

// NewKeyPair create new tezos keypair
func NewKeyPair(publicKey string, secretKey string) *KeyPair {
	return &KeyPair{secretKey: secretKey, publicKey: publicKey}
}

// KeyPair a struct contains tezos formatted public key and secret key
type KeyPair struct {
	secretKey string
	publicKey string
}

func (k *KeyPair) getPubKeyPrefix() string {
	if len(k.publicKey) < pubKeyPrefixLength {
		return ""
	}
	return k.publicKey[0:pubKeyPrefixLength]
}

func (k *KeyPair) getPubKeyHashPrefix() string {
	switch k.getPubKeyPrefix() {
	case secp256k1PubKeyPrefix:
		return secp256k1PubKeyHashPrefix
	case p256PubKeyPrefix:
		return pS256PubKeyHashPrefix
	case ed25519PubKeyPrefix:
		return ed25519PubKeyHashPrefix
	}
	return ""
}

func (k *KeyPair) getSecretKeyPrefix() string {
	if len(k.secretKey) < secretKeyPrefixLength {
		return ""
	}
	return k.secretKey[0:secretKeyPrefixLength]
}

// PubKeyHash returns the pubkey hash for given keypair
func (k *KeyPair) PubKeyHash() (string, error) {
	publicKey, err := k.decodedPubKey()

	if err != nil {
		return "", err
	}

	hash := blake2b.SumX(publicKeyHashLenght, publicKey)
	prefix := prefixMap[k.getPubKeyHashPrefix()]
	return base58CheckEncodePrefix(prefix, hash[:publicKeyHashLenght]), nil
}

// Validate return an error if the keypair is invalid
func (k *KeyPair) Validate() error {
	err := k.validatePublicKey()
	if err != nil {
		return err
	}

	err = k.validateSecretKey()
	if err != nil {
		return err
	}

	if k.CurveName() == "" {
		return fmt.Errorf("Unable to determine curve matching the key pair")
	}

	return nil
}

func (k *KeyPair) validateSecretKey() error {
	prefix := k.getSecretKeyPrefix()

	prefixBytes, ok := prefixMap[prefix]

	if !ok {
		return newUnkownPrefixErr(prefix)
	}

	secretKey, err := decodeKey(prefixBytes, k.secretKey)
	if err != nil {
		return err
	}

	if len(secretKey) != 32 {
		return fmt.Errorf("Invalid secret key length")
	}

	return nil
}

func (k *KeyPair) decodedPubKey() ([]byte, error) {
	prefix := k.getPubKeyPrefix()

	prefixBytes, ok := prefixMap[prefix]

	if !ok {
		return nil, newUnkownPrefixErr(prefix)
	}

	return decodeKey(prefixBytes, k.publicKey)
}

func (k *KeyPair) validatePublicKey() error {
	publicKey, err := k.decodedPubKey()
	if err != nil {
		return err
	}

	if len(publicKey) != 33 && len(publicKey) != 65 {
		return fmt.Errorf("Invalid public key length: %d", len(publicKey))
	}

	return nil
}

// X returns the x coordinate of the public key
func (k *KeyPair) X() []byte {
	x, _ := crypto.ECCoordinateFromPrivateKey(k.D(), k.CurveName())
	return x
}

// Y returns the y coordinate of the public key
func (k *KeyPair) Y() []byte {
	_, y := crypto.ECCoordinateFromPrivateKey(k.D(), k.CurveName())
	return y
}

// CurveName returns the curveName used by this keypair
func (k *KeyPair) CurveName() string {
	hash, err := k.PubKeyHash()

	if err != nil {
		return ""
	}

	return getCurveFromPubkeyHash(hash)
}

// D return the D parameter of the elliptic curve
func (k *KeyPair) D() []byte {
	prefix := k.getSecretKeyPrefix()

	prefixBytes, ok := prefixMap[prefix]

	if !ok {
		return nil
	}

	secretKey, _ := decodeKey(prefixBytes, k.secretKey)
	return secretKey
}

// EncodeASN1 Return ASN1 encoding of the key pair
func (k *KeyPair) EncodeASN1() ([]byte, error) {
	if k.CurveName() == crypto.CurveED25519 {
		return nil, errors.New("Ed25519 is not yet supported")
	}

	key := big.NewInt(0).SetBytes(k.D())
	oid, _ := crypto.OIDFromNamedCurve(k.CurveName())
	type ecPrivateKey struct {
		Version       int
		PrivateKey    []byte
		NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
		PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
	}
	pk := &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{
		Curve: crypto.GetCurve(k.CurveName()),
		X:     big.NewInt(0).SetBytes(k.X()),
		Y:     big.NewInt(0).SetBytes(k.Y()),
	}, D: key}

	// This is the same as x509.MarshalECPrivateKey but we need to duplicate the implementation in order to support more curve
	paddedPrivateKey := make([]byte, (pk.Curve.Params().N.BitLen()+7)/8)

	return asn1.Marshal(ecPrivateKey{
		Version:       1,
		PrivateKey:    paddedPrivateKey,
		NamedCurveOID: oid,
		PublicKey:     asn1.BitString{Bytes: elliptic.Marshal(pk.Curve, pk.X, pk.Y)},
	})
}
