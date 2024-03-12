package qubic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/silenceper/pool"
	"io"
	"math/rand"
	"net/http"
	"time"
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
		return nil, errors.Wrap(err, "creating pool")
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
		return nil, errors.Wrap(err, "getting qubic pooled client connection")
	}
	return v.(*Client), nil
}

func (p *Pool) Put(c *Client) error {
	err := p.chPool.Put(c)
	if err != nil {
		return errors.Wrap(err, "putting qubic pooled client connection")
	}

	return nil
}

func (p *Pool) Close(c *Client) error {
	err := p.chPool.Close(c)
	if err != nil {
		return errors.Wrap(err, "closing qubic pool")
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
		return nil, errors.Wrap(err, "getting new random peer")
	}

	client, err := NewClient(ctx, peer, pcf.nodePort)
	if err != nil {
		return nil, errors.Wrap(err, "creating qubic client")
	}

	fmt.Printf("connected to: %s\n", peer)
	return client, nil
}

func (pcf *poolConnectionFactory) Close(v interface{}) error { return v.(*Client).Close() }

type response struct {
	Peers       []string `json:"peers"`
	Length      int      `json:"length"`
	LastUpdated int64    `json:"last_updated"`
}

func (pcf *poolConnectionFactory) getNewRandomPeer(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pcf.nodeFetcherUrl, nil)
	if err != nil {
		return "", errors.Wrap(err, "creating new request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "getting peers from node fetcher")
	}

	var resp response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "reading response body")
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", errors.Wrap(err, "unmarshalling response")
	}

	peer := resp.Peers[rand.Intn(len(resp.Peers))]

	fmt.Printf("Got %d new peers. Selected random %s\n", len(resp.Peers), peer)

	return peer, nil
}
