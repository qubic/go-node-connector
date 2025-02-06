package types

import "github.com/pkg/errors"

func NewSimpleTransferTransaction(sourceID, destinationID string, amount int64, targetTick uint32) (Transaction, error) {
	srcID := Identity(sourceID)
	destID := Identity(destinationID)
	srcPubKey, err := srcID.ToPubKey(false)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "converting src id string to pubkey")
	}
	destPubKey, err := destID.ToPubKey(false)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "converting dest id string to pubkey")
	}

	return Transaction{
		SourcePublicKey:      srcPubKey,
		DestinationPublicKey: destPubKey,
		Amount:               amount,
		Tick:                 targetTick,
		InputType:            0,
		InputSize:            0,
		Input:                nil,
	}, nil
}
