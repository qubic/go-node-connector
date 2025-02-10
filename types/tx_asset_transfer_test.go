package types

import (
	"bytes"
	"encoding/binary"
	"github.com/qubic/go-schnorrq"
	"testing"
)

func TestAssetTransaction(t *testing.T) {

	testData := []struct {
		name           string
		senderSeed     string
		senderIdentity string
		assetIssuer    string
		assetName      string
		newOwner       string
		numberOfUnits  int64
		transferFee    int64
		targetTick     uint32
	}{
		{
			name:           "TestAssetsTransfer_1",
			senderSeed:     "yfcqxawkwvhnwwxnhxqbzufpnbxxvkpuueermpcxoiugqokwbmurqjq",
			senderIdentity: "LZTPJBQKOYLBFEWWFVEFDFOOFEWCUSSNNKLOXGDQJGBTYUMJVAOSYHIGYDOM",
			assetIssuer:    "CFBMEMZOIDEXQAUXYYSZIURADQLAPWPMNJXQSNVQZAHYVOPYUKKJBJUCTVJL",
			assetName:      "CFB",
			newOwner:       "UIJLDDELETUYEHFKZPQGVOOOTLHCNQWAZAXHLSXWMEDLRQEWKNSJVZIGFPBD",
			numberOfUnits:  1,
			transferFee:    100,
			targetTick:     12332145,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			payload, err := NewAssetTransferPayload(data.assetName, data.assetIssuer, data.newOwner, data.numberOfUnits)
			if err != nil {
				t.Fatalf("creating asset transfer payload: %s", err)
			}

			assetTransferTransaction, err := NewAssetTransferTransaction(data.senderIdentity, data.targetTick, data.transferFee, payload)
			if err != nil {
				t.Fatalf("creating asser transfer transaction: %s", err)
			}

			if data.transferFee != assetTransferTransaction.Amount {
				t.Fatalf("asset transfer transaction amount does not match expected value. expected %d, got %d", data.transferFee, assetTransferTransaction.Amount)
			}

			if assetTransferTransaction.InputType != QxTransferInputType {
				t.Fatalf("asset transfer transaction input type does not match expected value. expected %d, got %d", QxTransferInputType, assetTransferTransaction.InputType)
			}

			if assetTransferTransaction.InputSize != QxTransferInputSize {
				t.Fatalf("asset transfer transaction input size does not match expected value. expected %d, got %d", QxTransferInputSize, assetTransferTransaction.InputSize)
			}

			if data.targetTick != assetTransferTransaction.Tick {
				t.Fatalf("asset transfer target tick does not match expected value. expected %d, got %d", data.targetTick, assetTransferTransaction.Tick)
			}

			signer, _ := NewSigner(data.senderSeed)
			assetTransferTransaction, err = signer.SignTx(assetTransferTransaction)
			if err != nil {
				t.Fatalf("signing asset transfer transaction: %s", err)
			}

			unsignedDigest, err := assetTransferTransaction.GetUnsignedDigest()
			if err != nil {
				t.Fatalf("getting unsigned digest for asset transfer transaction")
			}

			err = schnorrq.Verify(assetTransferTransaction.SourcePublicKey, unsignedDigest, assetTransferTransaction.Signature)
			if err != nil {
				t.Fatalf("verifying asset transfer transaction signature: %s", err)
			}

		})
	}
}

func TestAssetTransferPayload_MarshallBinary(t *testing.T) {

	assetName := "CFB"
	assetIssuer := "CFBMEMZOIDEXQAUXYYSZIURADQLAPWPMNJXQSNVQZAHYVOPYUKKJBJUCTVJL"
	newOwner := "UIJLDDELETUYEHFKZPQGVOOOTLHCNQWAZAXHLSXWMEDLRQEWKNSJVZIGFPBD"
	numberOfUnits := int64(123432123)

	payload, err := NewAssetTransferPayload(assetName, assetIssuer, newOwner, numberOfUnits)
	if err != nil {
		t.Fatalf("creating asset transfer payload: %s", err)
	}

	binaryData, err := payload.MarshallBinary()
	if err != nil {
		t.Fatal("marshalling asset transfer payload to binary")
	}

	if len(binaryData) != 80 {
		t.Fatal("binary asset transfer payload does not match the expected length")
	}

	assetIssuerIdentity := Identity(assetIssuer)
	assetIssuerPublicKey, err := assetIssuerIdentity.ToPubKey(false)
	if err != nil {
		t.Fatalf("getting asset issuer public key: %s", err)
	}

	newOwnerIdentity := Identity(newOwner)
	newOwnerPublicKey, err := newOwnerIdentity.ToPubKey(false)
	if err != nil {
		t.Fatalf("getting new asset owner public key: %s", err)
	}

	var assetNameBytes [8]byte
	copy(assetNameBytes[:], assetName)

	var numberOfUnitsBytes [8]byte
	binary.LittleEndian.PutUint64(numberOfUnitsBytes[:], uint64(numberOfUnits)) // casting to uint64 works here, as it does not change the data, only the way it is interpreted

	if !bytes.Equal(binaryData[:32], assetIssuerPublicKey[:]) {
		t.Fatal("asset issuer public key does not match expected value")
	}

	if !bytes.Equal(binaryData[32:64], newOwnerPublicKey[:]) {
		t.Fatal("new asset owner public key does not match expected value")
	}

	if !bytes.Equal(binaryData[64:72], assetNameBytes[:]) {
		t.Fatal("asset name does not match expected value")
	}

	if !bytes.Equal(binaryData[72:80], numberOfUnitsBytes[:]) {
		t.Fatal("number of units does not match expected value")
	}

}
