package tcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/pkg/errors"
	"reflect"
)

type SenderReceiver interface {
	SendHeaderData(ctx context.Context, data RequestResponseHeader) error
	SendRequestData(ctx context.Context, data interface{}) error
	ReceiveDataAll() ([]byte, error)
}

func SendRequest(ctx context.Context, qc SenderReceiver, requestType uint8, responseType uint8, requestData interface{}, dest interface{}) error {
	packet := struct {
		Header      RequestResponseHeader
		RequestData interface{}
	}{
		RequestData: requestData,
	}
	size := binary.Size(packet.Header) + getSizeOfRequestData(requestData)
	packet.Header.SetSize(uint32(size))
	packet.Header.RandomizeDejaVu()
	packet.Header.Type = requestType

	err := qc.SendHeaderData(ctx, packet.Header)
	if err != nil {
		return errors.Wrap(err, "sending header data to conn")
	}

	err = qc.SendRequestData(ctx, packet.RequestData)
	if err != nil {
		return errors.Wrap(err, "sending request data to conn")
	}

	// Receive and process response
	buffer, err := qc.ReceiveDataAll()
	if err != nil {
		return errors.Wrap(err, "receiving response")
	}

	data := buffer[:]
	ptr := 0

	for ptr < len(data) {
		var header RequestResponseHeader
		headerSize := binary.Size(header)

		if len(data)-ptr < headerSize {
			// Not enough data for the header, break the loop
			break
		}

		err := binary.Read(bytes.NewReader(data[ptr:ptr+headerSize]), binary.BigEndian, &header)
		if err != nil {
			return errors.Wrap(err, "reading header data")
		}

		if header.Type != responseType {
			ptr += int(header.GetSize())
			continue
		}

		destSize := binary.Size(dest)

		if len(data)-ptr-headerSize < destSize {
			return errors.Errorf("Not enough data for the dest. Got: %d, expected %d", len(data)-ptr-headerSize, destSize)
			// Not enough data for the dest, break the loop
		}
		currentData := data[ptr+headerSize:]
		err = binary.Read(bytes.NewReader(currentData), binary.LittleEndian, dest)
		if err != nil {
			return errors.Wrapf(err, "reading response data:%s", currentData)
		}

		return nil
	}

	return nil
}

func getSizeOfRequestData(requestData interface{}) int {
	switch requestData.(type) {
	case nil:
		return 0
	default:
		return int(reflect.TypeOf(requestData).Size())
	}
}
