package qubic

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQubic_serializeBinary_requestIssuedAssetsByByUniverseIndex_thenProduceCorrectBinary(t *testing.T) {
	request := requestAssetsByUniverseIndex{
		RequestType:   requestTypeAssetByUniverseIndex,
		UniverseIndex: 4,
	}

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "03000000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestIssuedAssetsByFilter_thenProduceCorrectBinary(t *testing.T) {

	request, err := createAssetIssuancesByFilterRequest("", "")
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "00000600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestIssuedAssetsByFilter_givenIssuer_thenProduceCorrectBinary(t *testing.T) {

	request, err := createAssetIssuancesByFilterRequest("CFBMEMZOIDEXQAUXYYSZIURADQLAPWPMNJXQSNVQZAHYVOPYUKKJBJUCTVJL", "")
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "00000400000000000830bb63bf7d5e164ac8cbd38680630ff7670a1ebf39f7210b40bcdca253d05f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestIssuedAssetsByFilter_givenAssetName_thenProduceCorrectBinary(t *testing.T) {

	request, err := createAssetIssuancesByFilterRequest("", "RANDOM")
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0000020000000000000000000000000000000000000000000000000000000000000000000000000052414e444f4d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestIssuedAssetsByFilter_givenIssuerAndAssetName_thenProduceCorrectBinary(t *testing.T) {

	request, err := createAssetIssuancesByFilterRequest("CFBMEMZOIDEXQAUXYYSZIURADQLAPWPMNJXQSNVQZAHYVOPYUKKJBJUCTVJL", "CFB")
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "00000000000000000830bb63bf7d5e164ac8cbd38680630ff7670a1ebf39f7210b40bcdca253d05f434642000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_thenProduceCorrectBinary(t *testing.T) {
	request, err := createAssetOwnershipByFilterRequest("", "QX", "", 0)
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "01007800000000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_givenOwner_thenProduceCorrectBinary(t *testing.T) {
	request, err := createAssetOwnershipByFilterRequest("", "QX",
		"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 0)
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0100700000000000000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c90000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_givenContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createAssetOwnershipByFilterRequest("", "QX", "", 1)
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "01006800010000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_givenOwnerAndContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createAssetOwnershipByFilterRequest("", "QX",
		"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 1)
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0100600001000000000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c90000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}
