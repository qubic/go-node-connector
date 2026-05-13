package types

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type PublicPeers []string

func (pp *PublicPeers) UnmarshallFromReader(r io.Reader) error {
	var header RequestResponseHeader
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	if header.Type != ExchangePublicPeers {
		return fmt.Errorf("Invalid header type, expected %d, found %d", ExchangePublicPeers, header.Type)
	}

	var peers [4][4]byte

	err = binary.Read(r, binary.LittleEndian, &peers)
	if err != nil {
		return fmt.Errorf("reading public peers from reader: %w", err)
	}

	for _, peer := range peers {
		if peer == [4]byte{} {
			continue
		}
		ip := net.IP(peer[:])
		if ip == nil {
			continue
		}

		*pp = append(*pp, ip.String())
	}

	var nextHeader RequestResponseHeader
	err = binary.Read(r, binary.BigEndian, &nextHeader)
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	ignoredBytes := make([]byte, nextHeader.GetSize()-uint32(binary.Size(nextHeader)))
	_, err = r.Read(ignoredBytes)
	if err != nil {
		return fmt.Errorf("reading ignored bytes: %w", err)
	}

	return nil
}

func ipBytesToString(ip [4]byte) string {
	return string(rune(ip[0])) + "." + string(rune(ip[1])) + "." + string(rune(int(ip[2]))) + "." + string(rune(ip[3]))
}
