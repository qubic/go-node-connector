package types

import (
	"bytes"
	"strings"
	"testing"
)

// sizedHeader builds an 8-byte RequestResponseHeader wire image: Size[3], Type, DejaVu[4].
func sizedHeader(size uint32, typ uint8) []byte {
	return []byte{
		byte(size), byte(size >> 8), byte(size >> 16), // Size[3]
		typ,        // Type
		0, 0, 0, 0, // DejaVu
	}
}

// Regression for the unsigned underflow in the UnmarshallFromReader size handling.
//
// GetSize() is a node-controlled 24-bit value. When a node returns a header whose
// Size field is below sizeof(RequestResponseHeader) (8), GetSize()-uint32(headerSize)
// underflows to ~4.29e9 and make([]byte, ...) attempts a ~4 GB allocation, OOM-crashing
// any service that embeds this client. These paths are reached straight off the raw
// connection via sendRequest -> readPacketIntoDest -> UnmarshallFromReader. Each
// unmarshaler must reject the undersized header instead of allocating.
func TestUnmarshalRejectsHeaderSmallerThanHeaderSize(t *testing.T) {
	const tooSmall = uint32(4) // 0 < size < 8 -> GetSize()-8 underflows

	t.Run("Computors", func(t *testing.T) {
		var c Computors
		// Any Type != BroadcastComputors reaches the skip-branch make.
		err := c.UnmarshallFromReader(bytes.NewReader(sizedHeader(tooSmall, 99)))
		requireInvalidSize(t, err)
	})

	t.Run("TickInfo", func(t *testing.T) {
		var ti TickInfo
		// Type == 0 reaches the zero-padding skip-branch make.
		err := ti.UnmarshallFromReader(bytes.NewReader(sizedHeader(tooSmall, 0)))
		requireInvalidSize(t, err)
	})

	t.Run("SmartContractData", func(t *testing.T) {
		var scd SmartContractData
		err := scd.UnmarshallFromReader(bytes.NewReader(sizedHeader(tooSmall, ContractFunctionResponse)))
		requireInvalidSize(t, err)
	})

	t.Run("PublicPeers", func(t *testing.T) {
		var pp PublicPeers
		buf := sizedHeader(100, ExchangePublicPeers)   // valid first header
		buf = append(buf, make([]byte, 16)...)         // [4][4]byte peers payload
		buf = append(buf, sizedHeader(tooSmall, 0)...) // undersized nextHeader
		err := pp.UnmarshallFromReader(bytes.NewReader(buf))
		requireInvalidSize(t, err)
	})
}

func requireInvalidSize(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error for an undersized header, got nil (an unguarded make() would request ~4GB)")
	}
	if !strings.Contains(err.Error(), "invalid packet size") {
		t.Fatalf("expected an invalid-packet-size error, got: %v", err)
	}
}
