package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const QxAddress = "BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARMID"
const QxTransferInputType = 2
const QxTransferInputSize = 32 + 32 + 8 + 8

type AssetTransferPayload struct {
	issuer               [32]byte
	newOwnerAndPossessor [32]byte
	assetName            [8]uint8
	numberOfUnits        int64
}

func NewAssetTransferPayload(assetName, issuer, newOwnerAndPossessor string, numberOfUnits int64) (AssetTransferPayload, error) {

	issuerIdentity := Identity(issuer)
	issuerPubKey, err := issuerIdentity.ToPubKey(false)
	if err != nil {
		return AssetTransferPayload{}, fmt.Errorf("failed to obtain issuer public key: %w", err)
	}

	newOwnerAndPossessorIdentity := Identity(newOwnerAndPossessor)
	newOwnerAndPossessorPubKey, err := newOwnerAndPossessorIdentity.ToPubKey(false)
	if err != nil {
		return AssetTransferPayload{}, fmt.Errorf("failed to obtain new owner public key: %w", err)
	}

	if len(assetName) > 7 {
		return AssetTransferPayload{}, fmt.Errorf("asset name '%s' is longer than 7", assetName)
	}

	var assetNameBytes [8]byte
	copy(assetNameBytes[:], assetName)

	return AssetTransferPayload{
		issuer:               issuerPubKey,
		newOwnerAndPossessor: newOwnerAndPossessorPubKey,
		assetName:            assetNameBytes,
		numberOfUnits:        numberOfUnits,
	}, nil
}

func (atp *AssetTransferPayload) MarshallBinary() ([]byte, error) {

	var buff bytes.Buffer

	err := binary.Write(&buff, binary.LittleEndian, atp.issuer)
	if err != nil {
		return nil, fmt.Errorf("writing issuer public key to buffer: %w", err)
	}

	err = binary.Write(&buff, binary.LittleEndian, atp.newOwnerAndPossessor)
	if err != nil {
		return nil, fmt.Errorf("writing new owner and possessor public key to buffer: %w", err)
	}

	err = binary.Write(&buff, binary.LittleEndian, atp.assetName)
	if err != nil {
		return nil, fmt.Errorf("writing asset name to buffer: %w", err)
	}

	err = binary.Write(&buff, binary.LittleEndian, atp.numberOfUnits)
	if err != nil {
		return nil, fmt.Errorf("writing number of units to buffer: %w", err)
	}

	return buff.Bytes(), nil
}

func NewAssetTransferTransaction(sourceID string, targetTick uint32, transferFee int64, payload AssetTransferPayload) (Transaction, error) {

	sourceIdentity := Identity(sourceID)
	sourcePublicKey, err := sourceIdentity.ToPubKey(false)
	if err != nil {
		return Transaction{}, fmt.Errorf("converting source id to public key: %w", err)
	}

	destinationIdentity := Identity(QxAddress)
	destinationPublicKey, err := destinationIdentity.ToPubKey(false)
	if err != nil {
		return Transaction{}, fmt.Errorf("converting destination id to public key: %w", err)
	}

	input, err := payload.MarshallBinary()
	if err != nil {
		return Transaction{}, fmt.Errorf("marshalling transaction payload to binary format: %w", err)
	}

	return Transaction{
		SourcePublicKey:      sourcePublicKey,
		DestinationPublicKey: destinationPublicKey,
		Amount:               transferFee,
		Tick:                 targetTick,
		InputType:            QxTransferInputType,
		InputSize:            uint16(len(input)),
		Input:                input,
	}, nil

}
