package tcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/0xluk/go-qubic/data/tick"
	"github.com/pkg/errors"
	"reflect"
)

type SenderReceiver interface {
	SendHeaderData(ctx context.Context, data RequestResponseHeader) error
	SendRequestData(ctx context.Context, data interface{}) error
	ReceiveDataAll() ([]byte, error)
}

func SendTransaction(ctx context.Context, qc SenderReceiver, requestType uint8, responseType uint8, requestData interface{}, dest interface{}) error {
	err := sendTxReq(ctx, qc, requestType, requestData)
	if err != nil {
		return errors.Wrap(err, "sending request")
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = readResponse(ctx, qc, responseType, dest)
	if err != nil {
		return errors.Wrap(err, "reading response")
	}

	return nil
}

func SendGenericRequest(ctx context.Context, qc SenderReceiver, requestType uint8, responseType uint8, requestData interface{}, dest interface{}) error {
	err := sendReq(ctx, qc, requestType, requestData)
	if err != nil {
		return errors.Wrap(err, "sending request")
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = readResponse(ctx, qc, responseType, dest)
	if err != nil {
		return errors.Wrap(err, "reading response")
	}

	return nil
}

func SendGetTransactionsRequest(ctx context.Context, qc SenderReceiver, requestType uint8, responseType uint8, requestData interface{}, nrTx int) ([]tick.Transaction, error) {
	err := sendReq(ctx, qc, requestType, requestData)
	if err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	// Receive and process response
	buffer, err := qc.ReceiveDataAll()
	if err != nil {
		return nil, errors.Wrap(err, "receiving response")
	}

	data := buffer[:]
	ptr := 0
	txs := make([]tick.Transaction, 0, nrTx)

	for ptr < len(data) {
		var header RequestResponseHeader
		headerSize := binary.Size(header)

		if len(data)-ptr < headerSize {
			// Not enough data for the header, break the loop
			break
		}

		err := binary.Read(bytes.NewReader(data[ptr:ptr+headerSize]), binary.BigEndian, &header)
		if err != nil {
			return nil, errors.Wrap(err, "reading header data")
		}

		if header.Type != responseType {
			ptr += int(header.GetSize())
			continue
		}
		var tx tick.Transaction
		destSize := binary.Size(&tx)

		if len(data)-ptr-headerSize < destSize {
			return nil, errors.Errorf("Not enough data for the tx. Got: %d, expected %d", len(data)-ptr-headerSize, destSize)
			// Not enough data for the tx, break the loop
		}
		currentData := data[ptr+headerSize:]
		err = binary.Read(bytes.NewReader(currentData), binary.LittleEndian, &tx)
		if err != nil {
			return nil, errors.Wrapf(err, "reading response data:%s", currentData)
		}

		txs = append(txs, tx)
		ptr += int(header.GetSize())
	}

	return txs, nil
}

func sendReq(ctx context.Context, qc SenderReceiver, requestType uint8, requestData interface{}) error {
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

	return nil
}

func sendTxReq(ctx context.Context, qc SenderReceiver, requestType uint8, requestData interface{}) error {
	packet := struct {
		Header      RequestResponseHeader
		RequestData interface{}
	}{
		RequestData: requestData,
	}
	size := binary.Size(packet.Header) + getSizeOfRequestData(requestData)
	packet.Header.SetSize(uint32(size))
	packet.Header.ZeroDejaVu()
	packet.Header.Type = requestType

	err := qc.SendHeaderData(ctx, packet.Header)
	if err != nil {
		return errors.Wrap(err, "sending header data to conn")
	}

	err = qc.SendRequestData(ctx, packet.RequestData)
	if err != nil {
		return errors.Wrap(err, "sending request data to conn")
	}

	return nil
}

func readResponse(ctx context.Context, qc SenderReceiver, responseType uint8, dest interface{}) error {
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
	switch v := requestData.(type) {
	case nil:
		return 0
	case []byte:
		return len(v)
	default:
		return int(reflect.TypeOf(requestData).Size())
	}
}
