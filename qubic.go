package qubic

import (
	"context"
	"github.com/0xluk/go-qubic/foundation/tcp"
	"github.com/pkg/errors"
)

type Client struct {
	qc *tcp.QubicConnection
}

func NewClient(nodeIP, nodePort string) (Client, error) {
	qc, err := tcp.NewQubicConnection(nodeIP, nodePort)
	if err != nil {
		return Client{}, errors.Wrap(err, "creating qubic connection")
	}

	return Client{qc: qc}, nil
}

func (c Client) GetBalance(ctx context.Context, identity string) (GetBalanceResponse, error) {
	type requestPacket struct {
		PublicKey [32]byte
	}

	request := requestPacket{PublicKey: getPublicKeyFromIdentity(identity)}

	var result GetBalanceResponse
	err := tcp.SendRequest(ctx, c.qc, RequestBalanceType, RespondBalanceType, request, &result)
	if err != nil {
		return GetBalanceResponse{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (c Client) Close() error {
	if c.qc != nil {
		return c.qc.Close()
	}

	return nil
}