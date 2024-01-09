package tick

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
)

const (
	NUMBER_OF_TRANSACTIONS_PER_TICK = 1024
	SIGNATURE_SIZE                  = 64
	REQUEST_CURRENT_TICK_INFO       = 27
	RESPOND_CURRENT_TICK_INFO       = 28
	BROADCAST_FUTURE_TICK_DATA      = 8
	REQUEST_TICK_DATA               = 16
	REQUEST_TICK_TRANSACTIONS       = 29
	BROADCAST_TRANSACTION           = 24
)

type TickData struct {
	ComputorIndex uint16
	Epoch         uint16
	Tick          uint32
	Millisecond   uint16
	Second        uint8
	Minute        uint8
	Hour          uint8
	Day           uint8
	Month         uint8
	Year          uint8
	UnionData     [256]byte
	//VarStruct     struct {
	//	Proposal struct {
	//		URISize uint8
	//		URI     [255]byte
	//	}
	//	Ballot struct {
	//		Zero              uint8
	//		Votes             [(676*3 + 7) / 8]byte // Adjusted for padding
	//		QuasiRandomNumber uint8
	//	}
	//}
	Timelock           [32]byte
	TransactionDigests [NUMBER_OF_TRANSACTIONS_PER_TICK][32]byte `json:",omitempty"`
	ContractFees       [1024]int64                               `json:",omitempty"`
	Signature          [SIGNATURE_SIZE]byte
}

type CurrentTickInfo struct {
	TickDuration            uint16
	Epoch                   uint16
	Tick                    uint32
	NumberOfAlignedVotes    uint16
	NumberOfMisalignedVotes uint16
}

type RequestTickData struct {
	Tick uint32
}

type TransactionHeader struct {
	SourcePublicKey      [32]byte
	DestinationPublicKey [32]byte
	Amount               int64
	Tick                 uint32
	InputType            uint16
	InputSize            uint16
}

func (th *TransactionHeader) MarshallBinary() ([]byte, error) {
	var buff bytes.Buffer
	err := binary.Write(&buff, binary.LittleEndian, th.SourcePublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "writing source pubkey to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.DestinationPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "writing dest pubkey to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.Amount)
	if err != nil {
		return nil, errors.Wrap(err, "writing amount to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.Tick)
	if err != nil {
		return nil, errors.Wrap(err, "writing tick to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.InputType)
	if err != nil {
		return nil, errors.Wrap(err, "writing input type to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.InputSize)
	if err != nil {
		return nil, errors.Wrap(err, "writing input size to buf")
	}

	return buff.Bytes(), nil
}

type TransactionData struct {
	Header    TransactionHeader
	Input     []byte
	Signature [64]byte
}

func (td *TransactionData) MarshallBinary() ([]byte, error) {
	var out []byte
	binaryHeader, err := td.Header.MarshallBinary()
	if err != nil {
		return nil, errors.Wrap(err, "writing txData to buf")
	}

	out = append(out, binaryHeader...)
	out = append(out, td.Input...)
	out = append(out, td.Signature[:]...)

	return out, nil
}

type TransactionInput []byte

type TransactionSig [65]byte

type TransactionHash [60]byte

type Transaction struct {
	Data TransactionData
	Hash TransactionHash
}

type RequestTickTransactions struct {
	Tick             uint32
	TransactionFlags [NUMBER_OF_TRANSACTIONS_PER_TICK / 8]uint8
}
