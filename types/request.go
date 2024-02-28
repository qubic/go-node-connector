package types

import "math/rand"

// request and response types
const (
	CurrentTickInfoRequest  = 27
	CurrentTickInfoResponse = 28
	BroadcastFutureTickData = 8
	TickDataRequest         = 16
	TickTransactionsRequest = 29
	BroadcastTransaction    = 24
	TxStatusRequest         = 201
	TxStatusResponse        = 202
	EndResponse             = 35
	BalanceTypeRequest      = 31
	BalanceTypeResponse     = 32
	QuorumTickRequest       = 14
	QuorumTickResponse      = 3
	ComputorsRequest        = 11
	BroadcastComputors      = 2
	ExchangePublicPeers     = 0
)

type RequestResponseHeader struct {
	Size   [3]uint8
	Type   uint8
	DejaVu uint32
}

func (h *RequestResponseHeader) GetSize() uint32 {
	// Convert the array to a 32-bit unsigned integer
	size := uint32(h.Size[0]) | (uint32(h.Size[1]) << 8) | (uint32(h.Size[2]) << 16)

	// Apply the bitwise AND operation to keep the lower 24 bits
	result := size & 0xFFFFFF

	return result
}

func (h *RequestResponseHeader) SetSize(size uint32) {
	h.Size[0] = uint8(size)
	h.Size[1] = uint8(size >> 8)
	h.Size[2] = uint8(size >> 16)
}

func (h *RequestResponseHeader) IsDejaVuZero() bool {
	return h.DejaVu == 0
}

func (h *RequestResponseHeader) ZeroDejaVu() {
	h.DejaVu = 0
}

func (h *RequestResponseHeader) RandomizeDejaVu() {
	h.DejaVu = uint32(rand.Int31())
	if h.DejaVu == 0 {
		h.DejaVu = 1
	}
}
