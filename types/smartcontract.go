package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

type SmartContractData struct {
	Data []byte
}

func (scd *SmartContractData) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	if header.Type == EndResponse {
		return nil
	}

	if header.Type != ContractFunctionResponse {
		return fmt.Errorf("Invalid header type, expected %d, found %d", ContractFunctionResponse, header.Type)
	}

	if header.GetSize() < uint32(binary.Size(header)) {
		return fmt.Errorf("invalid packet size %d from node", header.GetSize())
	}
	data := make([]byte, header.GetSize()-uint32(binary.Size(header)))

	err = binary.Read(r, binary.LittleEndian, data)
	if err != nil {
		return fmt.Errorf("reading data: %w", err)
	}

	scd.Data = data

	return nil
}
