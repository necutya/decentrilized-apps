package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type Signer struct {
	priv *ecdsa.PrivateKey
}

func New() *Signer {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return &Signer{priv: priv}
}

func (s *Signer) Sign(payload []byte) ([]byte, error) {
	h := sha256.Sum256(payload)
	return ecdsa.SignASN1(rand.Reader, s.priv, h[:])
}

func (s *Signer) PublicKeyPEM() string {
	der, err := x509.MarshalPKIXPublicKey(&s.priv.PublicKey)
	if err != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func VerifySignature(pubKeyPEM string, payload, sig []byte) error {
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return errors.New("invalid PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not an ECDSA public key")
	}
	h := sha256.Sum256(payload)
	if !ecdsa.VerifyASN1(ecPub, h[:], sig) {
		return errors.New("signature verification failed")
	}
	return nil
}
