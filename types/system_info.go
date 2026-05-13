package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

type SystemInfo struct {
	Version int16

	Epoch             uint16
	Tick              uint32
	InitialTick       uint32
	LatestCreatedTick uint32

	InitialMillisecond uint16
	InitialSecond      uint8
	InitialMinute      uint8
	InitialHour        uint8
	InitialDay         uint8
	InitialMonth       uint8
	InitialYear        uint8

	NumberOfEntities     uint32
	NumberOfTransactions uint32

	RandomMiningSeed  [32]byte
	SolutionThreshold int32

	TotalSpectrumAmount uint64

	CurrentEntityBalanceDustThreshold uint64

	TargetTickVoteSignature uint32
	ComputorPacketSignature uint64

	Reserve1 uint64
	Reserve2 uint64
	Reserve3 uint64
	Reserve4 uint64
}

func (si *SystemInfo) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return fmt.Errorf("reading system info response header: %w", err)
	}

	if header.Type == EndResponse {
		return nil
	}

	if header.Type != SystemInfoResponse {
		return fmt.Errorf("invalid header type. expected %d, found %d", SystemInfoResponse, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, si)
	if err != nil {
		return fmt.Errorf("reading system information response: %w", err)
	}
	return nil
}
