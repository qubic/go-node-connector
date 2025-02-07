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

// SignTx Returns the signed transaction. The original transaction object is not modified, and the returned value should be used after signing.
func (s *Signer) SignTx(tx Transaction) (Transaction, error) {

	if tx.SourcePublicKey != s.pubKey {
		return Transaction{}, errors.New("source public key does not match signer")
	}

	subSeed, err := GetSubSeed(s.seed)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "getting sub-seed")
	}

	unsignedDigest, err := tx.GetUnsignedDigest()
	if err != nil {
		return Transaction{}, errors.Wrap(err, "getting unsigned transaction digest")
	}

	signature, err := schnorrq.Sign(subSeed, tx.SourcePublicKey, unsignedDigest)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "creating signature")
	}

	return Transaction{
		SourcePublicKey:      tx.SourcePublicKey,
		DestinationPublicKey: tx.DestinationPublicKey,
		Amount:               tx.Amount,
		Tick:                 tx.Tick,
		InputType:            tx.InputType,
		InputSize:            tx.InputSize,
		Input:                tx.Input,
		Signature:            signature,
	}, nil
}
