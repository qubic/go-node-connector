package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type LiveServiceClient struct {
	BaseUrl string
}

func NewLiveServiceClient(baseUrl string) LiveServiceClient {
	return LiveServiceClient{
		BaseUrl: baseUrl,
	}
}

func (lsc *LiveServiceClient) GetTickInfo() (*TickInfoResponse, error) {

	request, err := http.NewRequest(http.MethodGet, lsc.BaseUrl+"/v1/tick-info", nil)
	if err != nil {
		return nil, fmt.Errorf("creating tick info request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("performing tick info request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, lsc.handleHttpError(response.Body)
	}

	var responseBody TickInfoResponse
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return nil, fmt.Errorf("decoding tick info response: %w", err)
	}

	return &responseBody, nil
}

func (lsc *LiveServiceClient) BroadcastTransaction(tx Transaction) (*TransactionBroadcastResponse, error) {

	if tx.Signature == [64]byte{} {
		return nil, errors.New("cannot broadcast unsigned transaction")
	}

	encodedTransaction, err := tx.EncodeToBase64()
	if err != nil {
		return nil, fmt.Errorf("encoding transaction: %w", err)
	}

	requestPayload := TransactionBroadcastRequest{
		EncodedTransaction: encodedTransaction,
	}

	buff := new(bytes.Buffer)
	err = json.NewEncoder(buff).Encode(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("encoding transaction broadcast payload: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, lsc.BaseUrl+"/v1/broadcast-transaction", buff)
	if err != nil {
		return nil, fmt.Errorf("creating transaction broadcast request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("performing transaction broadcast request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, lsc.handleHttpError(response.Body)
	}

	var responseBody TransactionBroadcastResponse
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return nil, fmt.Errorf("decoding transaction broadcast response: %w", err)
	}

	return &responseBody, nil
}

func (lsc *LiveServiceClient) handleHttpError(responseBody io.Reader) error {

	data, err := io.ReadAll(responseBody)
	if err != nil {
		return fmt.Errorf("reading error body: %w", err)
	}
	errorString := string(data)

	return fmt.Errorf("response status not OK : %s", errorString)
}

type TransactionBroadcastRequest struct {
	EncodedTransaction string `json:"encodedTransaction"`
}

type TransactionBroadcastResponse struct {
	PeersBroadcasted   uint32 `json:"peersBroadcasted"`
	EncodedTransaction string `json:"encodedTransaction"`
	TransactionId      string `json:"transactionId"`
}

type TickInfoResponse struct {
	TickInfo struct {
		Tick        uint32 `json:"tick"`
		Duration    uint32 `json:"duration"`
		Epoch       uint32 `json:"epoch"`
		InitialTick uint32 `json:"initialTick"`
	} `json:"tickInfo"`
}
