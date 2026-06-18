package qubic

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/qubic/go-node-connector/v2/types"
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
		return nil, fmt.Errorf("getting Peers: %w", err)
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
		return types.PublicPeers{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetIssuedAssets(ctx context.Context, id string) (types.IssuedAssets, error) {

	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.IssuedAssets{}, fmt.Errorf("converting identity to public key: %w", err)
	}
	var result types.IssuedAssets
	err = qc.sendRequest(ctx, types.IssuedAssetsRequest, pubKey, &result)
	if err != nil {
		return types.IssuedAssets{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil

}

func (qc *Client) GetPossessedAssets(ctx context.Context, id string) (types.PossessedAssets, error) {

	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.PossessedAssets{}, fmt.Errorf("converting identity to public key: %w", err)
	}
	var result types.PossessedAssets
	err = qc.sendRequest(ctx, types.PossessedAssetsRequest, pubKey, &result)
	if err != nil {
		return types.PossessedAssets{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetOwnedAssets(ctx context.Context, id string) (types.OwnedAssets, error) {

	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.OwnedAssets{}, fmt.Errorf("converting identity to public key: %w", err)
	}
	var result types.OwnedAssets
	err = qc.sendRequest(ctx, types.OwnedAssetsRequest, pubKey, &result)
	if err != nil {
		return types.OwnedAssets{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetIdentity(ctx context.Context, id string) (types.AddressInfo, error) {
	identity := types.Identity(id)
	pubKey, err := identity.ToPubKey(false)
	if err != nil {
		return types.AddressInfo{}, fmt.Errorf("converting identity to public key: %w", err)
	}

	var result types.AddressInfo
	err = qc.sendRequest(ctx, types.BalanceTypeRequest, pubKey, &result)
	if err != nil {
		return types.AddressInfo{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetTickInfo(ctx context.Context) (types.TickInfo, error) {
	var result types.TickInfo

	err := qc.sendRequest(ctx, types.CurrentTickInfoRequest, nil, &result)
	if err != nil {
		return types.TickInfo{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetSystemInfo(ctx context.Context) (types.SystemInfo, error) {
	var result types.SystemInfo

	err := qc.sendRequest(ctx, types.SystemInfoRequest, nil, &result)
	if err != nil {
		return types.SystemInfo{}, fmt.Errorf("sending req to node: %w", err)
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
		return types.TransactionStatus{}, fmt.Errorf("sending generic req: %w", err)
	}

	return result, nil
}

func (qc *Client) GetTickData(ctx context.Context, tickNumber uint32) (types.TickData, error) {
	tickInfo, err := qc.GetTickInfo(ctx)
	if err != nil {
		return types.TickData{}, fmt.Errorf("getting tick info: %w", err)
	}

	if tickInfo.Tick < tickNumber {
		return types.TickData{}, fmt.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	}

	request := struct{ Tick uint32 }{Tick: tickNumber}

	var result types.TickData
	err = qc.sendRequest(ctx, types.TickDataRequest, request, &result)
	if err != nil {
		return types.TickData{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetTickTransactions(ctx context.Context, tickNumber uint32) (types.Transactions, error) {
	tickData, err := qc.GetTickData(ctx, tickNumber)
	if err != nil {
		return types.Transactions{}, fmt.Errorf("getting tick data: %w", err)
	}

	nrTx := getTickTransactionsNrTx(tickData)
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
		return nil, fmt.Errorf("sending transaction req: %w", err)
	}

	var validTxs = make([]types.Transaction, 0, nrTx)
	for _, tx := range result {
		// check if it's a 0 transaction
		if tx.Signature == [64]byte{} {
			continue
		}

		validTxs = append(validTxs, tx)
	}

	return validTxs, nil
}

// start counting from backwards to see how many transactions are in the tick data
func getTickTransactionsNrTx(tickData types.TickData) int {
	for i := len(tickData.TransactionDigests) - 1; i >= 0; i-- {
		if tickData.TransactionDigests[i] != [32]byte{} {
			return i + 1
		}
	}

	return 0
}

// TickPrefetch bundles the per-tick responses gathered in one prefetch batch.
type TickPrefetch struct {
	Tick         uint32
	TickData     types.TickData
	QuorumVotes  types.QuorumVotes
	Transactions types.Transactions
}

// PrefetchResult is the full result of a single pipelined prefetch batch:
// the once-per-batch responses plus one TickPrefetch per requested tick,
// ordered by tick ascending.
type PrefetchResult struct {
	SystemInfo types.SystemInfo
	TickInfo   types.TickInfo
	Ticks      []TickPrefetch
}

// prefetchOp is a single request/response pair in a pipelined batch. Each op is
// tagged with a unique DejaVu nonce: the node echoes it in the response header,
// which is how we route each response back to the dest that requested it - the
// node does not guarantee responses arrive in request order.
type prefetchOp struct {
	requestType uint8
	requestData interface{}
	dest        ReaderUnmarshaler
	dejaVu      uint32
	done        bool
}

// PrefetchTicks pipelines a batch of requests over the single underlying
// connection: it writes SystemInfo + TickInfo + (QuorumVotes, TickData,
// TickTransactions) for every tick in [startTick, startTick+nrTicks) back to
// back, then drains every response in the same order. This pays roughly one
// round-trip for the whole range instead of one per request.
//
// nrTicks must be > 0. The range is not validated against the chain tip - the
// caller is responsible for ensuring the requested ticks exist. Fail-fast: the
// first write or read error aborts the batch, closes the (now unaligned)
// connection so it is never reused, and returns the error.
//
// The node does not guarantee responses arrive in request order, so each
// request carries a unique DejaVu nonce that the node echoes in the response
// header; the read phase routes every response back to its op by that nonce.
func (qc *Client) PrefetchTicks(ctx context.Context, startTick, nrTicks uint32) (PrefetchResult, error) {
	if nrTicks == 0 {
		return PrefetchResult{}, errors.New("nrTicks must be greater than 0")
	}

	result := PrefetchResult{Ticks: make([]TickPrefetch, nrTicks)}

	ops := make([]prefetchOp, 0, 2+int(nrTicks)*3)

	// once-per-batch requests
	ops = append(ops,
		prefetchOp{requestType: types.SystemInfoRequest, dest: &result.SystemInfo},
		prefetchOp{requestType: types.CurrentTickInfoRequest, dest: &result.TickInfo},
	)

	// per-tick requests; dest pointers stay stable because Ticks is preallocated
	for i := uint32(0); i < nrTicks; i++ {
		tick := startTick + i
		result.Ticks[i].Tick = tick

		ops = append(ops,
			prefetchOp{requestType: types.QuorumTickRequest, requestData: newQuorumVotesRequest(tick), dest: &result.Ticks[i].QuorumVotes},
			prefetchOp{requestType: types.TickDataRequest, requestData: newTickDataRequest(tick), dest: &result.Ticks[i].TickData},
			prefetchOp{requestType: types.TickTransactionsRequest, requestData: newAllTickTransactionsRequest(tick), dest: &result.Ticks[i].Transactions},
		)
	}

	// assign a unique, non-zero DejaVu to every op so responses can be matched
	byDejaVu := make(map[uint32]*prefetchOp, len(ops))
	for i := range ops {
		dejaVu := uint32(rand.Int31())
		for dejaVu == 0 || byDejaVu[dejaVu] != nil {
			dejaVu = uint32(rand.Int31())
		}
		ops[i].dejaVu = dejaVu
		byDejaVu[dejaVu] = &ops[i]
	}

	// write phase: pipeline every request without reading
	for i := range ops {
		op := &ops[i]
		packet, err := serializeRequestWithDejaVu(op.requestType, op.requestData, op.dejaVu)
		if err != nil {
			return PrefetchResult{}, fmt.Errorf("serializing prefetch request type %d: %w", op.requestType, err)
		}
		if err := qc.writePacketToConn(ctx, packet); err != nil {
			qc.Close()
			return PrefetchResult{}, fmt.Errorf("writing prefetch request type %d: %w", op.requestType, err)
		}
	}

	// read phase: demultiplex interleaved packets and route them by DejaVu
	if err := qc.drainPrefetch(ctx, ops, byDejaVu); err != nil {
		qc.Close()
		return PrefetchResult{}, fmt.Errorf("draining prefetch responses: %w", err)
	}

	// drop the zeroed transactions returned for empty/requested-but-absent slots
	for i := range result.Ticks {
		result.Ticks[i].Transactions = filterValidTransactions(result.Ticks[i].Transactions)
	}

	return result, nil
}

// drainPrefetch reads packets off the connection until every op's full response
// has been buffered, then decodes each buffer into its dest. The node interleaves
// packets from different responses, so packets are routed to per-op buffers by the
// DejaVu nonce echoed in each header and framed by the header's Size field. A
// multi-packet response (quorum votes, transactions) is complete once its
// terminating EndResponse arrives; a single-packet response is complete on its
// data packet. Unsolicited ExchangePublicPeers packets are skipped.
func (qc *Client) drainPrefetch(ctx context.Context, ops []prefetchOp, byDejaVu map[uint32]*prefetchOp) error {
	headerSize := binary.Size(types.RequestResponseHeader{})

	buffers := make(map[uint32]*bytes.Buffer, len(ops))
	for i := range ops {
		buffers[ops[i].dejaVu] = &bytes.Buffer{}
	}

	remaining := len(ops)
	for remaining > 0 {
		readDeadline := time.Now().Add(defaultTimeout)
		if deadline, ok := ctx.Deadline(); ok {
			readDeadline = deadline
		}
		if err := qc.conn.SetReadDeadline(readDeadline); err != nil {
			return fmt.Errorf("setting read deadline: %w", err)
		}

		headerBytes := make([]byte, headerSize)
		if _, err := io.ReadFull(qc.conn, headerBytes); err != nil {
			return fmt.Errorf("reading packet header: %w", err)
		}

		size := uint32(headerBytes[0]) | uint32(headerBytes[1])<<8 | uint32(headerBytes[2])<<16
		respType := headerBytes[3]
		dejaVu := binary.LittleEndian.Uint32(headerBytes[4:])

		bodyLen := int(size) - headerSize
		if bodyLen < 0 {
			return fmt.Errorf("invalid packet size %d for type %d", size, respType)
		}
		body := make([]byte, bodyLen)
		if _, err := io.ReadFull(qc.conn, body); err != nil {
			return fmt.Errorf("reading packet body for type %d: %w", respType, err)
		}

		// unsolicited peer-exchange packets aren't tied to any request
		if respType == types.ExchangePublicPeers {
			continue
		}

		op, ok := byDejaVu[dejaVu]
		if !ok {
			return fmt.Errorf("unexpected response: dejaVu %d, type %d", dejaVu, respType)
		}
		if op.done {
			// trailing packet (e.g. EndResponse) for an already-complete response
			continue
		}

		buf := buffers[dejaVu]
		buf.Write(headerBytes)
		buf.Write(body)

		if isResponseComplete(op.requestType, respType) {
			op.done = true
			remaining--
		}
	}

	qc.conn.SetReadDeadline(time.Time{})

	// every response is fully buffered now; decode each into its dest
	for i := range ops {
		op := &ops[i]
		if err := op.dest.UnmarshallFromReader(buffers[op.dejaVu]); err != nil {
			return fmt.Errorf("unmarshalling response type %d: %w", op.requestType, err)
		}
	}

	return nil
}

// isResponseComplete reports whether respType terminates the response for a given
// request type. Quorum and transaction responses stream multiple packets ending
// in EndResponse. All other responses are a single packet: the node replies with
// exactly one packet that is either the data packet or a bare EndResponse (e.g. an
// empty tick), so the first routed packet completes the response regardless of type.
func isResponseComplete(requestType, respType uint8) bool {
	switch requestType {
	case types.QuorumTickRequest, types.TickTransactionsRequest:
		return respType == types.EndResponse
	default:
		return true
	}
}

func newTickDataRequest(tick uint32) interface{} {
	return struct{ Tick uint32 }{Tick: tick}
}

func newQuorumVotesRequest(tick uint32) interface{} {
	return struct {
		Tick      uint32
		VoteFlags [(types.NumberOfComputors + 7) / 8]byte
		_Pad      [3]byte
	}{Tick: tick}
}

// newAllTickTransactionsRequest requests every transaction slot in the tick:
// a zero flag means "send this one", so an all-zero flag set asks for all 4096.
func newAllTickTransactionsRequest(tick uint32) interface{} {
	return struct {
		Tick             uint32
		TransactionFlags [types.NumberOfTransactionsPerTick / 8]uint8
	}{Tick: tick}
}

func filterValidTransactions(txs types.Transactions) types.Transactions {
	valid := make(types.Transactions, 0, len(txs))
	for _, tx := range txs {
		// skip zeroed (empty) transactions
		if tx.Signature == [64]byte{} {
			continue
		}
		valid = append(valid, tx)
	}
	return valid
}

func (qc *Client) SendRawTransaction(ctx context.Context, rawTx []byte) error {
	err := qc.sendRequest(ctx, types.BroadcastTransaction, rawTx, nil)
	if err != nil {
		return fmt.Errorf("sending req: %w", err)
	}

	return nil
}

func (qc *Client) GetQuorumVotes(ctx context.Context, tickNumber uint32) (types.QuorumVotes, error) {
	tickInfo, err := qc.GetTickInfo(ctx)
	if err != nil {
		return types.QuorumVotes{}, fmt.Errorf("getting tick info: %w", err)
	}

	if tickInfo.Tick < tickNumber {
		return types.QuorumVotes{}, fmt.Errorf("Requested tick %d is in the future. Latest tick is: %d", tickNumber, tickInfo.Tick)
	}

	request := struct {
		Tick      uint32
		VoteFlags [(types.NumberOfComputors + 7) / 8]byte
		_Pad      [3]byte
	}{Tick: tickNumber}

	var result types.QuorumVotes
	err = qc.sendRequest(ctx, types.QuorumTickRequest, request, &result)
	if err != nil {
		return types.QuorumVotes{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetComputors(ctx context.Context) (types.Computors, error) {
	var result types.Computors
	err := qc.sendRequest(ctx, types.ComputorsRequest, nil, &result)
	if err != nil {
		return types.Computors{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) QuerySmartContract(ctx context.Context, rcf RequestContractFunction, requestData []byte) (types.SmartContractData, error) {
	var result types.SmartContractData
	err := qc.sendSmartContractRequest(ctx, rcf, types.ContractFunctionRequest, requestData, &result)
	if err != nil {
		return types.SmartContractData{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetActiveIpos(ctx context.Context) (types.Ipos, error) {
	var result types.Ipos
	err := qc.sendRequest(ctx, types.ActiveIposRequest, nil, &result)
	if err != nil {
		return types.Ipos{}, fmt.Errorf("sending req to node: %w", err)
	}

	return result, nil
}

func (qc *Client) GetContractIpo(ctx context.Context, contractIndex uint32) (types.ContractIpo, error) {
	var result types.ContractIpo
	err := qc.sendRequest(ctx, types.ContractIpoRequest, contractIndex, &result)
	if err != nil {
		return types.ContractIpo{}, fmt.Errorf("requesting contract ipo for contract index %d: %w", contractIndex, err)
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
		return types.AssetPossessions{}, fmt.Errorf("creating request object: %w", err)
	}

	var result types.AssetPossessions
	err = qc.sendRequest(ctx, types.RequestAssets, request, &result)
	if err != nil {
		return types.AssetPossessions{}, fmt.Errorf("sending request to node: %w", err)
	}
	return result, nil
}

func (qc *Client) GetAssetOwnershipsByFilter(ctx context.Context, issuerIdentity, assetName,
	ownerIdentity string, ownerContract uint16) (types.AssetOwnerships, error) {

	request, err := createGetAssetOwnershipsByFilterRequest(
		AssetInformation{issuerIdentity, assetName},
		AssetHolderInformation{ownerIdentity, ownerContract})

	if err != nil {
		return types.AssetOwnerships{}, fmt.Errorf("creating request object: %w", err)
	}

	var result types.AssetOwnerships
	err = qc.sendRequest(ctx, types.RequestAssets, request, &result)
	if err != nil {
		return types.AssetOwnerships{}, fmt.Errorf("sending request to node: %w", err)
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
			return RequestAssetsByFilter{}, fmt.Errorf("converting issuer identity to public key: %w", err)
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
			return RequestAssetsByFilter{}, fmt.Errorf("converting owner identity to public key: %w", err)
		}
		owner = pubKey
	}

	var possessor = [32]byte{}
	if len(possessorInfo.Identity) > 0 {
		identity := types.Identity(possessorInfo.Identity)
		pubKey, err := identity.ToPubKey(false)
		if err != nil {
			return RequestAssetsByFilter{}, fmt.Errorf("converting possessor identity to public key: %w", err)
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
		return types.AssetIssuances{}, fmt.Errorf("creating request object: %w", err)
	}

	var result types.AssetIssuances
	err = qc.sendRequest(ctx, types.RequestAssets, request, &result)
	if err != nil {
		return types.AssetIssuances{}, fmt.Errorf("sending request to node: %w", err)
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
			return RequestAssetsByFilter{}, fmt.Errorf("converting issuer identity to public key: %w", err)
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
	RequestType   uint16 // 2B
	Flags         uint16 // 2B
	UniverseIndex uint32 // 4B
}

func (qc *Client) GetAssetIssuancesByUniverseIndex(ctx context.Context, index uint32) (types.AssetIssuances, error) {
	var result types.AssetIssuances
	err := qc.getAssetByUniverseIndex(ctx, index, &result)
	if err != nil {
		return types.AssetIssuances{}, err
	}
	return result, nil
}

func (qc *Client) GetAssetOwnershipsByUniverseIndex(ctx context.Context, index uint32) (types.AssetOwnerships, error) {
	var result types.AssetOwnerships
	err := qc.getAssetByUniverseIndex(ctx, index, &result)
	if err != nil {
		return types.AssetOwnerships{}, err
	}
	return result, nil
}

func (qc *Client) GetAssetPossessionsByUniverseIndex(ctx context.Context, index uint32) (types.AssetPossessions, error) {
	var result types.AssetPossessions
	err := qc.getAssetByUniverseIndex(ctx, index, &result)
	if err != nil {
		return types.AssetPossessions{}, err
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
		return fmt.Errorf("sending req to node: %w", err)
	}
	return nil

}

func (qc *Client) sendRequest(ctx context.Context, requestType uint8, requestData interface{}, dest ReaderUnmarshaler) error {
	packet, err := serializeRequest(ctx, requestType, requestData)
	if err != nil {
		return fmt.Errorf("serializing request for req type %d: %w", requestType, err)
	}

	err = qc.writePacketToConn(ctx, packet)
	if err != nil {
		return fmt.Errorf("sending packet to qubic conn for req type %d: %w", requestType, err)
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = qc.readPacketIntoDest(ctx, dest)
	if err != nil {
		return fmt.Errorf("reading response for req type %d: %w", requestType, err)
	}

	return nil
}

func (qc *Client) sendSmartContractRequest(ctx context.Context, rcf RequestContractFunction, requestType uint8, requestData []byte, dest ReaderUnmarshaler) error {
	packet, err := serializesSmartContractRequest(ctx, rcf, requestType, requestData)
	if err != nil {
		return fmt.Errorf("serializing request for req type %d: %w", requestType, err)
	}

	err = qc.writePacketToConn(ctx, packet)
	if err != nil {
		return fmt.Errorf("sending packet to qubic conn for req type %d: %w", requestType, err)
	}

	// if dest is nil then we don't care about the response
	if dest == nil {
		return nil
	}

	err = qc.readPacketIntoDest(ctx, dest)
	if err != nil {
		return fmt.Errorf("reading response for req type %d: %w", requestType, err)
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
		return fmt.Errorf("setting write deadline: %w", err)
	}
	defer qc.conn.SetWriteDeadline(time.Time{})

	_, err = qc.conn.Write(packet)
	if err != nil {
		return fmt.Errorf("writing serialized binary data to connection: %w", err)
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
		return fmt.Errorf("setting read deadline: %w", err)
	}
	defer qc.conn.SetReadDeadline(time.Time{})

	err = dest.UnmarshallFromReader(qc.conn)
	if err != nil {
		return fmt.Errorf("unmarshalling response: %w", err)
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
		return nil, fmt.Errorf("writing data to buff: %w", err)
	}

	return buff.Bytes(), nil
}

func serializeRequest(ctx context.Context, requestType uint8, requestData interface{}) ([]byte, error) {
	var dejaVu uint32
	if requestType != types.BroadcastTransaction {
		dejaVu = uint32(rand.Int31())
		if dejaVu == 0 {
			dejaVu = 1
		}
	}

	return serializeRequestWithDejaVu(requestType, requestData, dejaVu)
}

// serializeRequestWithDejaVu builds a request packet with an explicit DejaVu
// nonce, used by pipelined prefetch so responses can be matched back to requests.
func serializeRequestWithDejaVu(requestType uint8, requestData interface{}, dejaVu uint32) ([]byte, error) {
	serializedReqData, err := serializeBinary(requestData)
	if err != nil {
		return nil, fmt.Errorf("serializing req data: %w", err)
	}

	var header types.RequestResponseHeader

	packetHeaderSize := binary.Size(header)
	reqDataSize := len(serializedReqData)
	packetSize := uint32(packetHeaderSize + reqDataSize)

	header.SetSize(packetSize)
	header.DejaVu = dejaVu
	header.Type = requestType

	serializedHeaderData, err := serializeBinary(header)
	if err != nil {
		return nil, fmt.Errorf("serializing header data: %w", err)
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
		return nil, fmt.Errorf("serializing req contract function: %w", err)
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
		return nil, fmt.Errorf("serializing header data: %w", err)
	}

	serializedPacket := make([]byte, 0, packetSize)
	serializedPacket = append(serializedPacket, serializedHeaderData...)
	serializedPacket = append(serializedPacket, serializedReqContractFunction...)
	serializedPacket = append(serializedPacket, serializedReqData...)

	return serializedPacket, nil
}
