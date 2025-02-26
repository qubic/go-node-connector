package types

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAssets_AssetIssuance_UnmarshallFromReader(t *testing.T) {

	// RESPOND_ASSETS header + payload + END_RESPONSE header
	hexStr := "400000351cd1f26200000000000000000000000000000000000000000000000000000000000000000152414e444f4d000000000000000000f4f4320103000000080000231cd1f262"

	issuedAssetsBin, err := hex.DecodeString(hexStr)
	require.NoError(t, err)

	var result AssetIssuances
	err = result.UnmarshallFromReader(bytes.NewReader(issuedAssetsBin))
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assetIssuance := result[0].Asset
	assert.Equal(t, int8(0), assetIssuance.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{82, 65, 78, 68, 79, 77, 0}, assetIssuance.Name) // RANDOM
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, assetIssuance.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, assetIssuance.PublicKey)
	assert.Equal(t, uint8(1), assetIssuance.Type)
	index := result[0].UniverseIndex
	assert.Equal(t, uint32(3), index)
	tick := result[0].Tick
	assert.Equal(t, uint32(20116724), tick)

}

func TestAssets_AssetIssuances_UnmarshallFromReader(t *testing.T) {

	// RESPOND_ASSETS header + payload + RESPOND_ASSETS header + payload + END_RESPONSE header
	hexStr := "400000357ee06e4a0000000000000000000000000000000000000000000000000000000000000000014d4c4d0000000000000000000000002f18340100000000" +
		"400000357ee06e4a000000000000000000000000000000000000000000000000000000000000000001515641554c540000000000000000002f18340101000000" +
		"080000237ee06e4a"

	issuedAssetsBin, err := hex.DecodeString(hexStr)
	require.NoError(t, err)

	var result AssetIssuances
	err = result.UnmarshallFromReader(bytes.NewReader(issuedAssetsBin))
	require.NoError(t, err)

	assert.Len(t, result, 2)
	mlm := result[0]    // MLM
	qvault := result[1] // QVAULT

	assert.Equal(t, int8(0), mlm.Asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{77, 76, 77, 0, 0, 0, 0}, mlm.Asset.Name)
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, mlm.Asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, mlm.Asset.PublicKey)
	assert.Equal(t, uint8(1), mlm.Asset.Type)
	index := mlm.UniverseIndex
	assert.Equal(t, uint32(0), index)
	tick := mlm.Tick
	assert.Equal(t, uint32(20191279), tick)

	assert.Equal(t, int8(0), qvault.Asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{81, 86, 65, 85, 76, 84, 0}, qvault.Asset.Name)
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, qvault.Asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, mlm.Asset.PublicKey)
	assert.Equal(t, uint8(1), qvault.Asset.Type)
	index = qvault.UniverseIndex
	assert.Equal(t, uint32(1), index)
	tick = qvault.Tick
	assert.Equal(t, uint32(20191279), tick)
}

func TestAssets_AssetOwnerships_UnmarshallFromReader(t *testing.T) {

	// RESPOND_ASSETS header + payload + RESPOND_ASSETS header + payload + END_RESPONSE header
	hexStr := "4000003511dcb7b97b5efffa039860590ecc801ab2f9a95da0b97592398d3414db1d3e44cac79d9a020001000400000004000000000000005b1c34017b5eff00" +
		"4000003511dcb7b9feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10020001000400000001000000000000005b1c3401feb0fb00" +
		"0800002328af10a4"

	ownedAssetsBin, err := hex.DecodeString(hexStr)
	require.NoError(t, err)

	var result AssetOwnerships
	err = result.UnmarshallFromReader(bytes.NewReader(ownedAssetsBin))
	require.NoError(t, err)

	assert.Len(t, result, 2)
	owner1 := result[0]
	owner2 := result[1]

	asset := owner1.Asset
	publicKey, _ := hex.DecodeString("7b5efffa039860590ecc801ab2f9a95da0b97592398d3414db1d3e44cac79d9a")
	assert.Equal(t, [32]uint8(publicKey), asset.PublicKey)
	assert.Equal(t, uint8(2), asset.Type)
	assert.Equal(t, uint16(1), asset.ManagingContractIndex)
	assert.Equal(t, uint32(4), asset.IssuanceIndex)
	assert.Equal(t, int64(4), asset.NumberOfUnits)

	index := owner1.UniverseIndex
	assert.Equal(t, uint32(16735867), index)
	tick := owner1.Tick
	assert.Equal(t, uint32(20192347), tick)

	asset = owner2.Asset
	publicKey, _ = hex.DecodeString("feb0fb0e023c5f98ae9549112117ef3bf80608fcd252abc5772a07efd3f88b10")
	assert.Equal(t, [32]uint8(publicKey), asset.PublicKey)
	assert.Equal(t, uint8(2), asset.Type)
	assert.Equal(t, uint16(1), asset.ManagingContractIndex)
	assert.Equal(t, uint32(4), asset.IssuanceIndex)
	assert.Equal(t, int64(1), asset.NumberOfUnits)

	index = owner2.UniverseIndex
	assert.Equal(t, uint32(16494846), index)
	tick = owner2.Tick
	assert.Equal(t, uint32(20192347), tick)
}
