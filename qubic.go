package qubic

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
	"io"
	"net"
	"time"
)

type ReaderUnmarshaler interface {
	UnmarshallFromReader(r io.Reader) error
}

var defaultTimeout = 5 * time.Second

type Connection struct {
	conn  net.Conn
	Peers types.PublicPeers
}

func NewConnection(ctx context.Context, nodeIP, nodePort string) (*Connection, error) {
	timeout := defaultTimeout
	// Use the context deadline to calculate the timeout for net.DialTimeout
	deadline, ok := ctx.Deadline()
	if ok {
		timeout = time.Until(deadline)
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(nodeIP, nodePort), timeout)
	if err != nil {
		return nil, err
	}

	c := Connection{conn: conn}

	c.Peers, err = c.GetPeers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting Peers")
	}

	return &c, nil
}

func (qc *Connection) GetPeers(ctx context.Context) (types.PublicPeers, error) {
	var result types.PublicPeers
	err := qc.sendRequest(ctx, types.CurrentTickInfoRequest, nil, &result)
	if err != nil {
		return types.PublicPeers{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Connection) GetIdentity(ctx context.Context, id string) (types.AddressInfo, error) {
	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.AddressInfo{}, errors.Wrap(err, "converting identity to public key")
	}

	var result types.AddressInfo
	err = qc.sendRequest(ctx, types.BalanceTypeRequest, pubKey, &result)
	if err != nil {
		return types.AddressInfo{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Connection) GetTickInfo(ctx context.Context) (types.TickInfo, error) {
	var result types.TickInfo

	err := qc.sendRequest(ctx, types.CurrentTickInfoRequest, nil, &result)
	if err != nil {
		return types.TickInfo{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Connection) GetTxStatus(ctx context.Context, tick uint32, digest [32]byte, sig [64]byte) (types.TransactionStatus, error) {
	request := struct {
		Tick      uint32
		Digest    [32]byte
		Signature [64]byte
	}{
		Tick:      tick,
		Digest:    digest,
		Signature: sig,
	}

	var result types.TransactionStatus

	err := qc.sendRequest(ctx, types.TxStatusRequest, request, &result)
	if err != nil {
		return types.TransactionStatus{}, errors.Wrap(err, "sending generic req")
	}

	return result, nil
}

func (qc *Connection) GetTickData(ctx context.Context, tickNumber uint32) (types.TickData, error) {
	tickInfo, err := qc.GetTickInfo(ctx)
	if err != nil {
		return types.TickData{}, errors.Wrap(err, "getting tick info")
	}

	if tickInfo.Tick < tickNumber {
		return types.TickData{}, errors.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	}

	request := struct{ Tick uint32 }{Tick: tickNumber}

	var result types.TickData
	err = qc.sendRequest(ctx, types.TickDataRequest, request, &result)
	if err != nil {
		return types.TickData{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Connection) GetTickTransactions(ctx context.Context, tickNumber uint32) (types.Transactions, error) {
	tickData, err := qc.GetTickData(ctx, tickNumber)
	var nrTx int
	for _, digest := range tickData.TransactionDigests {
		if digest == [32]byte{} {
			continue
		}
		nrTx++
	}

	if nrTx == 0 {
		return types.Transactions{}, nil
	}

	requestTickTransactions := struct {
		Tick             uint32
		TransactionFlags [types.NumberOfTransactionsPerTick / 8]uint8
	}{Tick: tickNumber}

	for i := 0; i < (nrTx+7)/8; i++ {
		requestTickTransactions.TransactionFlags[i] = 0
	}
	for i := (nrTx + 7) / 8; i < types.NumberOfTransactionsPerTick/8; i++ {
		requestTickTransactions.TransactionFlags[i] = 1
	}

	var result types.Transactions
	err = qc.sendRequest(ctx, types.TickTransactionsRequest, requestTickTransactions, &result)
	if err != nil {
		return nil, errors.Wrap(err, "sending transaction req")
	}

	return result, nil
}

func (qc *Connection) SendRawTransaction(ctx context.Context, rawTx []byte) error {
	err := qc.sendRequest(ctx, types.BroadcastTransaction, rawTx, nil)
	if err != nil {
		return errors.Wrap(err, "sending req")
	}

	return nil
}

func (qc *Connection) GetQuorumVotes(ctx context.Context, tickNumber uint32) (types.QuorumVotes, error) {
	tickInfo, err := qc.GetTickInfo(ctx)
	if err != nil {
		return types.QuorumVotes{}, errors.Wrap(err, "getting tick info")
	}

	if tickInfo.Tick < tickNumber {
		return types.QuorumVotes{}, errors.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	}

	request := struct {
		Tick      uint32
		VoteFlags [(types.NumberOfComputors + 7) / 8]byte
	}{Tick: tickNumber}

	var result types.QuorumVotes
	err = qc.sendRequest(ctx, types.QuorumTickRequest, request, &result)
	if err != nil {
		return types.QuorumVotes{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Connection) GetComputors(ctx context.Context) (types.Computors, error) {
	var result types.Computors
	err := qc.sendRequest(ctx, types.ComputorsRequest, nil, &result)
	if err != nil {
		return types.Computors{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Connection) sendRequest(ctx context.Context, requestType uint8, requestData interface{}, dest ReaderUnmarshaler) error {
	packet, err := serializeRequest(ctx, requestType, requestData)
	if err != nil {
		return errors.Wrap(err, "serializing request")
	}
	err = qc.writePacketToConn(ctx, packet)
	if err != nil {
		return errors.Wrap(err, "sending packet to qubic conn")
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = qc.readPacketIntoDest(ctx, dest)
	if err != nil {
		return errors.Wrap(err, "reading response")
	}

	return nil
}

func (qc *Connection) writePacketToConn(ctx context.Context, packet []byte) error {
	if packet == nil {
		return nil
	}

	// context deadline overrides defaultTimeout deadline
	writeDeadline := time.Now().Add(defaultTimeout)
	deadline, ok := ctx.Deadline()
	if ok {
		writeDeadline = deadline
	}
	err := qc.conn.SetWriteDeadline(writeDeadline)
	if err != nil {
		return errors.Wrap(err, "setting write deadline")
	}
	defer qc.conn.SetWriteDeadline(time.Time{})

	_, err = qc.conn.Write(packet)
	if err != nil {
		return errors.Wrap(err, "writing serialized binary data to connection")
	}

	return nil
}

func (qc *Connection) readPacketIntoDest(ctx context.Context, dest ReaderUnmarshaler) error {
	if dest == nil {
		return nil
	}

	// context deadline overrides defaultTimeout deadline
	readDeadline := time.Now().Add(defaultTimeout)
	deadline, ok := ctx.Deadline()
	if ok {
		readDeadline = deadline
	}

	err := qc.conn.SetReadDeadline(readDeadline)
	if err != nil {
		return errors.Wrap(err, "setting read deadline")
	}
	defer qc.conn.SetReadDeadline(time.Time{})

	err = dest.UnmarshallFromReader(qc.conn)
	if err != nil {
		return errors.Wrap(err, "unmarshalling response")
	}

	return nil
}

// Close closes the connection
func (qc *Connection) Close() error {
	return qc.conn.Close()
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

func serializeRequest(ctx context.Context, requestType uint8, requestData interface{}) ([]byte, error) {
	serializedReqData, err := serializeBinary(requestData)
	if err != nil {
		return nil, errors.Wrap(err, "serializing req data")
	}

	var header types.RequestResponseHeader

	packetHeaderSize := binary.Size(header)
	reqDataSize := len(serializedReqData)
	packetSize := uint32(packetHeaderSize + reqDataSize)

	header.SetSize(packetSize)
	header.RandomizeDejaVu()
	header.Type = requestType

	serializedHeaderData, err := serializeBinary(header)
	if err != nil {
		return nil, errors.Wrap(err, "serializing header data")
	}

	serializedPacket := make([]byte, 0, packetSize)
	serializedPacket = append(serializedPacket, serializedHeaderData...)
	serializedPacket = append(serializedPacket, serializedReqData...)

	return serializedPacket, nil
}
