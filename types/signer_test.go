package types

import (
	"github.com/qubic/go-schnorrq"
	"testing"
)

func TestSigner_SignTx(t *testing.T) {

	testData := []struct {
		name                string
		senderSeed          string
		senderIdentity      string
		destinationIdentity string
	}{
		{
			name:                "TestSign_1",
			senderSeed:          "yfcqxawkwvhnwwxnhxqbzufpnbxxvkpuueermpcxoiugqokwbmurqjq",
			senderIdentity:      "LZTPJBQKOYLBFEWWFVEFDFOOFEWCUSSNNKLOXGDQJGBTYUMJVAOSYHIGYDOM",
			destinationIdentity: "UIJLDDELETUYEHFKZPQGVOOOTLHCNQWAZAXHLSXWMEDLRQEWKNSJVZIGFPBD",
		},
		{
			name:                "TestSign_2",
			senderSeed:          "oqrtktmxmowfwpliikiogiczvpelmuaamreundljwnnpjojvtsfhtgd",
			senderIdentity:      "UIJLDDELETUYEHFKZPQGVOOOTLHCNQWAZAXHLSXWMEDLRQEWKNSJVZIGFPBD",
			destinationIdentity: "ZFEEMHFUDDGUJBFXDVHXOHKDSELCAWDCUTASNOAMQDTZWILDTCDCSNQGHEGN",
		},
		{
			name:                "TestSign_3",
			senderSeed:          "eivjjmrusievohpkmqaxanvvkbsglxigqevbnxfqswmtuxbhmphjpzw",
			senderIdentity:      "ZFEEMHFUDDGUJBFXDVHXOHKDSELCAWDCUTASNOAMQDTZWILDTCDCSNQGHEGN",
			destinationIdentity: "COVLRIWUCHTKZFGXEFYFSVNWDXECAYJXXSLSKDETUCCDMTRTAWCLSFOCJJSA",
		},
		{
			name:                "TestSign_4",
			senderSeed:          "edsllpxbhvsrqdnhxpinwabnmkgjyrbszgbtcmuertkefhmsqtptcgj",
			senderIdentity:      "COVLRIWUCHTKZFGXEFYFSVNWDXECAYJXXSLSKDETUCCDMTRTAWCLSFOCJJSA",
			destinationIdentity: "CSPGQLVUJIUCFESXLVBHVYSGDRLDJNVUGWMKQLPYNFKRPQAAFOKILXBFIIUJ",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {

			tx, err := NewSimpleTransferTransaction(data.senderIdentity, data.destinationIdentity, 0, 0)
			if err != nil {
				t.Fatalf("creating simple transfer transaction: %s", err)
			}

			signer, err := NewSigner(data.senderSeed)
			if err != nil {
				t.Fatalf("creating signer: %s", err)
			}

			tx, err = signer.SignTx(tx)
			if err != nil {
				t.Fatalf("signing tx: %s", err)
			}

			unsignedDigest, err := tx.GetUnsignedDigest()
			if err != nil {
				t.Fatalf("getting unsigned digest: %s", err)
			}

			err = schnorrq.Verify(tx.SourcePublicKey, unsignedDigest, tx.Signature)
			if err != nil {
				t.Fatalf("verifying signature: %s", err)
			}

		})
	}

}
