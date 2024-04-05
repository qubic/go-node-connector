package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"io"
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
		Amount:               5,
		Tick:                 targetTick,
	}, nil
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
	CurrentTickOfNode uint32
	TickOfTx          uint32
	MoneyFlew         bool
	Executed          bool
	NotFound          bool
	Padding           [5]byte
	Digest            [32]byte
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

	err = binary.Read(r, binary.LittleEndian, ts)
	if err != nil {
		return errors.Wrap(err, "reading tx status data from reader")
	}

	return nil
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
