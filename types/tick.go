package types

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
)

const (
	NumberOfTransactionsPerTick = 1024
)

type TickData struct {
	ComputorIndex      uint16
	Epoch              uint16
	Tick               uint32
	Millisecond        uint16
	Second             uint8
	Minute             uint8
	Hour               uint8
	Day                uint8
	Month              uint8
	Year               uint8
	UnionData          [256]byte
	Timelock           [32]byte
	TransactionDigests [NumberOfTransactionsPerTick][32]byte `json:",omitempty"`
	ContractFees       [1024]int64                           `json:",omitempty"`
	Signature          [SignatureSize]byte
}

func (td *TickData) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return errors.Wrap(err, "reading tick data from reader")
	}

	if header.Type == EndResponse {
		return nil
	}

	if header.Type != BroadcastFutureTickData {
		return errors.Errorf("Invalid header type, expected %d, found %d", BroadcastFutureTickData, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, td)
	if err != nil {
		return errors.Wrap(err, "reading tick data from reader")
	}

	return nil
}

func (td *TickData) IsEmpty() bool {
	if td == nil {
		return true
	}

	return *td == TickData{}
}

type TickInfo struct {
	TickDuration            uint16
	Epoch                   uint16
	Tick                    uint32
	NumberOfAlignedVotes    uint16
	NumberOfMisalignedVotes uint16
	InitialTick uint32
}

func (ti *TickInfo) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader

		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return errors.Wrap(err, "reading header")
		}

		if header.Type == 0 {
			ignoredBytes := make([]byte, header.GetSize() - uint32(binary.Size(header)))
			_, err := r.Read(ignoredBytes)
			if err != nil {
				return errors.Wrap(err, "reading ignored bytes")
			}
			continue
		}

		if header.Type != CurrentTickInfoResponse {
			return errors.Errorf("Invalid header type, expected %d, found %d", CurrentTickInfoResponse, header.Type)
		}

		err = binary.Read(r, binary.LittleEndian, ti)
		if err != nil {
			return errors.Wrap(err, "reading tick data from reader")
		}

		break
	}


	return nil
}
