package tx

type SignedTransaction struct {
	RawTx     []byte
	Signature [64]byte
}
