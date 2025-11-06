//go:build !ci
// +build !ci

package qubic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const UniverseIndexOfRandom = 7 // can change over time

var nodeIp = "5.9.16.14"

func TestQubicIntegration_GetAssetIssuancesByUniverseIndexPayload(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
	assert.NoError(t, err)

	issuances, err := client.GetAssetIssuancesByUniverseIndex(context.Background(), UniverseIndexOfRandom)
	assert.NoError(t, err)

	assert.Len(t, issuances, 1)
	issuance := issuances[0]
	assert.Equal(t, UniverseIndexOfRandom, int(issuance.UniverseIndex))
	assert.Greater(t, issuance.Tick, uint32(20173192)) // ep 150 start
	asset := issuance.Asset
	assert.Equal(t, int8(0), asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{82, 65, 78, 68, 79, 77, 0}, asset.Name) // RANDOM
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, asset.PublicKey)
	assert.Equal(t, uint8(1), asset.Type)
}

func TestQubicIntegration_GetAssetOwnershipsByUniverseIndexPayload(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
	assert.NoError(t, err)

	ownerships, err := client.GetAssetOwnershipsByUniverseIndex(context.Background(), 16697282)
	assert.NoError(t, err)
	assert.Len(t, ownerships, 1)

	ownership := ownerships[0]
	assert.Equal(t, 16697282, int(ownership.UniverseIndex))
	assert.Positive(t, ownership.UniverseIndex)
	assert.Greater(t, ownership.Tick, uint32(20173192)) // ep 150 start

	asset := ownership.Asset
	assert.Equal(t, asset.ManagingContractIndex, uint16(1))          // 1 = QX
	assert.Equal(t, byte(2), asset.Type)                             // ownership type
	assert.Equal(t, UniverseIndexOfRandom, int(asset.IssuanceIndex)) // RANDOM
	// owner information
	assert.Positive(t, asset.NumberOfUnits)
	assert.Len(t, asset.PublicKey, 32)
}

func TestQubicIntegration_GetPossessionsByUniverseIndexPayload(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
	assert.NoError(t, err)

	possessions, err := client.GetAssetPossessionsByUniverseIndex(context.Background(), 16697283)
	assert.NoError(t, err)
	assert.Len(t, possessions, 1)

	possession := possessions[0] // take first
	assert.Positive(t, possession.UniverseIndex)
	assert.Equal(t, 16697283, int(possession.UniverseIndex))
	assert.Greater(t, possession.Tick, uint32(20173192)) // ep 150 start

	asset := possession.Asset
	assert.Equal(t, asset.ManagingContractIndex, uint16(1)) // 1 = QX
	assert.Equal(t, byte(3), asset.Type)                    // possession type
	assert.Positive(t, asset.OwnershipIndex)                // RANDOM
	// owner information
	assert.GreaterOrEqual(t, asset.NumberOfUnits, int64(1))
	assert.Len(t, asset.PublicKey, 32)
}

func TestQubicIntegration_GetAssetIssuancesByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
	assert.NoError(t, err)

	issuances, err := client.GetAssetIssuancesByFilter(context.Background(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB", "RANDOM")
	assert.NoError(t, err)

	assert.Len(t, issuances, 1)

	issuance := issuances[0]
	assert.Equal(t, 7, int(issuance.UniverseIndex))    // universe index can change over time
	assert.Greater(t, issuance.Tick, uint32(20173192)) // ep 150 start
	asset := issuance.Asset
	assert.Equal(t, int8(0), asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{82, 65, 78, 68, 79, 77, 0}, asset.Name) // RANDOM
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, asset.PublicKey)
	assert.Equal(t, uint8(1), asset.Type) // issuance type
}

func TestQubicIntegration_GetAssetOwnershipsByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
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
	//assert.Equal(t, 16697282, int(ownership.UniverseIndex))
	assert.Greater(t, ownership.Tick, uint32(20173192)) // ep 150 start

	asset := ownership.Asset
	assert.Equal(t, asset.ManagingContractIndex, uint16(1))          // 1 = QX
	assert.Equal(t, byte(2), asset.Type)                             // ownership type
	assert.Equal(t, UniverseIndexOfRandom, int(asset.IssuanceIndex)) // RANDOM
	// owner information
	assert.Positive(t, asset.NumberOfUnits)
	assert.Len(t, asset.PublicKey, 32)
}

func TestQubicIntegration_GetAssetPossessionsByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
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
	//assert.Equal(t, 16697283, int(possession.UniverseIndex))
	assert.Greater(t, possession.Tick, uint32(20173192)) // ep 150 start

	asset := possession.Asset
	assert.Equal(t, asset.ManagingContractIndex, uint16(1)) // 1 = QX
	assert.Equal(t, byte(3), asset.Type)                    // possession type
	assert.Positive(t, asset.OwnershipIndex)                // RANDOM
	// owner information
	assert.GreaterOrEqual(t, asset.NumberOfUnits, int64(1))
	assert.Len(t, asset.PublicKey, 32)
}
