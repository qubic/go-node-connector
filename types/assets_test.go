package types

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAssets_AssetIssuances_UnmarshallFromReader(t *testing.T) {

	// RESPOND_ASSETS header (400000351cd1f262) + payload + END_RESPONSE header (080000231cd1f262)
	hexStr := "400000351cd1f26200000000000000000000000000000000000000000000000000000000000000000152414e444f4d000000000000000000f4f4320103000000080000231cd1f262"

	issuedAssetsBin, err := hex.DecodeString(hexStr)
	require.NoError(t, err)

	var result AssetIssuances
	err = result.UnmarshallFromReader(bytes.NewReader(issuedAssetsBin))
	require.NoError(t, err)

	assetIssuance := result[0].Asset
	assert.Len(t, result, 1)
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
