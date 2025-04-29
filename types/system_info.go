package types

import (
	"encoding/binary"
	"github.com/pkg/errors"
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

	reserve0 uint64
	reserve1 uint64
	reserve2 uint64
	reserve3 uint64
	reserve4 uint64
}

func (si *SystemInfo) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return errors.Wrap(err, "reading system info response header")
	}

	if header.Type == EndResponse {
		return nil
	}

	if header.Type != SystemInfoResponse {
		return errors.Wrapf(err, "invalid header type. expected %d, found %d", SystemInfoResponse, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, &si.Version)
	if err != nil {
		return errors.Wrap(err, "reading system version")
	}

	err = binary.Read(r, binary.LittleEndian, &si.Epoch)
	if err != nil {
		return errors.Wrap(err, "reading system epoch")
	}

	err = binary.Read(r, binary.LittleEndian, &si.Tick)
	if err != nil {
		return errors.Wrap(err, "reading system tick")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialTick)
	if err != nil {
		return errors.Wrap(err, "reading system initial tick")
	}

	err = binary.Read(r, binary.LittleEndian, &si.LatestCreatedTick)
	if err != nil {
		return errors.Wrap(err, "reading system latest created tick")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialMillisecond)
	if err != nil {
		return errors.Wrap(err, "reading system initial millisecond")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialSecond)
	if err != nil {
		return errors.Wrap(err, "reading system initial second")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialMinute)
	if err != nil {
		return errors.Wrap(err, "reading system initial minute")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialHour)
	if err != nil {
		return errors.Wrap(err, "reading system initial hour")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialDay)
	if err != nil {
		return errors.Wrap(err, "reading system initial day")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialMonth)
	if err != nil {
		return errors.Wrap(err, "reading system initial month")
	}

	err = binary.Read(r, binary.LittleEndian, &si.InitialYear)
	if err != nil {
		return errors.Wrap(err, "reading system initial year")
	}

	err = binary.Read(r, binary.LittleEndian, &si.NumberOfEntities)
	if err != nil {
		return errors.Wrap(err, "reading system number of entities")
	}

	err = binary.Read(r, binary.LittleEndian, &si.NumberOfTransactions)
	if err != nil {
		return errors.Wrap(err, "reading system number of transactions")
	}

	err = binary.Read(r, binary.LittleEndian, &si.RandomMiningSeed)
	if err != nil {
		return errors.Wrap(err, "reading system random mining seed")
	}

	err = binary.Read(r, binary.LittleEndian, &si.SolutionThreshold)
	if err != nil {
		return errors.Wrap(err, "reading system solution threshold")
	}

	err = binary.Read(r, binary.LittleEndian, &si.TotalSpectrumAmount)
	if err != nil {
		return errors.Wrap(err, "reading system total spectrum amount")
	}

	err = binary.Read(r, binary.LittleEndian, &si.CurrentEntityBalanceDustThreshold)
	if err != nil {
		return errors.Wrap(err, "reading system current entity balance dust threshold")
	}

	err = binary.Read(r, binary.LittleEndian, &si.TargetTickVoteSignature)
	if err != nil {
		return errors.Wrap(err, "reading system target tick vote signature")
	}

	err = binary.Read(r, binary.LittleEndian, &si.reserve0)
	if err != nil {
		return errors.Wrap(err, "reading system info packet reserve 0")
	}

	err = binary.Read(r, binary.LittleEndian, &si.reserve1)
	if err != nil {
		return errors.Wrap(err, "reading system info packet reserve 1")
	}

	err = binary.Read(r, binary.LittleEndian, &si.reserve2)
	if err != nil {
		return errors.Wrap(err, "reading system info packet reserve 2")
	}

	err = binary.Read(r, binary.LittleEndian, &si.reserve3)
	if err != nil {
		return errors.Wrap(err, "reading system info packet reserve 3")
	}

	err = binary.Read(r, binary.LittleEndian, &si.reserve4)
	if err != nil {
		return errors.Wrap(err, "reading system info packet reserve 4")
	}

	return nil
}
