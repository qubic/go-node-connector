package types

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type LiveServiceClient struct {
	BaseUrl                      string
	TransactionBroadcastEndpoint string
	TickInfoEndpoint             string

	TickBroadcastOffset uint32
}

func LiveServiceClientWithDefaults() LiveServiceClient {
	return LiveServiceClient{
		BaseUrl:                      "https://rpc.qubic.org",
		TransactionBroadcastEndpoint: "/v1/broadcast-transaction",
		TickInfoEndpoint:             "/v1/tick-info",
		TickBroadcastOffset:          15,
	}
}

func NewLiveServiceClient(baseUrl, transactionBroadcastEndpoint, tickInfoEndpoint string, tickBroadcastOffset uint32) LiveServiceClient {
	return LiveServiceClient{
		BaseUrl:                      baseUrl,
		TransactionBroadcastEndpoint: transactionBroadcastEndpoint,
		TickInfoEndpoint:             tickInfoEndpoint,
		TickBroadcastOffset:          tickBroadcastOffset,
	}
}

func (lsc *LiveServiceClient) GetTickInfo() (*TickInfoResponse, error) {

	request, err := http.NewRequest(http.MethodGet, lsc.BaseUrl+lsc.TickInfoEndpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating tick info request")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "performing tick info request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, lsc.handleHttpError(response.Body)
	}

	var responseBody TickInfoResponse
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return nil, errors.Wrap(err, "decoding tick info response")
	}

	return &responseBody, nil
}

func (lsc *LiveServiceClient) GetScheduledTick() (uint32, error) {
	tickInfo, err := lsc.GetTickInfo()
	if err != nil {
		return 0, errors.Wrap(err, "getting current tick info")
	}

	return tickInfo.TickInfo.Tick + lsc.TickBroadcastOffset, nil

}

func (lsc *LiveServiceClient) BroadcastTransaction(tx Transaction) (*TransactionBroadcastResponse, error) {

	encodedTransaction, err := tx.EncodeToBase64()
	if err != nil {
		return nil, errors.Wrap(err, "encoding transaction")
	}

	requestPayload := TransactionBroadcastRequest{
		EncodedTransaction: encodedTransaction,
	}

	buff := new(bytes.Buffer)
	err = json.NewEncoder(buff).Encode(requestPayload)
	if err != nil {
		return nil, errors.Wrap(err, "encoding transaction broadcast payload")
	}

	request, err := http.NewRequest(http.MethodPost, lsc.BaseUrl+lsc.TransactionBroadcastEndpoint, buff)
	if err != nil {
		return nil, errors.Wrap(err, "creating transaction broadcast request")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "performing transaction broadcast request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, lsc.handleHttpError(response.Body)
	}

	var responseBody TransactionBroadcastResponse
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return nil, errors.Wrap(err, "decoding transaction broadcast response")
	}

	return &responseBody, nil
}

func (lsc *LiveServiceClient) handleHttpError(responseBody io.ReadCloser) error {

	data, err := io.ReadAll(responseBody)
	if err != nil {
		return errors.Wrap(err, "reading error body")
	}
	errorString := string(data)

	return errors.Errorf("response status not OK : %s", errorString)
}

type TransactionBroadcastRequest struct {
	EncodedTransaction string `json:"encodedTransaction"`
}

type TransactionBroadcastResponse struct {
	PeersBroadcasted   uint32 `json:"peersBroadcasted"`
	EncodedTransaction string `json:"encodedTransaction"`
}

type TickInfoResponse struct {
	TickInfo struct {
		Tick        uint32 `json:"tick"`
		Duration    uint32 `json:"duration"`
		Epoch       uint32 `json:"epoch"`
		InitialTick uint32 `json:"initialTick"`
	} `json:"tickInfo"`
}
