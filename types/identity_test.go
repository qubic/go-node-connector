package types

import (
	"encoding/hex"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func TestGetIdentityFromPubkeyUpperCase(t *testing.T) {
	pubKey := [32]byte{230, 252, 58, 173, 75, 89, 77, 130, 191, 49, 3, 161, 16, 22, 216, 13, 232, 131, 222, 135, 59, 206, 196, 142, 144, 57, 98, 134, 80, 59, 38, 19}
	expectedIdentity := "QJRRSSKMJRDKUDTYVNYGAMQPULKAMILQQYOWBEXUDEUWQUMNGDHQYLOAJMEB"

	var ID Identity

	got, err := ID.FromPubKey(pubKey, false)
	if err != nil {
		t.Fatalf("Got err when getting identity key. err: %s", err.Error())
	}

	if cmp.Diff(string(got[:]), expectedIdentity) != "" {
		t.Fatalf("Mismatched return value. Expected: %s, got: %s", expectedIdentity, got)
	}
}

func TestGetIdentityFromPubkeyLowerCase(t *testing.T) {
	pubKey := [32]byte{230, 252, 58, 173, 75, 89, 77, 130, 191, 49, 3, 161, 16, 22, 216, 13, 232, 131, 222, 135, 59, 206, 196, 142, 144, 57, 98, 134, 80, 59, 38, 19}
	expectedIdentity := strings.ToLower("QJRRSSKMJRDKUDTYVNYGAMQPULKAMILQQYOWBEXUDEUWQUMNGDHQYLOAJMEB")

	var ID Identity

	got, err := ID.FromPubKey(pubKey, true)
	if err != nil {
		t.Fatalf("Got err when getting identity key. err: %s", err.Error())
	}

	if cmp.Diff(string(got[:]), expectedIdentity) != "" {
		t.Fatalf("Mismatched return value. Expected: %s, got: %s", expectedIdentity, got)
	}
}

func TestGetPubKeyFromIdentity(t *testing.T) {
	identity := "QJRRSSKMJRDKUDTYVNYGAMQPULKAMILQQYOWBEXUDEUWQUMNGDHQYLOAJMEB"
	expectedPubKey := [32]byte{230, 252, 58, 173, 75, 89, 77, 130, 191, 49, 3, 161, 16, 22, 216, 13, 232, 131, 222, 135, 59, 206, 196, 142, 144, 57, 98, 134, 80, 59, 38, 19}

	ID := Identity(identity)
	got, err := ID.ToPubKey()
	if err != nil {
		t.Fatalf("Got err when creating qubic id from identity. err: %s", err.Error())
	}
	if cmp.Diff(got, expectedPubKey) != "" {
		t.Fatalf("Mismatched return value. Expected: %s, got: %s", hex.EncodeToString(expectedPubKey[:]), hex.EncodeToString(got[:]))
	}
}
