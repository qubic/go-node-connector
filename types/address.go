package types

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
)

const (
	SpectrumDepth       = 24
)

type AddressData struct {
	PublicKey                  [32]byte
	IncomingAmount             int64
	OutgoingAmount             int64
	NumberOfIncomingTransfers  uint32
	NumberOfOutgoingTransfers  uint32
	LatestIncomingTransferTick uint32
	LatestOutgoingTransferTick uint32
}

type AddressInfo struct {
	AddressData   AddressData
	Tick          uint32
	SpectrumIndex int32
	Siblings      [SpectrumDepth][32]byte
}

func (ai *AddressInfo) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return errors.Wrap(err, "reading header")
	}

	if header.Type != BalanceTypeResponse {
		return errors.Errorf("Invalid header type, expected %d, found %d", BalanceTypeResponse, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, ai)
	if err != nil {
		return errors.Wrap(err, "reading addr info data from reader")
	}

	return nil
}
