package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"io"
)

const SendManyMaxTransfers = 25
const QutilAddress = "EAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVWRF"
const QutilSendManyInputType = 1
const QutilSendManyFee = 10
const QutilSendManyInputSize = 1000

type SendManyTransferPayload struct {
	addresses       [SendManyMaxTransfers][32]byte
	amounts         [SendManyMaxTransfers]int64
	filledTransfers int8
	totalAmount     int64
}

type SendManyTransfer struct {
	AddressID Identity
	Amount    int64
}

func (smp *SendManyTransferPayload) AddTransfer(transfer SendManyTransfer) error {
	if smp.filledTransfers == SendManyMaxTransfers {
		return errors.Errorf("max %d send many transfers allowed", SendManyMaxTransfers)
	}

	address, err := transfer.AddressID.ToPubKey(false)
	if err != nil {
		return errors.Wrap(err, "converting address id to byte form")
	}

	smp.addresses[smp.filledTransfers] = address
	smp.amounts[smp.filledTransfers] = transfer.Amount
	smp.filledTransfers += 1
	smp.totalAmount += transfer.Amount

	return nil
}

func (smp *SendManyTransferPayload) AddTransfers(transfers []SendManyTransfer) error {
	if int(smp.filledTransfers)+len(transfers) > SendManyMaxTransfers {
		return errors.Errorf("max %d send many transfers allowed", SendManyMaxTransfers)
	}

	for _, transfer := range transfers {
		err := smp.AddTransfer(transfer)
		if err != nil {
			return errors.Wrapf(err, "adding transfer %+v", transfer)
		}
	}

	return nil
}

func (smp *SendManyTransferPayload) GetTransfers() ([]SendManyTransfer, error) {
	transfers := make([]SendManyTransfer, 0, SendManyMaxTransfers)
	for index, address := range smp.addresses {
		if address == [32]byte{} {
			continue
		}
		var addrID Identity
		addrID, err := addrID.FromPubKey(address, false)
		if err != nil {
			return nil, errors.Wrapf(err, "getting address identity from bytes %v", address)
		}
		transfers = append(transfers, SendManyTransfer{AddressID: addrID, Amount: smp.amounts[index]})
	}

	return transfers, nil
}

// GetTotalAmount returns total amount of transfers + SC fee
func (smp *SendManyTransferPayload) GetTotalAmount() int64 {
	return smp.totalAmount + QutilSendManyFee
}

func (smp *SendManyTransferPayload) MarshallBinary() ([]byte, error) {
	var buff bytes.Buffer
	err := binary.Write(&buff, binary.LittleEndian, smp.addresses)
	if err != nil {
		return nil, errors.Wrap(err, "writing addresses to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, smp.amounts)
	if err != nil {
		return nil, errors.Wrap(err, "writing amounts to buf")
	}

	return buff.Bytes(), nil
}

func (smp *SendManyTransferPayload) UnmarshallBinary(b []byte) error {
	reader := bytes.NewReader(b)

	err := binary.Read(reader, binary.LittleEndian, &smp.addresses)
	if err != nil {
		return errors.Wrap(err, "reading addresses from reader")
	}

	err = binary.Read(reader, binary.LittleEndian, &smp.amounts)
	if err != nil {
		return errors.Wrap(err, "reading amounts from reader")
	}

	return nil
}

func NewSimpleTransferTransaction(sourceID, destinationID string, amount int64, targetTick uint32) (Transaction, error) {
	srcID := Identity(sourceID)
	destID := Identity(destinationID)
	srcPubKey, err := srcID.ToPubKey(false)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "converting src id string to pubkey")
	}
	destPubKey, err := destID.ToPubKey(false)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "converting dest id string to pubkey")
	}

	return Transaction{
		SourcePublicKey:      srcPubKey,
		DestinationPublicKey: destPubKey,
		Amount:               amount,
		Tick:                 targetTick,
		InputType:            0,
		InputSize:            0,
		Input:                nil,
	}, nil
}

func NewSendManyTransferTransaction(sourceID string, targetTick uint32, payload SendManyTransferPayload) (Transaction, error) {
	srcID := Identity(sourceID)
	destID := Identity(QutilAddress)
	srcPubKey, err := srcID.ToPubKey(false)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "converting src id string to pubkey")
	}
	destPubKey, err := destID.ToPubKey(false)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "converting dest id string to pubkey")
	}

	input, err := payload.MarshallBinary()
	if err != nil {
		return Transaction{}, errors.Wrap(err, "binary marshalling payload")
	}

	return Transaction{
		SourcePublicKey:      srcPubKey,
		DestinationPublicKey: destPubKey,
		Amount:               payload.GetTotalAmount(),
		Tick:                 targetTick,
		InputType:            QutilSendManyInputType,
		InputSize:            QutilSendManyInputSize,
		Input:                input,
	}, nil
}

type Transaction struct {
	SourcePublicKey      [32]byte
	DestinationPublicKey [32]byte
	Amount               int64
	Tick                 uint32
	InputType            uint16
	InputSize            uint16
	Input                []byte
	Signature            [64]byte
}

func (tx *Transaction) GetUnsignedDigest() ([32]byte, error) {
	serialized, err := tx.MarshallBinary()
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "marshalling tx data")
	}

	// create digest with data without signature
	digest, err := k12Hash(serialized[:len(serialized)-64])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "hashing tx data")
	}

	return digest, nil
}

func (tx *Transaction) MarshallBinary() ([]byte, error) {
	var buff bytes.Buffer
	_, err := buff.Write(tx.SourcePublicKey[:])
	if err != nil {
		return nil, errors.Wrap(err, "writing source public key to buffer")
	}

	_, err = buff.Write(tx.DestinationPublicKey[:])
	if err != nil {
		return nil, errors.Wrap(err, "writing destination public key to buffer")
	}
	err = binary.Write(&buff, binary.LittleEndian, tx.Amount)
	if err != nil {
		return nil, errors.Wrap(err, "writing amount to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, tx.Tick)
	if err != nil {
		return nil, errors.Wrap(err, "writing tick to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, tx.InputType)
	if err != nil {
		return nil, errors.Wrap(err, "writing input type to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, tx.InputSize)
	if err != nil {
		return nil, errors.Wrap(err, "writing input size to buf")
	}

	_, err = buff.Write(tx.Input)
	if err != nil {
		return nil, errors.Wrap(err, "writing input to buffer")
	}

	_, err = buff.Write(tx.Signature[:])
	if err != nil {
		return nil, errors.Wrap(err, "writing signature to buffer")
	}

	return buff.Bytes(), nil
}

func (tx *Transaction) UnmarshallBinary(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &tx.SourcePublicKey)
	if err != nil {
		return errors.Wrap(err, "reading source public key from reader")
	}

	err = binary.Read(r, binary.LittleEndian, &tx.DestinationPublicKey)
	if err != nil {
		return errors.Wrap(err, "reading destination public key from reader")
	}

	err = binary.Read(r, binary.LittleEndian, &tx.Amount)
	if err != nil {
		return errors.Wrap(err, "reading amount from reader")
	}

	err = binary.Read(r, binary.LittleEndian, &tx.Tick)
	if err != nil {
		return errors.Wrap(err, "reading tick from reader")
	}

	err = binary.Read(r, binary.LittleEndian, &tx.InputType)
	if err != nil {
		return errors.Wrap(err, "reading input type from reader")
	}

	err = binary.Read(r, binary.LittleEndian, &tx.InputSize)
	if err != nil {
		return errors.Wrap(err, "reading input size from reader")
	}

	tx.Input = make([]byte, tx.InputSize)
	err = binary.Read(r, binary.LittleEndian, &tx.Input)
	if err != nil {
		return errors.Wrap(err, "reading input from reader")
	}

	err = binary.Read(r, binary.LittleEndian, &tx.Signature)
	if err != nil {
		return errors.Wrap(err, "reading signature from reader")
	}

	return nil
}

func (tx *Transaction) Digest() ([32]byte, error) {
	serialized, err := tx.MarshallBinary()
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "marshalling tx data")
	}

	digest, err := k12Hash(serialized)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "hashing tx data")
	}

	return digest, nil
}

func (tx *Transaction) ID() (string, error) {
	digest, err := tx.Digest()
	if err != nil {
		return "", errors.Wrap(err, "getting digest")
	}

	var id Identity
	id, err = id.FromPubKey(digest, true)
	if err != nil {
		return "", errors.Wrap(err, "getting id from pubkey")
	}

	return id.String(), nil
}

func (tx *Transaction) EncodeToBase64() (string, error) {
	txPacket, err := tx.MarshallBinary()
	if err != nil {
		return "", errors.Wrap(err, "binary marshalling")
	}

	return base64.StdEncoding.EncodeToString(txPacket[:]), nil
}

type Transactions []Transaction

func (txs *Transactions) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return errors.Wrap(err, "reading header")
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != BroadcastTransaction {
			return errors.Errorf("Invalid header type, expected %d, found %d", BroadcastTransaction, header.Type)
		}

		var tx Transaction

		err = tx.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "unmarshalling transaction")
		}

		*txs = append(*txs, tx)
	}

	return nil
}

type TransactionStatus struct {
	CurrentTickOfNode  uint32
	Tick               uint32
	TransactionsStatus []IndividualTransactionStatus
}

type IndividualTransactionStatus struct {
	TxID      string
	MoneyFlew bool
}

type transactionStatusResponse struct {
	CurrentTickOfNode  uint32
	Tick               uint32
	TxCount            uint32
	MoneyFlew          [(NumberOfTransactionsPerTick + 7) / 8]byte
	TransactionDigests [][32]byte
}

func (ts *TransactionStatus) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader

	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return errors.Wrap(err, "reading header")
	}

	if header.Type != TxStatusResponse {
		return errors.Errorf("Invalid header type, expected %d, found %d", TxStatusResponse, header.Type)
	}

	var response transactionStatusResponse

	err = binary.Read(r, binary.LittleEndian, &response.CurrentTickOfNode)
	if err != nil {
		return errors.Wrap(err, "reading current tick of node")
	}

	err = binary.Read(r, binary.LittleEndian, &response.Tick)
	if err != nil {
		return errors.Wrap(err, "reading tick")
	}

	err = binary.Read(r, binary.LittleEndian, &response.TxCount)
	if err != nil {
		return errors.Wrap(err, "reading tx count")
	}

	err = binary.Read(r, binary.LittleEndian, &response.MoneyFlew)
	if err != nil {
		return errors.Wrap(err, "reading reading money flew")
	}

	response.TransactionDigests = make([][32]byte, response.TxCount)
	err = binary.Read(r, binary.LittleEndian, &response.TransactionDigests)
	if err != nil {
		return errors.Wrap(err, "reading tx digests")
	}

	return nil
}

func (ts *TransactionStatus) convertResponseToModel(response transactionStatusResponse) error {
	transactionsStatus := make([]IndividualTransactionStatus, 0, response.TxCount)

	for index, txDigest := range response.TransactionDigests {
		var id Identity
		id, err := id.FromPubKey(txDigest, true)
		if err != nil {
			return errors.Wrap(err, "converting digest to id")
		}
		moneyFlew := getMoneyFlewFromBits(response.MoneyFlew, index)

		txStatus := IndividualTransactionStatus{
			TxID:      id.String(),
			MoneyFlew: moneyFlew,
		}

		transactionsStatus = append(transactionsStatus, txStatus)
	}

	*ts = TransactionStatus{
		CurrentTickOfNode:  response.CurrentTickOfNode,
		Tick:               response.Tick,
		TransactionsStatus: transactionsStatus,
	}

	return nil
}

func getMoneyFlewFromBits(input [(NumberOfTransactionsPerTick + 7) / 8]byte, digestIndex int) bool {
	pos := digestIndex / 8
	bitIndex := digestIndex % 8

	return getNthBit(input[pos], bitIndex)
}

func getNthBit(input byte, bitIndex int) bool {
	// Shift the input byte to the right by the bitIndex positions
	// This isolates the bit at the bitIndex position at the least significant bit position
	shifted := input >> bitIndex

	// Extract the least significant bit using a bitwise AND operation with 1
	// If the least significant bit is 1, the result will be 1; otherwise, it will be 0
	return shifted&1 == 1
}

func k12Hash(data []byte) ([32]byte, error) {
	h := k12.NewDraft10([]byte{}) // Using K12 for hashing, equivalent to KangarooTwelve(temp, 96, h, 64).
	_, err := h.Write(data)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "k12 hashing")
	}

	var out [32]byte
	_, err = h.Read(out[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading k12 digest")
	}

	return out, nil
}
