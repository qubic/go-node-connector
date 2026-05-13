package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Ipo struct {
	ContractIndex uint32
	AssetName     [8]byte
}

func (ipo *Ipo) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ipo.ContractIndex)
	if err != nil {
		return fmt.Errorf("reading contract index from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ipo.AssetName)
	if err != nil {
		return fmt.Errorf("reading asset name from reader: %w", err)
	}

	return nil

}

type Ipos []Ipo

func (ipos *Ipos) UnmarshallFromReader(r io.Reader) error {

	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != ActiveIpoResponse {
			return fmt.Errorf("invalid header type, expected %d, found %d", ActiveIpoResponse, header.Type)
		}

		var ipo Ipo
		err = ipo.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling active ipo: %w", err)
		}
		*ipos = append(*ipos, ipo)
	}

	return nil

}

type ContractIpo struct {
	ContractIndex uint32
	TickNumber    uint32
	PubKeys       [NumberOfComputors][32]byte
	Prices        [NumberOfComputors]int64
}

func (ci *ContractIpo) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	if header.Type != ContractIpoResponse {
		return fmt.Errorf("Invalid header type, expected %d, found %d", ContractIpoResponse, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, ci)
	if err != nil {
		return fmt.Errorf("reading contract ipo data from reader: %w", err)
	}

	return nil
}
