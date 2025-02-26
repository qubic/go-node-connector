package qubic

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQubic_serializeBinary_givenGetAssetByUniverseIndexPayload_thenProduceCorrectBinary(t *testing.T) {
	request := requestAssetsByUniverseIndex{
		RequestType:   RequestTypeAssetByUniverseIndex,
		UniverseIndex: 4,
	}

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "03000000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}
