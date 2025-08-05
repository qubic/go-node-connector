package qubic

import (
	"encoding/hex"
	"github.com/qubic/go-node-connector/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQubic_serializeBinary_requestIssuedAssetsByByUniverseIndex_thenProduceCorrectBinary(t *testing.T) {
	request := RequestAssetsByUniverseIndex{
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
	request, err := createByFilterRequest(requestTypeAssetOwnershipRecords, AssetInformation{"", "QX"},
		AssetHolderInformation{"", 0},
		AssetHolderInformation{"", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "01007800000000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_givenOwner_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetOwnershipsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0100700000000000000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c90000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_givenContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetOwnershipsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"", 1})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "01006800010000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestOwnedAssetsByFilter_givenOwnerAndContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetOwnershipsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 1})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0100600001000000000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c90000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"}, AssetHolderInformation{"", 0}, AssetHolderInformation{"", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "02007800000000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenOwner_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 0},
		AssetHolderInformation{"", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0200700000000000000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c90000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenOwnerContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"", 1},
		AssetHolderInformation{"", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "02006800010000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenOwnerAndOwnerContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 1},
		AssetHolderInformation{"", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0200600001000000000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c90000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenPossessor_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"", 0},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 0})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "02005800000000000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c9"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenPossessorContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"", 0},
		AssetHolderInformation{"", 1})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "02003800000001000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenPossessorAndPossessorContract_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"", "QX"},
		AssetHolderInformation{"", 0},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 1})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "02001800000001000000000000000000000000000000000000000000000000000000000000000000515800000000000000000000000000000000000000000000000000000000000000000000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c9"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func TestQubic_serializeBinary_requestPossessedAssetsByFilter_givenAllFilters_thenProduceCorrectBinary(t *testing.T) {
	request, err := createGetAssetPossessionsByFilterRequest(AssetInformation{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "QX"},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 1},
		AssetHolderInformation{"KXRSTAAGZKJZCCSHJKCSPTUSUZTAIESNWZJZRTFMBAIVTIPXPUYCFYVFWAZL", 1})
	assert.NoError(t, err)

	bytes, err := serializeBinary(request)
	assert.NoError(t, err)

	expectedHex := "0200000001000100000000000000000000000000000000000000000000000000000000000000000051580000000000004477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c94477ab04b56ece48bccf40c617fd791a4088d1893a65f201a694abc60d5035c9"
	assert.Equal(t, expectedHex, hex.EncodeToString(bytes))
}

func Test_getTickTransactionsNrTx(t *testing.T) {
	tickData := types.TickData{
		ComputorIndex: 0,
		Epoch:         0,
		Tick:          0,
		Millisecond:   0,
		Second:        0,
		Minute:        0,
		Hour:          0,
		Day:           0,
		Month:         0,
		Year:          0,
		Timelock:      [32]byte{},
		TransactionDigests: [1024][32]byte{
			{0x01, 0x02, 0x03, 0x04},
			{0x05, 0x06, 0x07, 0x08},
			[32]byte{},
			{0x09, 0x0A, 0x0B, 0x0C},
			[32]byte{},
		},
		ContractFees: [1024]int64{},
		Signature:    [64]byte{},
	}

	nrTx := getTickTransactionsNrTx(tickData)
	assert.Equal(t, 4, nrTx)

	tickData = types.TickData{
		ComputorIndex: 0,
		Epoch:         0,
		Tick:          0,
		Millisecond:   0,
		Second:        0,
		Minute:        0,
		Hour:          0,
		Day:           0,
		Month:         0,
		Year:          0,
		Timelock:      [32]byte{},
		TransactionDigests: [1024][32]byte{
			[32]byte{},
			[32]byte{},
		},
		ContractFees: [1024]int64{},
		Signature:    [64]byte{},
	}

	nrTx = getTickTransactionsNrTx(tickData)
	assert.Equal(t, 0, nrTx)
}
