package fourq

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	eccFourq "github.com/cloudflare/circl/ecc/fourq"
	"github.com/cloudflare/circl/xof/k12"
	"github.com/cloudflare/fourq"
	"github.com/pkg/errors"
)

// verify checks the signature for a given public key and messageDigest.
func Verify(publicKey, messageDigest, signature []byte) error {
	pubKeyHex := hex.EncodeToString(publicKey)
	digestHex := hex.EncodeToString(messageDigest)
	sigHex := hex.EncodeToString(signature)

	_, _, _ = pubKeyHex, digestHex, sigHex

	if len(signature) != 64 {
		return errors.New("sig length is not 64")
	}

	pubKeyFixed, err := byteSliceToFixed32(publicKey)
	if err != nil {
		return errors.Wrap(err, "byte slice converting pubkey")
	}

	var A eccFourq.Point
	ok := A.Unmarshal(&pubKeyFixed)
	if !ok {
		return errors.New("encoding pubkey to point A")
	}

	if publicKey[15]&0x80 != 0 || signature[15]&0x80 != 0 || signature[62]&0xC0 != 0 || signature[63] != 0 {
		return errors.New("pubkey format validation failed")
	}
	temp := append(append(signature[:32], publicKey...), messageDigest...)

	h := k12.NewDraft10([]byte{}) // Using K12 for hashing, equivalent to KangarooTwelve(temp, 96, h, 64).
	_, err = h.Write(temp)
	if err != nil {
		return errors.Wrap(err, "k12 hashing")
	}

	var hash [64]byte
	_, err = h.Read(hash[:])
	if err != nil {
		return errors.Wrap(err, "reading k12 digest")
	}

	cutsig, err := byteSliceToFixed32(signature[32:])
	if err != nil {
		return errors.Wrap(err, "byte slice converting sig")
	}

	x := hex.EncodeToString(hash[:])
	_ = x

	//var p1 eccFourq.Point
	//p1.ScalarBaseMult(&cutsig)
	//
	//var p2 eccFourq.Point
	//p2.ScalarMult(&hash, &A)
	//
	//var addedP eccFourq.Point
	//addedP.Add(&p1, &p2)
	//

	r, err := eccMulDouble(cutsig, hash, A)
	if err != nil {
		return errors.Wrap(err, "mul double")
	}

	var encodedR [32]byte
	r.Marshal(&encodedR)

	if r.IsOnCurve() {
		fmt.Println("is on curve")
	}

	if subtle.ConstantTimeCompare(encodedR[:], signature[:32]) == 0 {
		return errors.New("sig and encoded comparison failed")
	}

	return nil
}

func eccMulDouble(k [32]byte, l [64]byte, Q eccFourq.Point) (eccFourq.Point, error) {
	var p1 eccFourq.Point
	p1.ScalarBaseMult(&k)

	var decodedQ [32]byte
	Q.Marshal(&decodedQ)

	p2Bytes, ok := fourq.ScalarMult(decodedQ, l[:], false)
	if !ok {
		return eccFourq.Point{}, errors.New("p2 mult failed")
	}

	var p2 eccFourq.Point
	p2.Unmarshal(&p2Bytes)

	var addedP eccFourq.Point
	addedP.Add(&p1, &p2)

	return addedP, nil
}

func byteSliceToFixed32(s []byte) ([32]byte, error) {
	if len(s) != 32 {
		return [32]byte{}, errors.New("slice length is less than 32")
	}
	var arr [32]byte

	copy(arr[:], s)

	return arr, nil
}
