package qubic

import (
	"context"
	"encoding/binary"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/foundation/tcp"
	"github.com/qubic/go-node-connector/types"
)

type Client struct {
	qc *tcp.QubicConnection
}

func NewClient(ctx context.Context, nodeIP, nodePort string) (*Client, error) {
	qc, err := tcp.NewQubicConnection(ctx, nodeIP, nodePort)
	if err != nil {
		return nil, errors.Wrap(err, "creating qubic connection")
	}

	return &Client{qc: qc}, nil
}

func (c *Client) GetIdentity(ctx context.Context, id string) (types.GetIdentityResponse, error) {
	type requestPacket struct {
		PublicKey [32]byte
	}

	request := requestPacket{PublicKey: getPublicKeyFromIdentity(id)}

	var result types.GetIdentityResponse
	err := tcp.SendGenericRequest(ctx, c.qc, types.BalanceTypeRequest, types.BalanceTypeResponse, request, &result)
	if err != nil {
		return types.GetIdentityResponse{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (c *Client) GetTickInfo(ctx context.Context) (types.CurrentTickInfo, error) {
	var result types.CurrentTickInfo

	err := tcp.SendGenericRequest(ctx, c.qc, types.CurrentTickInfoRequest, types.CurrentTickInfoResponse, nil, &result)
	if err != nil {
		return types.CurrentTickInfo{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (c *Client) GetTxStatus(ctx context.Context, qc *tcp.QubicConnection, tick uint32, digest [32]byte, sig [64]byte) (types.ResponseTxStatus, error) {
	request := types.RequestTxStatus{
		Tick:      tick,
		Digest:    digest,
		Signature: sig,
	}

	var result types.ResponseTxStatus

	err := tcp.SendGenericRequest(ctx, c.qc, types.TxStatusRequest, types.TxStatusResponse, request, &result)
	if err != nil {
		return types.ResponseTxStatus{}, errors.Wrap(err, "sending generic req")
	}

	return result, nil
}

func (c *Client) GetTickTransactions(ctx context.Context, tickNumber uint32) ([]types.Transaction, error) {
	tickData, err := c.GetTickData(ctx, tickNumber)
	var nrTx int
	for _, digest := range tickData.TransactionDigests {
		if digest == [32]byte{} {
			continue
		}
		nrTx++
	}

	requestTickTransactions := types.RequestTickTransactions{Tick: tickNumber}
	for i := 0; i < (nrTx+7)/8; i++ {
		requestTickTransactions.TransactionFlags[i] = 0
	}
	for i := (nrTx + 7) / 8; i < types.NumberOfTransactionsPerTick/8; i++ {
		requestTickTransactions.TransactionFlags[i] = 1
	}

	txs, err := tcp.SendGetTransactionsRequest(ctx, c.qc, types.TickTransactionsRequest, types.BroadcastTransaction, requestTickTransactions, nrTx)
	if err != nil {
		return nil, errors.Wrap(err, "sending transaction req")
	}

	transactions := make([]types.Transaction, 0, len(txs))

	for _, txData := range txs {
		hash, err := getHashFromTxData(txData)
		if err != nil {
			return nil, errors.Wrapf(err, "getting hash from tx data: %+v", txData)
		}
		transactions = append(transactions, types.Transaction{Data: txData, Hash: hash})
	}

	return transactions, nil
}

func (c *Client) GetTickData(ctx context.Context, tickNumber uint32) (types.TickData, error) {
	tickInfo, err := c.GetTickInfo(ctx)
	if err != nil {
		return types.TickData{}, errors.Wrap(err, "getting tick info")
	}

	if tickInfo.Tick < tickNumber {
		return types.TickData{}, errors.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	}

	request := types.RequestTickData{Tick: tickNumber}

	var result types.TickData
	err = tcp.SendGenericRequest(ctx, c.qc, types.TickDataRequest, types.BroadcastFutureTickData, request, &result)
	if err != nil {
		return types.TickData{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (c *Client) SendRawTransaction(ctx context.Context, rawTx []byte) error {
	err := tcp.SendTransaction(ctx, c.qc, types.BroadcastTransaction, 0, rawTx, nil)
	if err != nil {
		return errors.Wrap(err, "sending req")
	}

	return nil
}

func (c *Client) GetQuorumTickData(ctx context.Context, tickNumber uint32) (types.ResponseQuorumTickData, error) {
	//tickInfo, err := c.GetTickInfo(ctx)
	//if err != nil {
	//	return types.ResponseQuorumTickData{}, errors.Wrap(err, "getting tick info")
	//}
	//
	//if tickInfo.Tick < tickNumber {
	//	return types.ResponseQuorumTickData{}, errors.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	//}

	request := types.RequestQuorumTickData{Tick: tickNumber}

	//var result types.ResponseQuorumTickData
	quorumTicks, err := tcp.SendGetQuorumTickDataRequest(ctx, c.qc, types.QuorumTickRequest, types.QuorumTickResponse, request)
	if err != nil {
		return types.ResponseQuorumTickData{}, errors.Wrap(err, "sending req to node")
	}

	return types.ResponseQuorumTickData{QuorumData: quorumTicks}, nil
}

func (c *Client) Close() error {
	if c.qc != nil {
		return c.qc.Close()
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

func getHashFromTxData(txData types.TransactionData) (types.TransactionHash, error) {
	txDataMarshalledBytes, err := txData.MarshallBinary()
	if err != nil {
		return types.TransactionHash{}, errors.Wrap(err, "marshalling")
	}

	h := k12.NewDraft10([]byte{})
	_, err = h.Write(txDataMarshalledBytes)
	if err != nil {
		return types.TransactionHash{}, errors.Wrap(err, "writing msg to k12")
	}

	var digest [32]byte
	_, err = h.Read(digest[:])
	if err != nil {
		return types.TransactionHash{}, errors.Wrap(err, "reading hash from k12")
	}

	hash, err := getTxHashFromDigestFromPubKey(digest)
	if err != nil {
		return types.TransactionHash{}, errors.Wrap(err, "getting id from pubkey")
	}

	return hash, err
}

func getTxHashFromDigestFromPubKey(digest [32]byte) ([60]byte, error) {
	var hash [60]byte

	for i := 0; i < 4; i++ {
		var publicKeyFragment = binary.LittleEndian.Uint64(digest[i*8 : (i+1)*8])
		for j := 0; j < 14; j++ {
			hash[i*14+j] = byte((publicKeyFragment % 26) + 'a')
			publicKeyFragment /= 26
		}
	}

	h := k12.NewDraft10([]byte{})
	_, err := h.Write(hash[:])
	if err != nil {
		return [60]byte{}, errors.Wrap(err, "writing msg to k12")
	}

	var identityBytesChecksum [3]byte
	_, err = h.Read(identityBytesChecksum[:])
	if err != nil {
		return [60]byte{}, errors.Wrap(err, "reading hash from k12")
	}

	var identityBytesChecksumInt uint64
	identityBytesChecksumInt = uint64(identityBytesChecksum[0]) | (uint64(identityBytesChecksum[1]) << 8) | (uint64(identityBytesChecksum[2]) << 16)
	identityBytesChecksumInt &= 0x3FFFF

	for i := 0; i < 4; i++ {
		hash[56+i] = byte((identityBytesChecksumInt % 26) + 'a')
		identityBytesChecksumInt /= 26
	}

	return hash, nil
}
