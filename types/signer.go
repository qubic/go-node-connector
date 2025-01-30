package types

import (
	"github.com/pkg/errors"
	"github.com/qubic/go-schnorrq"
)

var DefaultSigner = Signer{
	SignFunc: schnorrq.Sign,
}

type Signer struct {
	SignFunc func(subSeed [32]byte, pubKey [32]byte, messageDigest [32]byte) ([64]byte, error)
}

func NewSigner(signFunc func(subSeed [32]byte, pubKey [32]byte, messageDigest [32]byte) ([64]byte, error)) *Signer {
	return &Signer{SignFunc: signFunc}
}

func (s *Signer) SignTx(tx *Transaction, sourceSeed string) error {

	subSeed, err := GetSubSeed(sourceSeed)
	if err != nil {
		return errors.Wrap(err, "getting sub-seed")
	}

	unsignedDigest, err := tx.GetUnsignedDigest()
	if err != nil {
		return errors.Wrap(err, "getting unsigned transaction digest")
	}

	signature, err := s.SignFunc(subSeed, tx.SourcePublicKey, unsignedDigest)
	if err != nil {
		return errors.Wrap(err, "creating signature")
	}
	tx.Signature = signature

	return nil
}
