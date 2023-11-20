package qubic

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

func getPublicKeyFromIdentity(identity string) [32]byte {
	publicKeyBuffer := make([]byte, 32)

	for i := 0; i < 4; i++ {
		value := uint64(0)
		for j := 13; j >= 0; j-- {
			if identity[i*14+j] < 'A' || identity[i*14+j] > 'Z' {
				return [32]byte{} // Error condition: invalid character in identity
			}

			value = value*26 + uint64(identity[i*14+j]-'A')
		}

		// Copy the 8-byte value into publicKeyBuffer
		for k := 0; k < 8; k++ {
			publicKeyBuffer[i*8+k] = byte(value >> (k * 8))
		}
	}

	var pubKey [32]byte
	copy(pubKey[:], publicKeyBuffer[:32])

	return pubKey
}
