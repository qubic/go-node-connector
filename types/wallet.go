package types

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/cloudflare/circl/ecc/fourq"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/pkg/errors"
	"math/big"
)

const seedLength = 55

type Wallet struct {
	PubKey   [32]byte
	PrivKey  [32]byte
	Identity Identity
}

func NewWallet(seed string) (Wallet, error) {
	privKey, err := getPrivateKey(seed)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting privKey")
	}

	pubKey, err := getPublicKey(privKey)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting pubkey")
	}

	var id Identity
	id, err = id.FromPubKey(pubKey, false)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting identity string")
	}

	return Wallet{
		PubKey:   pubKey,
		PrivKey:  privKey,
		Identity: id,
	}, nil
}

func NewDerivedWallet(seed string, index uint64) (Wallet, error) {
	subseed, err := GetSubSeed(seed)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting subseed")
	}

	derivedSubseed, err := GetDerivedSubseed(subseed, index)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting derived subseed")
	}

	privKey, err := getPrivateKeyFromSubseed(derivedSubseed)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting privKey")
	}

	pubKey, err := getPublicKey(privKey)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting pubkey")
	}

	var id Identity
	id, err = id.FromPubKey(pubKey, false)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting identity string")
	}

	return Wallet{
		PubKey:   pubKey,
		PrivKey:  privKey,
		Identity: id,
	}, nil
}

func getPrivateKey(seed string) ([32]byte, error) {
	subseed, err := GetSubSeed(seed)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "getting subseed")
	}

	h := k12.NewDraft10([]byte{})
	_, err = h.Write(subseed[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "writing msg to k12")
	}

	var privKey [32]byte
	_, err = h.Read(privKey[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading hash from k12")
	}

	return privKey, nil
}

func getPrivateKeyFromSubseed(subseed [32]byte) ([32]byte, error) {
	h := k12.NewDraft10([]byte{})
	_, err := h.Write(subseed[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "writing msg to k12")
	}

	var privKey [32]byte
	_, err = h.Read(privKey[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading hash from k12")
	}

	return privKey, nil
}

func getPublicKey(pk [32]byte) ([32]byte, error) {
	var p fourq.Point
	p.ScalarBaseMult(&pk)

	pubKey, err := encode(p)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "encoding fourq point to pubkey")
	}

	return pubKey, nil
}

func encode(p fourq.Point) ([32]byte, error) {
	x11 := new(big.Int).SetUint64(binary.LittleEndian.Uint64(p.X[1][8:]))
	temp1 := new(big.Int).And(x11, big.NewInt(0x4000000000000000))
	temp1.Lsh(temp1, 1)

	x01 := new(big.Int).SetUint64(binary.LittleEndian.Uint64(p.X[0][8:]))
	temp2 := new(big.Int).And(x01, big.NewInt(0x4000000000000000))
	temp2.Lsh(temp2, 1)

	var yBytes [32]byte
	copy(yBytes[:16], p.Y[0][:])
	copy(yBytes[16:], p.Y[1][:])

	pEncoded := yBytes

	var xBytes [32]byte
	copy(xBytes[:16], p.X[0][:])
	copy(xBytes[16:], p.X[1][:])
	if new(big.Int).SetBytes(xBytes[:]) == big.NewInt(0) {
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, temp1.Uint64())
		for i := 0; i < 8; i++ {
			pEncoded[3*8+i] |= bytes[i]
		}

		return pEncoded, nil
	}

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, temp2.Uint64())
	for i := 0; i < 8; i++ {
		pEncoded[3*8+i] |= bytes[i]
	}

	return pEncoded, nil
}

func GetSubSeed(seed string) ([32]byte, error) {
	if len(seed) != seedLength {
		return [32]byte{}, errors.Errorf("Invalid seed length. Expected %d, got: %d", seedLength, len(seed))
	}

	var seedBytes [seedLength]byte
	for i := 0; i < 55; i++ {
		seedBytes[i] = seed[i] - 'a'
	}

	h := k12.NewDraft10([]byte{})
	_, err := h.Write(seedBytes[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "writing msg to k12")
	}

	var subseed [32]byte
	_, err = h.Read(subseed[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading hash from k12")
	}

	return subseed, nil
}

func GetDerivedSubseed(subseed [32]byte, index uint64) ([32]byte, error) {
	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, index)

	h := k12.NewDraft10([]byte{})
	_, err := h.Write(subseed[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "writing subseed to k12")
	}

	var subseedHash [32]byte
	_, err = h.Read(subseedHash[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading subseed hash")
	}

	h = k12.NewDraft10([]byte{})
	_, err = h.Write(indexBytes)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "writing index to k12")
	}

	var indexHash [32]byte
	_, err = h.Read(indexHash[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading index hash")
	}

	h = k12.NewDraft10([]byte{})
	_, err = h.Write(append(subseedHash[:], indexHash[:]...))
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "writing combined hashes to k12")
	}

	var derivedSubseed [32]byte
	_, err = h.Read(derivedSubseed[:])
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "reading derived subseed from k12")
	}

	return derivedSubseed, nil
}

func GenerateRandomSeed() string {
	const charset = "abcdefghijklmnopqrstuvwxyz"

	var seed [55]byte

	for i := 0; i < 55; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		seed[i] = charset[randomIndex.Int64()]
	}

	return string(seed[:])
}
