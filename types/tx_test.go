package types

import (
	"testing"
)

func TestSimpleTransaction(t *testing.T) {

	testData := []struct {
		name            string
		senderSeed      string
		destinationSeed string
		amount          int64
	}{
		{
			name:            "TestSimpleRandom_1",
			senderSeed:      GenerateRandomSeed(),
			destinationSeed: GenerateRandomSeed(),
			amount:          0,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			senderWallet, err := NewWallet(data.senderSeed)
			if err != nil {
				t.Fatalf("creating sender wallet: %s", err)
			}

			destinationWallet, err := NewWallet(data.destinationSeed)
			if err != nil {
				t.Fatalf("creating destination wallet: %s", err)
			}

			senderIdentityString := senderWallet.Identity.String()
			destinationIdentityString := destinationWallet.Identity.String()

			/*lsc := NewLiveServiceClient("http://localhost:8080")

			currentTickInfo, err := lsc.GetTickInfo()
			if err != nil {
				t.Fatalf("getting current tick info: %s", err)
			}

			targetTick := currentTickInfo.TickInfo.Tick + 15 // Apply tick offset*/

			simpleTransaction, err := NewSimpleTransferTransaction(senderIdentityString, destinationIdentityString, data.amount, 0 /*targetTick*/)
			if err != nil {
				t.Fatalf("creating simple transaction: %s", err)
			}

			signer := NewSigner(data.senderSeed)
			err = signer.SignTx(&simpleTransaction)
			if err != nil {
				t.Fatalf("signing simple transaction: %s", err)
			}

			/*broadcastResponse, err := lsc.BroadcastTransaction(simpleTransaction)
			if err != nil {
				t.Fatalf("broadcasting simple transaction: %s", err)
			}

			encodedTransaction, err := simpleTransaction.EncodeToBase64()
			if err != nil {
				t.Fatalf("encoding transaction to base64 for verification")
			}

			if broadcastResponse.EncodedTransaction != encodedTransaction {
				t.Fatal("encoded transaction check failed")
			}*/

		})
	}

}

func TestSendManyTransaction(t *testing.T) {

	testData := []struct {
		name       string
		senderSeed string
		transfers  map[string]int64
	}{
		{
			name:       "TestSendManyRandom_1",
			senderSeed: GenerateRandomSeed(),
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

			senderWallet, err := NewWallet(data.senderSeed)
			if err != nil {
				t.Fatalf("creating sender wallet: %s", err)
			}

			senderIdentityString := senderWallet.Identity.String()

			var transfers []SendManyTransfer

			for id, amount := range data.transfers {

				transfers = append(transfers, SendManyTransfer{
					AddressID: Identity(id),
					Amount:    amount,
				})
			}

			var payload SendManyTransferPayload
			err = payload.AddTransfers(transfers)
			if err != nil {
				t.Fatalf("adding transfers to send many payload")
			}

			/*lsc := NewLiveServiceClient("http://localhost:8080")

			currentTickInfo, err := lsc.GetTickInfo()
			if err != nil {
				t.Fatalf("getting current tick info: %s", err)
			}

			targetTick := currentTickInfo.TickInfo.Tick + 15 // Apply tick offset*/

			sendManyTransaction, err := NewSendManyTransferTransaction(senderIdentityString, 0 /*targetTick*/, payload)
			if err != nil {
				t.Fatalf("creating send many transaction: %s", err)
			}

			signer := NewSigner(data.senderSeed)
			err = signer.SignTx(&sendManyTransaction)
			if err != nil {
				t.Fatalf("signing send many transaction: %s", err)
			}

			expectedAmount := 10 + 20 + 30 + 40 + QutilSendManyFee

			if int64(expectedAmount) != sendManyTransaction.Amount {
				t.Fatal("transaction amount does not match expected value")
			}

			/*broadcastResponse, err := lsc.BroadcastTransaction(sendManyTransaction)
			if err != nil {
				t.Fatalf("broadcasting send many transaction: %s", err)
			}

			encodedTransaction, err := sendManyTransaction.EncodeToBase64()
			if err != nil {
				t.Fatalf("encoding transaction to base64 for verification")
			}

			if broadcastResponse.EncodedTransaction != encodedTransaction {
				t.Fatal("encoded transaction check failed")
			}*/

		})
	}

}

func TestAssetTransaction(t *testing.T) {

}
