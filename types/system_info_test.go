package types

import (
	"bytes"
	"encoding/binary"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystemInfo_UnmarshallFromReader(t *testing.T) {

	testData := []struct {
		name               string
		expectedSystemInfo SystemInfo
	}{
		{
			name: "Test_UnmarshallFromReader1",
			expectedSystemInfo: SystemInfo{
				Version:                           200,
				Epoch:                             100,
				Tick:                              23215487,
				InitialTick:                       20000000,
				LatestCreatedTick:                 23215486,
				InitialMillisecond:                10,
				InitialSecond:                     11,
				InitialMinute:                     12,
				InitialHour:                       13,
				InitialDay:                        14,
				InitialMonth:                      15,
				InitialYear:                       16,
				NumberOfEntities:                  999999,
				NumberOfTransactions:              909090909,
				RandomMiningSeed:                  [32]byte{0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02},
				SolutionThreshold:                 412,
				TotalSpectrumAmount:               98989898989898,
				CurrentEntityBalanceDustThreshold: 200,
				TargetTickVoteSignature:           0xFFAB,
				Reserve0:                          0,
				Reserve1:                          0,
				Reserve2:                          0,
				Reserve3:                          0,
				Reserve4:                          0,
			},
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.name, func(t *testing.T) {

			var buffer bytes.Buffer

			testHeader := RequestResponseHeader{
				Type: SystemInfoResponse,
			}
			testHeader.RandomizeDejaVu()
			testHeader.SetSize(128)

			err := binary.Write(&buffer, binary.LittleEndian, testHeader)
			assert.NoError(t, err)
			err = binary.Write(&buffer, binary.LittleEndian, testCase.expectedSystemInfo)
			assert.NoError(t, err)

			var got SystemInfo
			err = got.UnmarshallFromReader(&buffer)
			assert.NoError(t, err)

			if cmp.Diff(testCase.expectedSystemInfo, got) != "" {
				t.Fatalf("Mismatched return value. Expected: %v, got: %v", testCase.expectedSystemInfo, got)
			}

		})

	}

}
