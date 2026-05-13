package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cloudflare/circl/xof/k12"
	"github.com/qubic/go-schnorrq"
)

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
		return [32]byte{}, fmt.Errorf("marshalling tx data: %w", err)
	}

	// create digest with data without signature
	digest, err := k12Hash(serialized[:len(serialized)-64])
	if err != nil {
		return [32]byte{}, fmt.Errorf("hashing tx data: %w", err)
	}

	return digest, nil
}

func (tx *Transaction) MarshallBinary() ([]byte, error) {
	var buff bytes.Buffer
	_, err := buff.Write(tx.SourcePublicKey[:])
	if err != nil {
		return nil, fmt.Errorf("writing source public key to buffer: %w", err)
	}

	_, err = buff.Write(tx.DestinationPublicKey[:])
	if err != nil {
		return nil, fmt.Errorf("writing destination public key to buffer: %w", err)
	}
	err = binary.Write(&buff, binary.LittleEndian, tx.Amount)
	if err != nil {
		return nil, fmt.Errorf("writing amount to buf: %w", err)
	}

	err = binary.Write(&buff, binary.LittleEndian, tx.Tick)
	if err != nil {
		return nil, fmt.Errorf("writing tick to buf: %w", err)
	}

	err = binary.Write(&buff, binary.LittleEndian, tx.InputType)
	if err != nil {
		return nil, fmt.Errorf("writing input type to buf: %w", err)
	}

	err = binary.Write(&buff, binary.LittleEndian, tx.InputSize)
	if err != nil {
		return nil, fmt.Errorf("writing input size to buf: %w", err)
	}

	_, err = buff.Write(tx.Input)
	if err != nil {
		return nil, fmt.Errorf("writing input to buffer: %w", err)
	}

	_, err = buff.Write(tx.Signature[:])
	if err != nil {
		return nil, fmt.Errorf("writing signature to buffer: %w", err)
	}

	return buff.Bytes(), nil
}

func (tx *Transaction) UnmarshallBinary(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &tx.SourcePublicKey)
	if err != nil {
		return fmt.Errorf("reading source public key from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &tx.DestinationPublicKey)
	if err != nil {
		return fmt.Errorf("reading destination public key from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &tx.Amount)
	if err != nil {
		return fmt.Errorf("reading amount from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &tx.Tick)
	if err != nil {
		return fmt.Errorf("reading tick from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &tx.InputType)
	if err != nil {
		return fmt.Errorf("reading input type from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &tx.InputSize)
	if err != nil {
		return fmt.Errorf("reading input size from reader: %w", err)
	}

	tx.Input = make([]byte, tx.InputSize)
	err = binary.Read(r, binary.LittleEndian, &tx.Input)
	if err != nil {
		return fmt.Errorf("reading input from reader: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &tx.Signature)
	if err != nil {
		return fmt.Errorf("reading signature from reader: %w", err)
	}

	return nil
}

func (tx *Transaction) Digest() ([32]byte, error) {
	serialized, err := tx.MarshallBinary()
	if err != nil {
		return [32]byte{}, fmt.Errorf("marshalling tx data: %w", err)
	}

	digest, err := k12Hash(serialized)
	if err != nil {
		return [32]byte{}, fmt.Errorf("hashing tx data: %w", err)
	}

	return digest, nil
}

func (tx *Transaction) ID() (string, error) {
	digest, err := tx.Digest()
	if err != nil {
		return "", fmt.Errorf("getting digest: %w", err)
	}

	var id Identity
	id, err = id.FromPubKey(digest, true)
	if err != nil {
		return "", fmt.Errorf("getting id from pubkey: %w", err)
	}

	return id.String(), nil
}

func (tx *Transaction) EncodeToBase64() (string, error) {
	txPacket, err := tx.MarshallBinary()
	if err != nil {
		return "", fmt.Errorf("binary marshalling: %w", err)
	}

	return base64.StdEncoding.EncodeToString(txPacket[:]), nil
}

func (tx *Transaction) Sign(seed string) error {
	unsignedDigest, err := tx.GetUnsignedDigest()
	if err != nil {
		return fmt.Errorf("getting tx unsigned digest: %w", err)
	}

	subSeed, err := GetSubSeed(seed)
	if err != nil {
		return fmt.Errorf("getting subseed: %w", err)
	}

	sig, err := schnorrq.Sign(subSeed, tx.SourcePublicKey, unsignedDigest)
	if err != nil {
		return fmt.Errorf("signing transaction: %w", err)
	}
	tx.Signature = sig

	return nil
}

type Transactions []Transaction

func (txs *Transactions) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != BroadcastTransaction {
			return fmt.Errorf("Invalid header type, expected %d, found %d", BroadcastTransaction, header.Type)
		}

		var tx Transaction

		err = tx.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling transaction: %w", err)
		}

		*txs = append(*txs, tx)
	}

	return nil
}

type TransactionStatus struct {
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
		return fmt.Errorf("reading header: %w", err)
	}

	if header.Type != TxStatusResponse {
		return fmt.Errorf("Invalid header type, expected %d, found %d", TxStatusResponse, header.Type)
	}

	err = binary.Read(r, binary.LittleEndian, &ts.CurrentTickOfNode)
	if err != nil {
		return fmt.Errorf("reading current tick of node: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ts.Tick)
	if err != nil {
		return fmt.Errorf("reading tick: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ts.TxCount)
	if err != nil {
		return fmt.Errorf("reading tx count: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ts.MoneyFlew)
	if err != nil {
		return fmt.Errorf("reading reading money flew: %w", err)
	}

	ts.TransactionDigests = make([][32]byte, ts.TxCount)
	err = binary.Read(r, binary.LittleEndian, &ts.TransactionDigests)
	if err != nil {
		return fmt.Errorf("reading tx digests: %w", err)
	}

	return nil
}

func k12Hash(data []byte) ([32]byte, error) {
	h := k12.NewDraft10([]byte{}) // Using K12 for hashing, equivalent to KangarooTwelve(temp, 96, h, 64).
	_, err := h.Write(data)
	if err != nil {
		return [32]byte{}, fmt.Errorf("k12 hashing: %w", err)
	}

	var out [32]byte
	_, err = h.Read(out[:])
	if err != nil {
		return [32]byte{}, fmt.Errorf("reading k12 digest: %w", err)
	}

	return out, nil
}
