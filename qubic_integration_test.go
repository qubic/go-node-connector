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

	issuance := issuances[0]
	assert.Equal(t, uint32(3), issuance.UniverseIndex)
	assert.Greater(t, issuance.Tick, uint32(20168560))

	asset := issuance.Asset
	assert.Equal(t, int8(0), asset.NumberOfDecimalPlaces)
	assert.Equal(t, [7]int8{82, 65, 78, 68, 79, 77, 0}, asset.Name) // RANDOM
	assert.Equal(t, [7]int8{0, 0, 0, 0, 0, 0, 0}, asset.UnitOfMeasurement)
	assert.Equal(t, [32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, asset.PublicKey)
	assert.Equal(t, uint8(1), asset.Type)

	assert.Len(t, issuances, 1)
}
