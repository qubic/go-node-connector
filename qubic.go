package qubic

import (
	"context"
	"github.com/0xluk/go-qubic/data/identity"
	"github.com/0xluk/go-qubic/data/tick"
	"github.com/0xluk/go-qubic/data/tx"
	"github.com/0xluk/go-qubic/foundation/tcp"
	"github.com/pkg/errors"
)

type Client struct {
	Qc *tcp.QubicConnection
}

func NewClient(ctx context.Context, nodeIP, nodePort string) (*Client, error) {
	qc, err := tcp.NewQubicConnection(ctx, nodeIP, nodePort)
	if err != nil {
		return nil, errors.Wrap(err, "creating qubic connection")
	}

	return &Client{Qc: qc}, nil
}

func GetIdentity(ctx context.Context, qc *tcp.QubicConnection, id string) (identity.GetIdentityResponse, error) {
	type requestPacket struct {
		PublicKey [32]byte
	}

	request := requestPacket{PublicKey: getPublicKeyFromIdentity(id)}

	var result identity.GetIdentityResponse
	err := tcp.SendGenericRequest(ctx, qc, identity.RequestBalanceType, identity.RespondBalanceType, request, &result)
	if err != nil {
		return identity.GetIdentityResponse{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func GetTickInfo(ctx context.Context, qc *tcp.QubicConnection) (tick.CurrentTickInfo, error) {
	var result tick.CurrentTickInfo

	err := tcp.SendGenericRequest(ctx, qc, tick.REQUEST_CURRENT_TICK_INFO, tick.RESPOND_CURRENT_TICK_INFO, nil, &result)
	if err != nil {
		return tick.CurrentTickInfo{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func GetTxStatus(ctx context.Context, qc *tcp.QubicConnection, tick uint32, digest [32]byte, sig [64]byte) (tx.ResponseTxStatus, error) {
	request := tx.RequestTxStatus{
		Tick:      tick,
		Digest:    digest,
		Signature: sig,
	}

	var result tx.ResponseTxStatus

	err := tcp.SendGenericRequest(ctx, qc, tx.REQUEST_TX_STATUS, tx.RESPONSE_TX_STATUS, request, &result)
	if err != nil {
		return tx.ResponseTxStatus{}, errors.Wrap(err, "sending generic req")
	}

	return result, nil
}

func GetTickTransactions(ctx context.Context, qc *tcp.QubicConnection, tickNumber uint32) ([]tick.Transaction, error) {
	tickData, err := GetTickData(ctx, qc, tickNumber)
	var nrTx int
	for _, digest := range tickData.TransactionDigests {
		if digest == [32]byte{} {
			continue
		}
		nrTx++
	}


	requestTickTransactions := tick.RequestTickTransactions{Tick: tickNumber}
	for i := 0; i < (nrTx+7)/8; i++ {
		requestTickTransactions.TransactionFlags[i] = 0
	}
	for i := (nrTx + 7) / 8; i < tick.NUMBER_OF_TRANSACTIONS_PER_TICK/8; i++ {
		requestTickTransactions.TransactionFlags[i] = 1
	}

	txs, err := tcp.SendTransactionsRequest(ctx, qc, tick.REQUEST_TICK_TRANSACTIONS, tick.BROADCAST_TRANSACTION, requestTickTransactions, nrTx)
	if err != nil {
		return nil, errors.Wrap(err, "sending transaction req")
	}

	return txs, nil
}

func GetTickData(ctx context.Context, qc *tcp.QubicConnection, tickNumber uint32) (tick.TickData, error) {
	tickInfo, err := GetTickInfo(ctx, qc)
	if err != nil {
		return tick.TickData{}, errors.Wrap(err, "getting tick info")
	}

	if tickInfo.Tick < tickNumber {
		return tick.TickData{}, errors.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	}

	request := tick.RequestTickData{Tick: tickNumber}

	var result tick.TickData
	err = tcp.SendGenericRequest(ctx, qc, tick.REQUEST_TICK_DATA, tick.BROADCAST_FUTURE_TICK_DATA, request, &result)
	if err != nil {
		return tick.TickData{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func SendRawTransaction(ctx context.Context, qc *tcp.QubicConnection, tx tx.SignedTransaction) error {
	err := tcp.SendGenericRequest(ctx, qc, tick.BROADCAST_TRANSACTION, 0, tx, nil)
	if err != nil {
		return errors.Wrap(err, "sending req")
	}

	return nil
}

func (c Client) Close() error {
	if c.Qc != nil {
		return c.Qc.Close()
	}

	return nil
}

func getPublicKeyFromIdentity(identity string) [32]byte {
	publicKeyBuffer := make([]byte, 32)

	for i := 0; i < 4; i++ {
		value := uint64(0)
		for j := 13; j >= 0; j-- {
			if identity[i*14+j] < 'A' || identity[i*14+j] > 'Z' {
				return [32]byte{} // Error condition: invalid character in identity
			}

			value = value*26 + uint64(identity[i*14+j]-'A')
		}

		// Copy the 8-byte value into publicKeyBuffer
		for k := 0; k < 8; k++ {
			publicKeyBuffer[i*8+k] = byte(value >> (k * 8))
		}
	}

	var pubKey [32]byte
	copy(pubKey[:], publicKeyBuffer[:32])

	return pubKey
}