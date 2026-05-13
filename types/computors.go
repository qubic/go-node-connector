package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	SignatureSize     = 64
	NumberOfComputors = 676
)

type Computors struct {
	Epoch     uint16
	PubKeys   [NumberOfComputors][32]byte
	Signature [SignatureSize]byte
}

func (cs *Computors) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		headerSize := binary.Size(header)
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type != BroadcastComputors {
			ignoredbytes := make([]byte, header.GetSize()-uint32(headerSize))
			_, err := r.Read(ignoredbytes)
			if err != nil {
				return fmt.Errorf("reading ignored bytes: %w", err)
			}
			continue
		}

		err = binary.Read(r, binary.LittleEndian, cs)
		if err != nil {
			return fmt.Errorf("reading computors from reader: %w", err)
		}

		return nil
	}
}
