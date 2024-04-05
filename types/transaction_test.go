package types

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransaction_MarshallUnmarshall(t *testing.T) {
	initialTx := Transaction{
		SourcePublicKey:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		DestinationPublicKey: [32]byte{11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		Amount:               100,
		Tick:                 200,
		InputType:            300,
		InputSize:            10,
		Input:                []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Signature:            [64]byte{21, 22, 23, 24, 25, 26, 27, 28, 29, 30},
	}

	marshalled, err := initialTx.MarshallBinary()
	if err != nil {
		t.Fatalf("Got err when marshalling tx. err: %s", err.Error())
	}

	var unmarshalledTx Transaction
	err = unmarshalledTx.UnmarshallBinary(bytes.NewReader(marshalled))
	if err != nil {
		t.Fatalf("Got err when unmarshalling tx. err: %s", err.Error())
	}

	if cmp.Diff(initialTx, unmarshalledTx) != "" {
		t.Fatalf("Mismatched return value. Expected: %v, got: %v", initialTx, unmarshalledTx)
	}
}

func TestSendManyTransferPayload_Size(t *testing.T) {
	var payload SendManyTransferPayload
	b, err := payload.MarshallBinary()
	require.NoError(t, err, "binary marshalling payload")
	require.True(t, len(b) == QutilSendManyInputSize)
}
