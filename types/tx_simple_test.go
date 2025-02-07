package types

import "testing"

func TestSimpleTransaction(t *testing.T) {

	testData := []struct {
		name                string
		senderSeed          string
		senderIdentity      string
		destinationIdentity string
	}{
		{
			name:                "TestSimple_1",
			senderSeed:          "yfcqxawkwvhnwwxnhxqbzufpnbxxvkpuueermpcxoiugqokwbmurqjq",
			senderIdentity:      "LZTPJBQKOYLBFEWWFVEFDFOOFEWCUSSNNKLOXGDQJGBTYUMJVAOSYHIGYDOM",
			destinationIdentity: "UIJLDDELETUYEHFKZPQGVOOOTLHCNQWAZAXHLSXWMEDLRQEWKNSJVZIGFPBD",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			simpleTransaction, err := NewSimpleTransferTransaction(data.senderIdentity, data.destinationIdentity, 0, 0)
			if err != nil {
				t.Fatalf("creating simple transaction: %s", err)
			}

			signer, _ := NewSigner(data.senderSeed)
			simpleTransaction, err = signer.SignTx(simpleTransaction)
			if err != nil {
				t.Fatalf("signing simple transaction: %s", err)
			}

		})
	}
}
