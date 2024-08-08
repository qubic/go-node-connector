package types

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
)

type SmartContractData struct {
	Data []byte
}

func (scd *SmartContractData) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return errors.Wrap(err, "reading header")
	}

	if header.Type == EndResponse {
		return nil
	}

	if header.Type != ContractFunctionResponse {
		return errors.Errorf("Invalid header type, expected %d, found %d", ContractFunctionResponse, header.Type)
	}

	data := make([]byte, header.GetSize()-uint32(binary.Size(header)))

	err = binary.Read(r, binary.LittleEndian, data)
	if err != nil {
		return errors.Wrap(err, "reading data")
	}

	scd.Data = data

	return nil
}
