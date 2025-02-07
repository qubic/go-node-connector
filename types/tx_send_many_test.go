package types

import "testing"

func TestSendManyTransaction(t *testing.T) {

	testData := []struct {
		name           string
		senderSeed     string
		senderIdentity string
		transfers      map[string]int64
	}{
		{
			name:           "TestSendMany_1",
			senderSeed:     "yfcqxawkwvhnwwxnhxqbzufpnbxxvkpuueermpcxoiugqokwbmurqjq",
			senderIdentity: "LZTPJBQKOYLBFEWWFVEFDFOOFEWCUSSNNKLOXGDQJGBTYUMJVAOSYHIGYDOM",
			transfers: map[string]int64{
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA": 10,
				"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB": 20,
				"CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC": 30,
				"DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD": 40,
			},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			var transfers []SendManyTransfer

			for id, amount := range data.transfers {
				transfers = append(transfers, SendManyTransfer{
					AddressID: Identity(id),
					Amount:    amount,
				})
			}

			var payload SendManyTransferPayload
			err := payload.AddTransfers(transfers)
			if err != nil {
				t.Fatalf("adding transfers to send many payload")
			}

			sendManyTransaction, err := NewSendManyTransferTransaction(data.senderIdentity, 0, payload)
			if err != nil {
				t.Fatalf("creating send many transaction: %s", err)
			}

			signer, _ := NewSigner(data.senderSeed)
			sendManyTransaction, err = signer.SignTx(sendManyTransaction)
			if err != nil {
				t.Fatalf("signing send many transaction: %s", err)
			}

			expectedAmount := 10 + 20 + 30 + 40 + QutilSendManyFee

			if int64(expectedAmount) != sendManyTransaction.Amount {
				t.Fatal("transaction amount does not match expected value")
			}
		})
	}
}
