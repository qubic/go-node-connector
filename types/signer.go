package types

import (
	"github.com/pkg/errors"
	"github.com/qubic/go-schnorrq"
)

type Signer struct {
	seed string

	pubKey [32]byte
}

func NewSigner(seed string) (*Signer, error) {

	wallet, err := NewWallet(seed)
	if err != nil {
		return nil, errors.Wrap(err, "creating wallet")
	}

	pubKey := wallet.PubKey

	return &Signer{
		seed:   seed,
		pubKey: pubKey,
	}, nil
}

func (s *Signer) SignTx(tx *Transaction) error {

	if tx.SourcePublicKey != s.pubKey {
		return errors.New("source public key does not match")
	}

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
