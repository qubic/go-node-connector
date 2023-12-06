package balance

const (
	spectrumDepth      = 24
	RequestBalanceType = 31
	RespondBalanceType = 32
)

type Entity struct {
	PublicKey                  [32]byte
	IncomingAmount             int64
	OutgoingAmount             int64
	NumberOfIncomingTransfers  uint32
	NumberOfOutgoingTransfers  uint32
	LatestIncomingTransferTick uint32
	LatestOutgoingTransferTick uint32
}

type GetBalanceResponse struct {
	Entity        Entity
	Tick          uint32
	SpectrumIndex int32
	Siblings      [spectrumDepth][32]byte
}
