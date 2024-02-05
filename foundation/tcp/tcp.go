package tcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
	"reflect"
)

func SendTransaction(ctx context.Context, qc *QubicConnection, requestType uint8, responseType uint8, requestData interface{}, dest interface{}) error {
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

func SendGenericRequest(ctx context.Context, qc *QubicConnection, requestType uint8, responseType uint8, requestData interface{}, dest interface{}) error {
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

func SendGetTransactionsRequest(ctx context.Context, qc *QubicConnection, requestType uint8, responseType uint8, requestData interface{}, nrTx int) ([]types.TransactionData, error) {
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
	txs := make([]types.TransactionData, 0, nrTx)

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
		var txHeader types.TransactionHeader
		txHeaderSize := binary.Size(&txHeader)

		frameSize := len(data) - ptr - headerSize
		if frameSize < txHeaderSize {
			return nil, errors.Errorf("Not enough data for the txHeader. Got: %d, expected %d", len(data)-ptr-headerSize, txHeaderSize)
			// Not enough data for the txHeader, break the loop
		}

		offset := ptr + headerSize
		currentData := data[offset:]
		err = binary.Read(bytes.NewReader(currentData), binary.LittleEndian, &txHeader)
		if err != nil {
			return nil, errors.Wrapf(err, "reading response data:%s", currentData)
		}

		offset += txHeaderSize
		var input []byte
		if txHeader.InputSize != 0 {
			input = data[offset : offset+int(txHeader.InputSize)]
		}

		offset += int(txHeader.InputSize)
		var txSignature [64]byte
		copy(txSignature[:], data[offset:offset+64])

		txData := types.TransactionData{
			Header:    txHeader,
			Input:     input,
			Signature: txSignature,
		}
		txs = append(txs, txData)
		ptr += int(header.GetSize())
	}

	return txs, nil
}

func SendGetQuorumTickDataRequest(ctx context.Context, qc *QubicConnection, requestType uint8, responseType uint8, requestData interface{}) ([]types.QuorumTickData, error) {
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
	var quorumTicks []types.QuorumTickData

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
		var quorumTickData types.QuorumTickData
		quorumDataSize := binary.Size(&quorumTickData)

		frameSize := len(data) - ptr - headerSize
		if frameSize < quorumDataSize {
			return nil, errors.Errorf("Not enough data for the quorumTickData. Got: %d, expected %d", len(data)-ptr-headerSize, quorumDataSize)
			// Not enough data for the quorumTickData, break the loop
		}

		offset := ptr + headerSize
		currentData := data[offset:]
		err = binary.Read(bytes.NewReader(currentData), binary.LittleEndian, &quorumTickData)
		if err != nil {
			return nil, errors.Wrapf(err, "reading response data:%s", currentData)
		}

		quorumTicks = append(quorumTicks, quorumTickData)
		ptr += int(header.GetSize())
	}

	return quorumTicks, nil
}

func serializeBinary(data interface{}) ([]byte, error) {
	if data == nil {
		return nil, nil
	}

	var buff bytes.Buffer
	err := binary.Write(&buff, binary.LittleEndian, data)
	if err != nil {
		return nil, errors.Wrap(err, "writing data to buff")
	}

	return buff.Bytes(), nil
}

func sendReq(ctx context.Context, qc *QubicConnection, requestType uint8, requestData interface{}) error {
	serializedReqData, err := serializeBinary(requestData)
	if err != nil {
		return errors.Wrap(err, "serializing req data")
	}

	var header RequestResponseHeader

	packetHeaderSize := binary.Size(header)
	reqDataSize := len(serializedReqData)
	packetSize := uint32(packetHeaderSize + reqDataSize)

	header.SetSize(packetSize)
	header.RandomizeDejaVu()
	header.Type = requestType

	serializedHeaderData, err := serializeBinary(header)
	if err != nil {
		return errors.Wrap(err, "serializing header data")
	}

	serializedPacket := make([]byte, 0, packetSize)
	serializedPacket = append(serializedPacket, serializedHeaderData...)
	serializedPacket = append(serializedPacket, serializedReqData...)

	err = qc.SendRequestData(ctx, serializedPacket)
	if err != nil {
		return errors.Wrap(err, "sending request data to conn")
	}

	return nil
}

func sendTxReq(ctx context.Context, qc *QubicConnection, requestType uint8, requestData interface{}) error {
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

func readResponse(ctx context.Context, qc *QubicConnection, responseType uint8, dest interface{}) error {
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
