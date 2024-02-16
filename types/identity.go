package types

import (
	"encoding/binary"
	"fmt"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"unicode"
)

const ArbitratorIdentity = "AFZPUAIYVPNUYGJRQVLUKOPPVLHAZQTGLYAAUUNBXFTVTAMSBKQBLEIEPCVJ"

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

func (i *Identity) ToPubKey(isLowerCase bool) ([32]byte, error) {
	letters := []byte{'A', 'Z'}
	if isLowerCase {
		letters = []byte{'a', 'z'}
	}

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
			if idBytes[i * 14 + j] < letters[0] || idBytes[i * 14 + j] > letters[1] {
				return [32]byte{}, errors.New( "invalid conversion")
			}

			im := binary.LittleEndian.Uint64(pubKey[i*8 : (i+1)*8])
			im = im*26 + uint64(idBytes[i*14+j]-letters[0])
			imBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(imBytes, im)

			for k := 0; k < 8; k++ {
				pubKey[i*8+k] = imBytes[k]
			}
		}
	}

	return pubKey, nil
}

func (i *Identity) String() string {
	if i == nil {
		return ""
	}
	return string(*i)
}

// isValidIdFormat checks if the provided string has a valid ID format.
func isValidIdFormat(idStr string) bool {
	for _, c := range idStr {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}
