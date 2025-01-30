package example

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
)

// PerformSimpleTransaction Creates a simple transaction that sends funds from sourceAddress to destinationAddress and broadcasts it to the network.
func PerformSimpleTransaction(sourceAddress, sourceSeed, destinationAddress string, amount int64) (*types.TransactionBroadcastResponse, error) {

	// Get http live service client that uses the official Qubic RPC address.
	lsc := types.LiveServiceClientWithDefaults()
	// Specify how many ticks ahead of the current tick we would like to schedule the transaction.
	lsc.TickBroadcastOffset = 10

	// The service address may be changed in case usage of on-premise / alternative infrastructure is desired.
	//lsc.BaseUrl = "https://your-service.com"

	// Request the current network tick and add the specified offset.
	scheduledTick, err := lsc.GetScheduledTick()
	if err != nil {
		return nil, errors.Wrap(err, "obtaining scheduled tick number")
	}

	// Create the transaction.
	tx, err := types.NewSimpleTransferTransaction(sourceAddress, destinationAddress, amount, scheduledTick)
	if err != nil {
		return nil, errors.Wrap(err, "creating simple transaction")
	}

	// Sign the transaction.
	err = types.DefaultSigner.SignTx(&tx, sourceSeed)
	if err != nil {
		return nil, errors.Wrap(err, "signing transaction")
	}

	// Broadcast the transaction and obtain the response.
	response, err := lsc.BroadcastTransaction(tx)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting transaction")
	}

	// Get the transaction ID for logging purposes
	txId, err := tx.ID()
	if err != nil {
		return nil, errors.Wrap(err, "obtaining transaction id")
	}

	// Log the transaction ID, how many Qubic peers the transaction was broadcast to, and the scheduled tick.
	fmt.Printf("Broadcasted transaction %s to %d peers. Scheduled for execution on tick %d.\n", txId, response.PeersBroadcasted, scheduledTick)

	return response, nil
}

// PerformSendManyTransaction Creates a send-many transaction that sends funds from the sourceAddress to the recipients specified by sendManyTransfers.
func PerformSendManyTransaction(sourceAddress, sourceSeed string, sendManyTransfers []types.SendManyTransfer) (*types.TransactionBroadcastResponse, error) {

	// Get http live service client that uses the official Qubic RPC address.
	lsc := types.LiveServiceClientWithDefaults()
	// Specify how many ticks ahead of the current tick we would like to schedule the transaction.
	lsc.TickBroadcastOffset = 10

	// The service address may be changed in case usage of on-premise / alternative infrastructure is desired.
	//lsc.BaseUrl = "https://your-service.com"

	// Request the current network tick and add the specified offset.
	scheduledTick, err := lsc.GetScheduledTick()
	if err != nil {
		return nil, errors.Wrap(err, "obtaining scheduled tick number")
	}

	// Create the send-many payload. This specifies to whom to send funds and how much they receive.
	var sendManyPayload types.SendManyTransferPayload
	err = sendManyPayload.AddTransfers(sendManyTransfers)
	if err != nil {
		return nil, errors.Wrap(err, "adding transfers to send many payload")
	}

	// Create the send-many transaction.
	// Note that send-many transactions require the payment of the SC fee. This is added automatically when creating the transaction.
	tx, err := types.NewSendManyTransferTransaction(sourceAddress, scheduledTick, sendManyPayload)
	if err != nil {
		return nil, errors.Wrap(err, "creating send many transaction")
	}

	// Sign the transaction.
	err = types.DefaultSigner.SignTx(&tx, sourceSeed)
	if err != nil {
		return nil, errors.Wrap(err, "signing transaction")
	}

	// Broadcast the transaction and obtain the response.
	response, err := lsc.BroadcastTransaction(tx)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting transaction")
	}

	// Get the transaction ID for logging purposes
	txId, err := tx.ID()
	if err != nil {
		return nil, errors.Wrap(err, "obtaining transaction id")
	}

	// Log the transaction ID, how many Qubic peers the transaction was broadcast to, and the scheduled tick.
	fmt.Printf("Broadcasted transaction %s to %d peers. Scheduled for execution on tick %d.\n", txId, response.PeersBroadcasted, scheduledTick)

	return response, nil
}
