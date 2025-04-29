//go:build !ci
// +build !ci

package qubic

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var nodeIp = "45.152.160.18"

func TestQubicIntegration_GetAssetIssuancesByUniverseIndexPayload(t *testing.T) {
	client, err := NewClient(context.Background(), nodeIp, "21841")
	assert.NoError(t, err)

	issuances, err := client.GetAssetIssuancesByUniverseIndex(context.Background(), 3)
	assert.NoError(t, err)

	assert.Len(t, issuances, 1)
	issuance := issuances[0]
	assert.Equal(t, 3, int(issuance.UniverseIndex))
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
	assert.Equal(t, asset.ManagingContractIndex, uint16(1)) // 1 = QX
	assert.Equal(t, byte(2), asset.Type)                    // ownership type
	assert.Equal(t, uint32(3), asset.IssuanceIndex)         // RANDOM
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
	assert.Equal(t, asset.ManagingContractIndex, uint16(1)) // 1 = QX
	assert.Equal(t, byte(2), asset.Type)                    // ownership type
	assert.Equal(t, uint32(3), asset.IssuanceIndex)         // RANDOM
	// owner information
	assert.Positive(t, asset.NumberOfUnits)
	assert.Len(t, asset.PublicKey, 32)
}

func TestQubicIntegration_GetAssetPossessionsByFilter(t *testing.T) {
	client, err := NewClient(context.Background(), "91.210.226.50", "31841")
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

/*func TestClient_GetSystemInfo(t *testing.T) {
	client, err := NewClient(context.Background(), "", "31841")
	assert.NoError(t, err)

	systemInfo, err := client.GetSystemInfo(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, systemInfo)

	fmt.Printf("Version: %d\n", systemInfo.Version)
	fmt.Printf("Epoch: %d\n", systemInfo.Epoch)

	fmt.Printf("Tick: %d\n", systemInfo.Tick)
	fmt.Printf("Initial tick: %d\n", systemInfo.InitialTick)
	fmt.Printf("Latest created tick: %d\n", systemInfo.LatestCreatedTick)

	fmt.Printf("Number of entities: %d\n", systemInfo.NumberOfEntities)
	fmt.Printf("Number of transactions: %d\n", systemInfo.NumberOfTransactions)

	fmt.Printf("Solution threshold: %d\n", systemInfo.SolutionThreshold)

	fmt.Printf("Total spectrum amount: %d\n", systemInfo.TotalSpectrumAmount)

	fmt.Printf("Random mining seed: %x\n", systemInfo.RandomMiningSeed)

	fmt.Printf("Current entity balance dust threshold: %d\n", systemInfo.CurrentEntityBalanceDustThreshold)

	fmt.Printf("Target tick vote signature: %X\n", systemInfo.TargetTickVoteSignature)

}
*/
