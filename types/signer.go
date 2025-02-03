package types

import (
	"github.com/pkg/errors"
	"github.com/qubic/go-schnorrq"
)

type Signer struct {
	seed string
}

func NewSigner(seed string) *Signer {
	return &Signer{
		seed: seed,
	}
}

func (s *Signer) SignTx(tx *Transaction) error {

	subSeed, err := GetSubSeed(s.seed)
	if err != nil {
		return errors.Wrap(err, "getting sub-seed")
	}

	unsignedDigest, err := tx.GetUnsignedDigest()
	if err != nil {
		return errors.Wrap(err, "getting unsigned transaction digest")
	}

	signature, err := schnorrq.Sign(subSeed, tx.SourcePublicKey, unsignedDigest)
	if err != nil {
		return errors.Wrap(err, "creating signature")
	}
	tx.Signature = signature

	return nil
}
