package wallet

import (
	"encoding/binary"
	"fmt"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"unicode"
)

type QubicID struct {
	Data [32]byte
}

func NewQubicIDFromIdentity(identity string) (QubicID, error) {
	pubKey, err := fromIdentityString(identity)
	if err != nil {
		return QubicID{}, errors.Wrap(err, "getting pubkey from identity string")
	}

	return QubicID{Data: pubKey}, nil
}

func NewQubicID(pubKey [32]byte) QubicID {
	return QubicID{Data: pubKey}
}


func fromIdentityString(identity string) ([32]byte, error) {
	var buffer [32]byte

	if !isValidIdFormat(identity) {
		return [32]byte{}, fmt.Errorf("invalid ID format")
	}

	idBytes := []byte(identity)

	if len(idBytes) != 60 {
		return [32]byte{}, fmt.Errorf("invalid ID length, expected 60, found %d", len(idBytes))
	}

	for i := 0; i < 4; i++ {
		for j := 13; j >= 0; j-- {
			im := binary.LittleEndian.Uint64(buffer[i*8 : (i+1)*8])
			im = im*26 + uint64(idBytes[i*14+j]-'A')
			imBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(imBytes, im)

			for k := 0; k < 8; k++ {
				buffer[i*8+k] = imBytes[k]
			}
		}
	}

	return buffer, nil
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

func (qid *QubicID) GetIdentity() (string, error) {
	var identity [60]byte

	for i := 0; i < 4; i++ {
		var publicKeyFragment = binary.LittleEndian.Uint64(qid.Data[i*8 : (i+1)*8])
		for j := 0; j < 14; j++ {
			identity[i*14+j] = byte((publicKeyFragment % 26) + 'A')
			publicKeyFragment /= 26
		}
	}

	h := k12.NewDraft10([]byte{})
	_, err := h.Write(qid.Data[:])
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
		identity[56+i] = byte((identityBytesChecksumInt % 26) + 'A')
		identityBytesChecksumInt /= 26
	}

	return string(identity[:]), nil
}
