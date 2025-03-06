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

type Client struct {
	conn  net.Conn
	Peers types.PublicPeers
}

func NewClient(ctx context.Context, nodeIP, nodePort string) (*Client, error) {
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

	c := Client{conn: conn}

	c.Peers, err = c.getPeers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting Peers")
	}

	return &c, nil
}

func NewClientWithConn(ctx context.Context, conn net.Conn) (*Client, error) {
	return &Client{conn: conn}, nil
}

func (qc *Client) getPeers(ctx context.Context) (types.PublicPeers, error) {
	var result types.PublicPeers
	err := qc.sendRequest(ctx, types.CurrentTickInfoRequest, nil, &result)
	if err != nil {
		return types.PublicPeers{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Client) GetIssuedAssets(ctx context.Context, id string) (types.IssuedAssets, error) {

	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.IssuedAssets{}, errors.Wrap(err, "converting identity to public key")
	}
	var result types.IssuedAssets
	err = qc.sendRequest(ctx, types.IssuedAssetsRequest, pubKey, &result)
	if err != nil {
		return types.IssuedAssets{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil

}

func (qc *Client) GetPossessedAssets(ctx context.Context, id string) (types.PossessedAssets, error) {

	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.PossessedAssets{}, errors.Wrap(err, "converting identity to public key")
	}
	var result types.PossessedAssets
	err = qc.sendRequest(ctx, types.PossessedAssetsRequest, pubKey, &result)
	if err != nil {
		return types.PossessedAssets{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Client) GetOwnedAssets(ctx context.Context, id string) (types.OwnedAssets, error) {

	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.OwnedAssets{}, errors.Wrap(err, "converting identity to public key")
	}
	var result types.OwnedAssets
	err = qc.sendRequest(ctx, types.OwnedAssetsRequest, pubKey, &result)
	if err != nil {
		return types.OwnedAssets{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Client) GetIdentity(ctx context.Context, id string) (types.AddressInfo, error) {
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

func (qc *Client) GetTickInfo(ctx context.Context) (types.TickInfo, error) {
	var result types.TickInfo

	err := qc.sendRequest(ctx, types.CurrentTickInfoRequest, nil, &result)
	if err != nil {
		return types.TickInfo{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Client) GetTxStatus(ctx context.Context, tick uint32) (types.TransactionStatus, error) {
	request := struct {
		Tick uint32
	}{
		Tick: tick,
	}

	var result types.TransactionStatus
	err := qc.sendRequest(ctx, types.TxStatusRequest, request, &result)
	if err != nil {
		return types.TransactionStatus{}, errors.Wrap(err, "sending generic req")
	}

	return result, nil
}

func (qc *Client) GetTickData(ctx context.Context, tickNumber uint32) (types.TickData, error) {
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

func (qc *Client) GetTickTransactions(ctx context.Context, tickNumber uint32) (types.Transactions, error) {
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

func (qc *Client) SendRawTransaction(ctx context.Context, rawTx []byte) error {
	err := qc.sendRequest(ctx, types.BroadcastTransaction, rawTx, nil)
	if err != nil {
		return errors.Wrap(err, "sending req")
	}

	return nil
}

func (qc *Client) GetQuorumVotes(ctx context.Context, tickNumber uint32) (types.QuorumVotes, error) {
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

func (qc *Client) GetComputors(ctx context.Context) (types.Computors, error) {
	var result types.Computors
	err := qc.sendRequest(ctx, types.ComputorsRequest, nil, &result)
	if err != nil {
		return types.Computors{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

func (qc *Client) QuerySmartContract(ctx context.Context, rcf RequestContractFunction, requestData []byte) (types.SmartContractData, error) {
	var result types.SmartContractData
	err := qc.sendSmartContractRequest(ctx, rcf, types.ContractFunctionRequest, requestData, &result)
	if err != nil {
		return types.SmartContractData{}, errors.Wrap(err, "sending req to node")
	}

	return result, nil
}

const requestTypeAssetIssuanceRecords uint16 = 0
const requestTypeAssetOwnershipRecords uint16 = 1
const requestTypeAssetPossessionRecords uint16 = 2
const requestTypeAssetByUniverseIndex uint16 = 3

const flagAnyIssuer uint16 = 0b10
const flagAnyAssetName uint16 = 0b100
const flagAnyOwner uint16 = 0b1000
const flagAnyOwnerContract uint16 = 0b10000
const flagAnyPossessor uint16 = 0b100000
const flagAnyPossessorContract uint16 = 0b1000000

type RequestAssetsByFilter struct {
	RequestType                uint16
	Flags                      uint16
	OwnershipManagingContract  uint16
	PossessionManagingContract uint16
	Issuer                     [32]byte
	AssetName                  [8]byte
	Owner                      [32]byte
	Possessor                  [32]byte
}

func (qc *Client) GetAssetPossessionsByFilter(ctx context.Context, issuerIdentity, assetName,
	ownerIdentity, possessorIdentity string, ownerContract, possessorContract uint16) (types.AssetPossessions, error) {

	request, err := createGetAssetPossessionsByFilterRequest(
		AssetInformation{issuerIdentity, assetName},
		AssetHolderInformation{ownerIdentity, ownerContract},
		AssetHolderInformation{possessorIdentity, possessorContract})

	if err != nil {
		return types.AssetPossessions{}, errors.Wrap(err, "creating request object")
	}

	var result types.AssetPossessions
	err = qc.sendRequest(ctx, types.RequestAssets, request, &result)
	if err != nil {
		return types.AssetPossessions{}, errors.Wrap(err, "sending request to node")
	}
	return result, nil
}

func (qc *Client) GetAssetOwnershipsByFilter(ctx context.Context, issuerIdentity, assetName,
	ownerIdentity string, ownerContract uint16) (types.AssetOwnerships, error) {

	request, err := createGetAssetOwnershipsByFilterRequest(
		AssetInformation{issuerIdentity, assetName},
		AssetHolderInformation{ownerIdentity, ownerContract})

	if err != nil {
		return types.AssetOwnerships{}, errors.Wrap(err, "creating request object")
	}

	var result types.AssetOwnerships
	err = qc.sendRequest(ctx, types.RequestAssets, request, &result)
	if err != nil {
		return types.AssetOwnerships{}, errors.Wrap(err, "sending request to node")
	}
	return result, nil
}

type AssetInformation struct {
	Identity string
	Name     string
}

type AssetHolderInformation struct {
	Identity string
	Contract uint16
}

func createGetAssetOwnershipsByFilterRequest(assetInfo AssetInformation, ownerInfo AssetHolderInformation) (RequestAssetsByFilter, error) {
	return createByFilterRequest(requestTypeAssetOwnershipRecords, assetInfo, ownerInfo, AssetHolderInformation{})
}

func createGetAssetPossessionsByFilterRequest(assetInfo AssetInformation, ownerInfo, possessorInfo AssetHolderInformation) (RequestAssetsByFilter, error) {
	return createByFilterRequest(requestTypeAssetPossessionRecords, assetInfo, ownerInfo, possessorInfo)
}

func createByFilterRequest(requestType uint16, assetInfo AssetInformation, ownerInfo, possessorInfo AssetHolderInformation) (RequestAssetsByFilter, error) {
	var issuer = [32]byte{}
	if len(assetInfo.Identity) > 0 {
		identity := types.Identity(assetInfo.Identity)
		pubKey, err := identity.ToPubKey(false)
		if err != nil {
			return RequestAssetsByFilter{}, errors.Wrap(err, "converting issuer identity to public key")
		}
		issuer = pubKey
	}

	if len(assetInfo.Name) == 0 {
		return RequestAssetsByFilter{}, errors.New("asset name is required")
	}
	var name [8]byte
	copy(name[:], assetInfo.Name)

	var owner = [32]byte{}
	if len(ownerInfo.Identity) > 0 {
		identity := types.Identity(ownerInfo.Identity)
		pubKey, err := identity.ToPubKey(false)
		if err != nil {
			return RequestAssetsByFilter{}, errors.Wrap(err, "converting owner identity to public key")
		}
		owner = pubKey
	}

	var possessor = [32]byte{}
	if len(possessorInfo.Identity) > 0 {
		identity := types.Identity(possessorInfo.Identity)
		pubKey, err := identity.ToPubKey(false)
		if err != nil {
			return RequestAssetsByFilter{}, errors.Wrap(err, "converting possessor identity to public key")
		}
		possessor = pubKey
	}

	request := RequestAssetsByFilter{
		RequestType:                requestType,
		Flags:                      getFlags(ownerInfo, possessorInfo),
		OwnershipManagingContract:  ownerInfo.Contract,     // pad
		PossessionManagingContract: possessorInfo.Contract, // pad
		Issuer:                     issuer,
		AssetName:                  name,
		Owner:                      owner,     // pad
		Possessor:                  possessor, // pad
	}

	return request, nil
}

func getFlags(ownerInfo, possessorInfo AssetHolderInformation) uint16 {
	var flags uint16 = 0
	if len(ownerInfo.Identity) == 0 {
		flags |= flagAnyOwner
	}
	if len(possessorInfo.Identity) == 0 {
		flags |= flagAnyPossessor
	}
	if ownerInfo.Contract == 0 {
		flags |= flagAnyOwnerContract
	}
	if possessorInfo.Contract == 0 {
		flags |= flagAnyPossessorContract
	}
	return flags
}

func (qc *Client) GetAssetIssuancesByFilter(ctx context.Context, issuerIdentity, assetName string) (types.AssetIssuances, error) {
	request, err := createAssetIssuancesByFilterRequest(issuerIdentity, assetName)
	if err != nil {
		return types.AssetIssuances{}, errors.Wrap(err, "creating request object")
	}

	var result types.AssetIssuances
	err = qc.sendRequest(ctx, types.RequestAssets, request, &result)
	if err != nil {
		return types.AssetIssuances{}, errors.Wrap(err, "sending request to node")
	}
	return result, nil
}

func createAssetIssuancesByFilterRequest(issuerIdentity, assetName string) (RequestAssetsByFilter, error) {
	var flags uint16 = 0

	var issuer [32]byte
	if len(issuerIdentity) == 0 {
		flags |= flagAnyIssuer
		issuer = [32]byte{}
	} else {
		identity := types.Identity(issuerIdentity)
		pubKey, err := identity.ToPubKey(false)
		if err != nil {
			return RequestAssetsByFilter{}, errors.Wrap(err, "converting issuer identity to public key")
		}
		issuer = pubKey
	}

	var name [8]byte
	copy(name[:], assetName)
	if len(assetName) == 0 {
		flags |= flagAnyAssetName
	}

	request := RequestAssetsByFilter{
		RequestType:                requestTypeAssetIssuanceRecords,
		Flags:                      flags,
		OwnershipManagingContract:  0, // pad
		PossessionManagingContract: 0, // pad
		Issuer:                     issuer,
		AssetName:                  name,
		Owner:                      [32]byte{}, // pad
		Possessor:                  [32]byte{}, // pad
	}

	return request, nil
}

type RequestAssetsByUniverseIndex struct {
	RequestType   uint16    // 2B
	Flags         uint16    // 2B
	UniverseIndex uint32    // 4B
	Padding       [104]byte // 104B
}

func (qc *Client) GetAssetIssuancesByUniverseIndex(ctx context.Context, index uint32) (types.AssetIssuances, error) {
	var result types.AssetIssuances
	err := qc.getAssetByUniverseIndex(ctx, index, &result)
	if err != nil {
		return types.AssetIssuances{}, errors.Wrap(err, "get asset by universe index")
	}
	return result, nil
}

func (qc *Client) GetAssetOwnershipsByUniverseIndex(ctx context.Context, index uint32) (types.AssetOwnerships, error) {
	var result types.AssetOwnerships
	err := qc.getAssetByUniverseIndex(ctx, index, &result)
	if err != nil {
		return types.AssetOwnerships{}, errors.Wrap(err, "get asset by universe index")
	}
	return result, nil
}

func (qc *Client) GetAssetPossessionsByUniverseIndex(ctx context.Context, index uint32) (types.AssetPossessions, error) {
	var result types.AssetPossessions
	err := qc.getAssetByUniverseIndex(ctx, index, &result)
	if err != nil {
		return types.AssetPossessions{}, errors.Wrap(err, "get asset by universe index")
	}
	return result, nil
}

func (qc *Client) getAssetByUniverseIndex(ctx context.Context, index uint32, destination ReaderUnmarshaler) error {

	request := RequestAssetsByUniverseIndex{
		RequestType:   requestTypeAssetByUniverseIndex,
		UniverseIndex: index,
	}

	err := qc.sendRequest(ctx, types.RequestAssets, request, destination)
	if err != nil {
		return errors.Wrap(err, "sending req to node")
	}
	return nil

}

func (qc *Client) sendRequest(ctx context.Context, requestType uint8, requestData interface{}, dest ReaderUnmarshaler) error {
	packet, err := serializeRequest(ctx, requestType, requestData)
	if err != nil {
		return errors.Wrapf(err, "serializing request for req type %d", requestType)
	}

	err = qc.writePacketToConn(ctx, packet)
	if err != nil {
		return errors.Wrapf(err, "sending packet to qubic conn for req type %d", requestType)
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = qc.readPacketIntoDest(ctx, dest)
	if err != nil {
		return errors.Wrapf(err, "reading response for req type %d", requestType)
	}

	return nil
}

func (qc *Client) sendSmartContractRequest(ctx context.Context, rcf RequestContractFunction, requestType uint8, requestData []byte, dest ReaderUnmarshaler) error {
	packet, err := serializesSmartContractRequest(ctx, rcf, requestType, requestData)
	if err != nil {
		return errors.Wrapf(err, "serializing request for req type %d", requestType)
	}

	err = qc.writePacketToConn(ctx, packet)
	if err != nil {
		return errors.Wrapf(err, "sending packet to qubic conn for req type %d", requestType)
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = qc.readPacketIntoDest(ctx, dest)
	if err != nil {
		return errors.Wrapf(err, "reading response for req type %d", requestType)
	}

	return nil
}

func (qc *Client) writePacketToConn(ctx context.Context, packet []byte) error {
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

func (qc *Client) readPacketIntoDest(ctx context.Context, dest ReaderUnmarshaler) error {
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
func (qc *Client) Close() error {
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
	if requestType == types.BroadcastTransaction {
		header.ZeroDejaVu()
	} else {
		header.RandomizeDejaVu()
	}

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

type RequestContractFunction struct {
	ContractIndex uint32
	InputType     uint16
	InputSize     uint16
}

func serializesSmartContractRequest(ctx context.Context, rcf RequestContractFunction, requestType uint8, requestData []byte) ([]byte, error) {
	serializedReqData := requestData
	serializedReqContractFunction, err := serializeBinary(rcf)
	if err != nil {
		return nil, errors.Wrap(err, "serializing req contract function")
	}

	var header types.RequestResponseHeader

	packetHeaderSize := binary.Size(header)
	reqDataSize := len(serializedReqData)
	reqContractFunctionSize := len(serializedReqContractFunction)
	packetSize := uint32(packetHeaderSize + reqContractFunctionSize + reqDataSize)

	header.RandomizeDejaVu()

	header.Type = requestType
	header.SetSize(packetSize)

	serializedHeaderData, err := serializeBinary(header)
	if err != nil {
		return nil, errors.Wrap(err, "serializing header data")
	}

	serializedPacket := make([]byte, 0, packetSize)
	serializedPacket = append(serializedPacket, serializedHeaderData...)
	serializedPacket = append(serializedPacket, serializedReqContractFunction...)
	serializedPacket = append(serializedPacket, serializedReqData...)

	return serializedPacket, nil
}
