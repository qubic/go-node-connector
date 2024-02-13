package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"unicode"
)

const (
	NumberOfTransactionsPerTick = 1024
	SignatureSize               = 64
	CurrentTickInfoRequest      = 27
	CurrentTickInfoResponse     = 28
	BroadcastFutureTickData     = 8
	TickDataRequest             = 16
	TickTransactionsRequest     = 29
	BroadcastTransaction        = 24
	ArbitratorIdentity = "AFZPUAIYVPNUYGJRQVLUKOPPVLHAZQTGLYAAUUNBXFTVTAMSBKQBLEIEPCVJ"
)

type TickData struct {
	ComputorIndex      uint16
	Epoch              uint16
	Tick               uint32
	Millisecond        uint16
	Second             uint8
	Minute             uint8
	Hour               uint8
	Day                uint8
	Month              uint8
	Year               uint8
	UnionData          [256]byte
	Timelock           [32]byte
	TransactionDigests [NumberOfTransactionsPerTick][32]byte `json:",omitempty"`
	ContractFees       [1024]int64                           `json:",omitempty"`
	Signature          [SignatureSize]byte
}

type CurrentTickInfo struct {
	TickDuration            uint16
	Epoch                   uint16
	Tick                    uint32
	NumberOfAlignedVotes    uint16
	NumberOfMisalignedVotes uint16
}

type RequestTickData struct {
	Tick uint32
}

type TransactionHeader struct {
	SourcePublicKey      [32]byte
	DestinationPublicKey [32]byte
	Amount               int64
	Tick                 uint32
	InputType            uint16
	InputSize            uint16
}

func (th *TransactionHeader) MarshallBinary() ([]byte, error) {
	var buff bytes.Buffer
	err := binary.Write(&buff, binary.LittleEndian, th.SourcePublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "writing source pubkey to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.DestinationPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "writing dest pubkey to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.Amount)
	if err != nil {
		return nil, errors.Wrap(err, "writing amount to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.Tick)
	if err != nil {
		return nil, errors.Wrap(err, "writing tick to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.InputType)
	if err != nil {
		return nil, errors.Wrap(err, "writing input type to buf")
	}

	err = binary.Write(&buff, binary.LittleEndian, th.InputSize)
	if err != nil {
		return nil, errors.Wrap(err, "writing input size to buf")
	}

	return buff.Bytes(), nil
}

type TransactionData struct {
	Header    TransactionHeader
	Input     []byte
	Signature [64]byte
}

func (td *TransactionData) MarshallBinary() ([]byte, error) {
	var out []byte
	binaryHeader, err := td.Header.MarshallBinary()
	if err != nil {
		return nil, errors.Wrap(err, "writing txData to buf")
	}

	out = append(out, binaryHeader...)
	out = append(out, td.Input...)
	out = append(out, td.Signature[:]...)

	return out, nil
}

type TransactionInput []byte

type TransactionSig [65]byte

type TransactionHash [60]byte

type Transaction struct {
	Data TransactionData
	Hash TransactionHash
}

type RequestTickTransactions struct {
	Tick             uint32
	TransactionFlags [NumberOfTransactionsPerTick / 8]uint8
}

const (
	TxStatusRequest  = 201
	TxStatusResponse = 202
)

type RequestTxStatus struct {
	Tick      uint32
	Digest    [32]byte
	Signature [64]byte
}

type ResponseTxStatus struct {
	CurrentTickOfNode uint32
	TickOfTx          uint32
	MoneyFlew         bool
	Executed          bool
	NotFound          bool
	Padding           [5]byte
	Digest            [32]byte
}

const (
	SpectrumDepth       = 24
	BalanceTypeRequest  = 31
	BalanceTypeResponse = 32
)

type GetIdentityResponse struct {
	Entity        Entity
	Tick          uint32
	SpectrumIndex int32
	Siblings      [SpectrumDepth][32]byte
}

type Entity struct {
	PublicKey                  [32]byte
	IncomingAmount             int64
	OutgoingAmount             int64
	NumberOfIncomingTransfers  uint32
	NumberOfOutgoingTransfers  uint32
	LatestIncomingTransferTick uint32
	LatestOutgoingTransferTick uint32
}

const (
	NumberOfComputors  = 676
	QuorumTickRequest  = 14
	QuorumTickResponse = 3
	ComputorsRequest= 11
	BroadcastComputors = 2
	MinimumQuorumVotes = 451
)

type RequestQuorumTickData struct {
	Tick      uint32
	VoteFlags [(NumberOfComputors + 7) / 8]byte
}

type ResponseQuorumTickData struct {
	QuorumData []QuorumTickData
}

type QuorumTickData struct {
	ComputorIndex uint16
	Epoch         uint16
	Tick          uint32

	Millisecond uint16
	Second      uint8
	Minute      uint8
	Hour        uint8
	Day         uint8
	Month       uint8
	Year        uint8

	PreviousResourceTestingDigest uint64
	SaltedResourceTestingDigest   uint64

	PreviousSpectrumDigest [32]byte
	PreviousUniverseDigest [32]byte
	PreviousComputerDigest [32]byte

	SaltedSpectrumDigest [32]byte
	SaltedUniverseDigest [32]byte
	SaltedComputerDigest [32]byte

	TxDigest                 [32]byte
	ExpectedNextTickTxDigest [32]byte

	Signature [SignatureSize]byte
}

type ResponseComputors struct {
	Computors Computors
}

type Computors struct {
	Epoch     uint16
	PubKeys   [NumberOfComputors][32]byte
	Signature [SignatureSize]byte
}

type Identity string

// FromPubKey creates a new identity from a public key
// this DOES NOT alter the original value, you should only rely on the returned value
func (i *Identity) FromPubKey(pubKey [32]byte, isLowerCase bool) (Identity, error) {
	letter := 'A'
	if isLowerCase {
		letter = 'a'
	}


	var identity [60]byte

	for i := 0; i < 4; i++ {
		var publicKeyFragment = binary.LittleEndian.Uint64(pubKey[i*8 : (i+1)*8])
		for j := 0; j < 14; j++ {
			identity[i*14+j] = byte((publicKeyFragment % 26) + uint64(letter))
			publicKeyFragment /= 26
		}
	}

	h := k12.NewDraft10([]byte{})
	_, err := h.Write(pubKey[:])
	if err != nil {
		return "", errors.Wrap(err, "writing msg to k12")
	}

	var identityBytesChecksum [3]byte
	_, err = h.Read(identityBytesChecksum[:])
	if err != nil {
		return "", errors.Wrap(err, "reading hash from k12")
	}

	var identityBytesChecksumInt uint64
	identityBytesChecksumInt = uint64(identityBytesChecksum[0]) | (uint64(identityBytesChecksum[1]) << 8) | (uint64(identityBytesChecksum[2]) << 16)
	identityBytesChecksumInt &= 0x3FFFF

	for i := 0; i < 4; i++ {
		identity[56+i] = byte((identityBytesChecksumInt % 26) + uint64(letter))
		identityBytesChecksumInt /= 26
	}

	return Identity(identity[:]), nil
}

func (i *Identity) ToPubKey() ([32]byte, error) {
	var pubKey [32]byte

	if !isValidIdFormat(string(*i)) {
		return [32]byte{}, fmt.Errorf("invalid ID format")
	}

	idBytes := []byte(string(*i))

	if len(idBytes) != 60 {
		return [32]byte{}, fmt.Errorf("invalid ID length, expected 60, found %d", len(idBytes))
	}

	for i := 0; i < 4; i++ {
		for j := 13; j >= 0; j-- {
			if idBytes[i * 14 + j] < 'A' || idBytes[i * 14 + j] > 'Z'{
				return [32]byte{}, errors.New( "invalid conversion")
			}

			im := binary.LittleEndian.Uint64(pubKey[i*8 : (i+1)*8])
			im = im*26 + uint64(idBytes[i*14+j]-'A')
			imBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(imBytes, im)

			for k := 0; k < 8; k++ {
				pubKey[i*8+k] = imBytes[k]
			}
		}
	}

	return pubKey, nil
}

// isValidIdFormat checks if the provided string has a valid ID format.
func isValidIdFormat(idStr string) bool {
	for _, c := range idStr {
		if !(unicode.IsUpper(c) && unicode.IsLetter(c)) {
			return false
		}
	}
	return true
}
