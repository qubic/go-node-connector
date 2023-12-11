package tx

const (
	REQUEST_TX_STATUS  = 201
	RESPONSE_TX_STATUS = 202
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
	Padding           [2]byte
	Digest            [32]byte
}


