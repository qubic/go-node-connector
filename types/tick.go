package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	NumberOfTransactionsPerTick = 4096
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
	ContractFees       [NumberOfTransactionsPerTick]int64    `json:",omitempty"`
	Signature          [SignatureSize]byte
}

func (td *TickData) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return fmt.Errorf("reading tick data from reader: %w", err)
	}

	if header.Type == EndResponse {
		return nil
	}

	if header.Type != BroadcastFutureTickData {
		return fmt.Errorf("Invalid header type, expected %d, found %d", BroadcastFutureTickData, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, &td.ComputorIndex)
	if err != nil {
		return fmt.Errorf("reading computor index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Epoch)
	if err != nil {
		return fmt.Errorf("reading epoch: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Tick)
	if err != nil {
		return fmt.Errorf("reading tick: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Millisecond)
	if err != nil {
		return fmt.Errorf("reading millisecond: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Second)
	if err != nil {
		return fmt.Errorf("reading second: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Minute)
	if err != nil {
		return fmt.Errorf("reading minute: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Hour)
	if err != nil {
		return fmt.Errorf("reading hour: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Day)
	if err != nil {
		return fmt.Errorf("reading day: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Month)
	if err != nil {
		return fmt.Errorf("reading month: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Year)
	if err != nil {
		return fmt.Errorf("reading year: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Timelock)
	if err != nil {
		return fmt.Errorf("reading timelock: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.TransactionDigests)
	if err != nil {
		return fmt.Errorf("reading transaction digests: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.ContractFees)
	if err != nil {
		return fmt.Errorf("reading contract fees: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &td.Signature)
	if err != nil {
		return fmt.Errorf("reading signature: %w", err)
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
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == 0 {
			ignoredBytes := make([]byte, header.GetSize()-uint32(binary.Size(header)))
			_, err := r.Read(ignoredBytes)
			if err != nil {
				return fmt.Errorf("reading ignored bytes: %w", err)
			}
			continue
		}

		if header.Type != CurrentTickInfoResponse {
			return fmt.Errorf("Invalid header type, expected %d, found %d", CurrentTickInfoResponse, header.Type)
		}

		err = binary.Read(r, binary.LittleEndian, ti)
		if err != nil {
			return fmt.Errorf("reading tick data from reader: %w", err)
		}

		break
	}

	return nil
}
