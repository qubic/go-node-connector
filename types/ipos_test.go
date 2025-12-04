package types

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIpos_UnmarshallFromReader(t *testing.T) {
	header, err := base64.StdEncoding.DecodeString("FAAAQeKmLXg=")
	require.NoError(t, err)

	payload, err := base64.StdEncoding.DecodeString("KgAAAFRFU1QAAAAA")
	require.NoError(t, err)

	endResponseHeader, err := base64.StdEncoding.DecodeString("CAAAI/PqVtQ=")
	require.NoError(t, err)

	message := append(header, payload...)
	message = append(message, endResponseHeader...)

	var result Ipos
	err = result.UnmarshallFromReader(bytes.NewReader(message))
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 42, int(result[0].ContractIndex))
	assert.Equal(t, []byte("TEST\x00\x00\x00\x00"), result[0].AssetName[:])
}

func TestIpo_UnmarshallBinary(t *testing.T) {
	payload, err := base64.StdEncoding.DecodeString("KgAAAFRFU1QAAAAA")
	require.NoError(t, err)
	var result Ipo
	err = result.UnmarshallBinary(bytes.NewReader(payload))
	require.NoError(t, err)
	assert.Equal(t, 42, int(result.ContractIndex))
	assert.Equal(t, []byte("TEST\x00\x00\x00\x00"), result.AssetName[:])
}
