package types

import "fmt"

func NewSimpleTransferTransaction(sourceID, destinationID string, amount int64, targetTick uint32) (Transaction, error) {
	srcID := Identity(sourceID)
	destID := Identity(destinationID)
	srcPubKey, err := srcID.ToPubKey(false)
	if err != nil {
		return Transaction{}, fmt.Errorf("converting src id string to pubkey: %w", err)
	}
	destPubKey, err := destID.ToPubKey(false)
	if err != nil {
		return Transaction{}, fmt.Errorf("converting dest id string to pubkey: %w", err)
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
