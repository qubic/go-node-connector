package wallet

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
	PubKey [32]byte
	PrivKey [32]byte
	Identity string
}

func New(seed string) (Wallet, error) {
	privKey, err := getPrivateKey(seed)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting privKey")
	}

	pubKey, err := getPublicKey(privKey)
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting pubkey")
	}

	id := NewQubicID(pubKey)
	identity, err := id.GetIdentity()
	if err != nil {
		return Wallet{}, errors.Wrap(err, "getting identity string")
	}

	return Wallet{
		PubKey:   pubKey,
		PrivKey:  privKey,
		Identity: identity,
	}, nil
}

func getPrivateKey(seed string) ([32]byte, error) {
	subseed, err := getSubSeed(seed)
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
			pEncoded[3*8 + i] |= bytes[i]
		}

		return pEncoded, nil
	}

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, temp2.Uint64())
	for i := 0; i < 8; i++ {
		pEncoded[3*8 + i] |= bytes[i]
	}

	return pEncoded, nil
}

func getSubSeed(seed string) ([32]byte, error) {
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


