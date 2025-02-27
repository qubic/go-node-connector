//go:build !ci
// +build !ci

package qubic

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQubicIntegration_GetAssetByUniverseIndexPayload(t *testing.T) {
	client, err := NewClient(context.Background(), "176.223.119.131", "21841")
	assert.NoError(t, err)

	issuances, err := client.GetAssetsByUniverseIndex(context.Background(), 3)
	assert.NoError(t, err)

	assert.Len(t, issuances, 1)
	issuance := issuances[0]
	assert.Equal(t, uint32(3), issuance.UniverseIndex)
	assert.Greater(t, issuance.Tick, uint32(20173192)) // ep 150 start
	asset := issuance.Asset
	assert.Equal(t, int8(0), asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{82, 65, 78, 68, 79, 77, 0}, asset.Name) // RANDOM
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, asset.PublicKey)
	assert.Equal(t, uint8(1), asset.Type)

}

func TestQubicIntegration_GetAssetIssuancesByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), "176.223.119.131", "21841")
	assert.NoError(t, err)

	issuances, err := client.GetAssetIssuancesByFilter(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "RANDOM")
	assert.NoError(t, err)

	assert.Len(t, issuances, 1)

	issuance := issuances[0]
	assert.Equal(t, uint32(3), issuance.UniverseIndex)
	assert.Greater(t, issuance.Tick, uint32(20173192)) // ep 150 start
	asset := issuance.Asset
	assert.Equal(t, int8(0), asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{82, 65, 78, 68, 79, 77, 0}, asset.Name) // RANDOM
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, asset.PublicKey)
	assert.Equal(t, uint8(1), asset.Type) // issuance type
}

func TestQubicIntegration_GetAssetOwnershipsByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), "176.223.119.131", "21841")
	assert.NoError(t, err)

	ownerships, err := client.GetAssetOwnershipsByFilter(context.Background(),
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB",
		"RANDOM",
		"",
		1)
	assert.NoError(t, err)
	assert.NotEmpty(t, ownerships)

	ownership := ownerships[0]
	assert.Positive(t, ownership.UniverseIndex)
	assert.Greater(t, ownership.Tick, uint32(20173192)) // ep 150 start

	asset := ownership.Asset
	assert.Equal(t, asset.ManagingContractIndex, uint16(1)) // 1 = QX
	assert.Equal(t, byte(2), asset.Type)                    // ownership type
	assert.Equal(t, uint32(3), asset.IssuanceIndex)         // RANDOM
	// owner information
	assert.Positive(t, asset.NumberOfUnits)
	assert.Len(t, asset.PublicKey, 32)
}

func TestQubicIntegration_GetAssetPossessionsByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), "176.223.119.131", "21841")
	assert.NoError(t, err)

	possessions, err := client.GetAssetPossessionsByFilter(context.Background(),
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB",
		"RANDOM",
		"",
		"",
		1,
		1)
	assert.NoError(t, err)
	assert.NotEmpty(t, possessions)

	possession := possessions[0] // take first
	assert.Positive(t, possession.UniverseIndex)
	assert.Greater(t, possession.Tick, uint32(20173192)) // ep 150 start

	asset := possession.Asset
	assert.Equal(t, asset.ManagingContractIndex, uint16(1)) // 1 = QX
	assert.Equal(t, byte(3), asset.Type)                    // possession type
	assert.Positive(t, asset.OwnershipIndex)                // RANDOM
	// owner information
	assert.GreaterOrEqual(t, asset.NumberOfUnits, int64(1))
	assert.Len(t, asset.PublicKey, 32)
}
