package qubic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/qubic/go-node-connector/v2/types"
	"github.com/silenceper/pool"
)

type PoolConfig struct {
	InitialCap         int
	MaxCap             int
	MaxIdle            int
	IdleTimeout        time.Duration
	NodeFetcherUrl     string
	NodeFetcherTimeout time.Duration
	NodePort           string
}

func NewPoolConnection(config PoolConfig) (*Pool, error) {
	pcf := newPoolConnectionFactory(config.NodeFetcherTimeout, config.NodeFetcherUrl, config.NodePort)
	cfg := pool.Config{
		InitialCap: config.InitialCap,
		MaxIdle:    config.MaxIdle,
		MaxCap:     config.MaxCap,
		Factory:    pcf.Connect,
		Close:      pcf.Close,
		//The maximum idle time of the connection, the connection exceeding this time will be closed, which can avoid the problem of automatic failure when connecting to EOF when idle
		IdleTimeout: config.IdleTimeout,
	}
	chPool, err := pool.NewChannelPool(&cfg)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	p := Pool{chPool: chPool}

	return &p, nil
}

type Pool struct {
	chPool pool.Pool
}

func (p *Pool) Get() (*Client, error) {
	v, err := p.chPool.Get()
	if err != nil {
		return nil, fmt.Errorf("getting qubic pooled client connection: %w", err)
	}
	return v.(*Client), nil
}

func (p *Pool) Put(c *Client) error {
	err := p.chPool.Put(c)
	if err != nil {
		return fmt.Errorf("putting qubic pooled client connection: %w", err)
	}

	return nil
}

func (p *Pool) Close(c *Client) error {
	err := p.chPool.Close(c)
	if err != nil {
		return fmt.Errorf("closing qubic pool: %w", err)
	}

	return nil
}

type poolConnectionFactory struct {
	nodeFetcherTimeout time.Duration
	nodeFetcherUrl     string
	nodePort           string
}

func newPoolConnectionFactory(nodeFetcherTimeout time.Duration, nodeFetcherUrl string, nodePort string) *poolConnectionFactory {
	return &poolConnectionFactory{nodeFetcherTimeout: nodeFetcherTimeout, nodeFetcherUrl: nodeFetcherUrl, nodePort: nodePort}
}

func (pcf *poolConnectionFactory) Connect() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pcf.nodeFetcherTimeout)
	defer cancel()

	peer, err := pcf.getNewRandomPeer(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting new random peer: %w", err)
	}

	client, err := NewClient(ctx, peer, pcf.nodePort)
	if err != nil {
		return nil, fmt.Errorf("creating qubic client: %w", err)
	}

	fmt.Printf("connected to: %s\n", peer)
	return client, nil
}

func (pcf *poolConnectionFactory) Close(v interface{}) error { return v.(*Client).Close() }

type statusResponse struct {
	MaxTick          uint32         `json:"max_tick"`
	LastUpdate       int64          `json:"last_update"`
	ReliableNodes    []nodeResponse `json:"reliable_nodes"`
	MostReliableNode nodeResponse   `json:"most_reliable_node"`
}

type nodeResponse struct {
	Address    string            `json:"address"`
	Peers      types.PublicPeers `json:"peers"`
	LastTick   uint32            `json:"last_tick"`
	LastUpdate int64             `json:"last_update"`
}

func (pcf *poolConnectionFactory) getNewRandomPeer(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pcf.nodeFetcherUrl, nil)
	if err != nil {
		return "", fmt.Errorf("creating new request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("getting peers from node fetcher: %w", err)
	}

	var resp statusResponse
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", fmt.Errorf("unmarshalling response: %w", err)
	}

	peer := resp.ReliableNodes[rand.Intn(len(resp.ReliableNodes))]

	fmt.Printf("Got %d new peers. Selected random %s\n", len(resp.ReliableNodes), peer.Address)

	return peer.Address, nil
}
