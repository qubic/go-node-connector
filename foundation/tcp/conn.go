package tcp

import (
	"context"
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"net"
	"time"
)

type QubicConnection struct {
	conn net.Conn
}

func NewQubicConnection(nodeIP, nodePort string) (*QubicConnection, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(nodeIP, nodePort))
	if err != nil {
		return nil, err
	}
	return &QubicConnection{conn: conn}, nil
}

func (qc *QubicConnection) SendHeaderData(ctx context.Context, data RequestResponseHeader) error {
	// set write deadline only if context has a deadline
	deadline, ok := ctx.Deadline()
	if ok {
		err := qc.conn.SetWriteDeadline(deadline)
		if err != nil {
			return errors.Wrap(err, "setting write deadline")
		}
		defer qc.conn.SetWriteDeadline(time.Time{})
	}

	err := binary.Write(qc.conn, binary.LittleEndian, data)
	if err != nil {
		return errors.Wrap(err, "writing serialized binary data to connection")
	}

	return nil
}

func (qc *QubicConnection) SendRequestData(ctx context.Context, data interface{}) error {
	if data == nil {
		return nil
	}

	// set write deadline only if context has a deadline
	deadline, ok := ctx.Deadline()
	if ok {
		err := qc.conn.SetWriteDeadline(deadline)
		if err != nil {
			return errors.Wrap(err, "setting write deadline")
		}
		defer qc.conn.SetWriteDeadline(time.Time{})
	}

	err := binary.Write(qc.conn, binary.LittleEndian, data)
	if err != nil {
		return errors.Wrap(err, "writing serialized binary data to connection")
	}

	return nil
}

// ReceiveDataAll reads all available data from the connection
func (qc *QubicConnection) ReceiveDataAll() ([]byte, error) {
	err := qc.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "setting readResponse deadline")
	}
	defer qc.conn.SetReadDeadline(time.Time{})

	receivedData, err := io.ReadAll(qc.conn)
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return receivedData, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "reading from conn")
	}

	return receivedData, nil
}

// Close closes the connection
func (qc *QubicConnection) Close() error {
	return qc.conn.Close()
}
