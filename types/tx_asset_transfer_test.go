package types

import (
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
	}{
		{
			name:           "TestAssetsTransfer_1",
			senderSeed:     "yfcqxawkwvhnwwxnhxqbzufpnbxxvkpuueermpcxoiugqokwbmurqjq",
			senderIdentity: "LZTPJBQKOYLBFEWWFVEFDFOOFEWCUSSNNKLOXGDQJGBTYUMJVAOSYHIGYDOM",
			assetIssuer:    "CFBMEMZOIDEXQAUXYYSZIURADQLAPWPMNJXQSNVQZAHYVOPYUKKJBJUCTVJL",
			assetName:      "CFB",
			newOwner:       "UIJLDDELETUYEHFKZPQGVOOOTLHCNQWAZAXHLSXWMEDLRQEWKNSJVZIGFPBD",
			numberOfUnits:  1,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			payload, err := NewAssetTransferPayload(data.assetIssuer, data.newOwner, data.assetName, data.numberOfUnits)
			if err != nil {
				t.Fatalf("creating asset transfer payload: %s", err)
			}

			assetTransferTransaction, err := NewAssetTransferTransaction(data.senderIdentity, 0, 100, payload)
			if err != nil {
				t.Fatalf("creating asser transfer transaction: %s", err)
			}

			signer, _ := NewSigner(data.senderSeed)
			err = signer.SignTx(&assetTransferTransaction)
			if err != nil {
				t.Fatalf("signing asset transfer transaction: %s", err)
			}
		})
	}
}
