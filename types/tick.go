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

	err = binary.Read(r, binary.LittleEndian, &td.ComputorIndex)
	if err != nil {
		return errors.Wrap(err, "reading computor index")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Epoch)
	if err != nil {
		return errors.Wrap(err, "reading epoch")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Tick)
	if err != nil {
		return errors.Wrap(err, "reading tick")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Millisecond)
	if err != nil {
		return errors.Wrap(err, "reading millisecond")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Second)
	if err != nil {
		return errors.Wrap(err, "reading second")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Minute)
	if err != nil {
		return errors.Wrap(err, "reading minute")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Hour)
	if err != nil {
		return errors.Wrap(err, "reading hour")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Day)
	if err != nil {
		return errors.Wrap(err, "reading day")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Month)
	if err != nil {
		return errors.Wrap(err, "reading month")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Year)
	if err != nil {
		return errors.Wrap(err, "reading year")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Timelock)
	if err != nil {
		return errors.Wrap(err, "reading timelock")
	}

	err = binary.Read(r, binary.LittleEndian, &td.TransactionDigests)
	if err != nil {
		return errors.Wrap(err, "reading transaction digests")
	}

	err = binary.Read(r, binary.LittleEndian, &td.ContractFees)
	if err != nil {
		return errors.Wrap(err, "reading contract fees")
	}

	err = binary.Read(r, binary.LittleEndian, &td.Signature)
	if err != nil {
		return errors.Wrap(err, "reading signature")
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
	InitialTick             uint32
}

func (ti *TickInfo) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader

		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return errors.Wrap(err, "reading header")
		}

		if header.Type == 0 {
			ignoredBytes := make([]byte, header.GetSize()-uint32(binary.Size(header)))
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
