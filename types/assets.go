package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	AssetsDepth = 24
)

type AssetInfo struct {
	Tick          uint32
	UniverseIndex uint32
	Siblings      [AssetsDepth][32]byte
}

func (ai *AssetInfo) UnmarshallBinary(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &ai.Tick)
	if err != nil {
		return fmt.Errorf("reading asset tick: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ai.UniverseIndex)
	if err != nil {
		return fmt.Errorf("reading asset universe index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ai.Siblings)
	if err != nil {
		return fmt.Errorf("reading asset siblings: %w", err)
	}

	return nil
}

/* Issued asset */

type IssuedAssetData struct {
	PublicKey             [32]byte
	Type                  byte
	Name                  [7]int8
	NumberOfDecimalPlaces int8
	UnitOfMeasurement     [7]int8
}

type IssuedAsset struct {
	Data IssuedAssetData
	Info AssetInfo
}

type IssuedAssets []IssuedAsset

func (ia *IssuedAssets) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != IssuedAssetsResponse {
			return fmt.Errorf("Invalid header type, expected %d, found %d", IssuedAssetsResponse, header.Type)
		}

		var issuedAssetData IssuedAssetData
		err = issuedAssetData.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling issued asset data: %w", err)
		}

		var assetInfo AssetInfo
		err = assetInfo.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("reading issued asset info: %w", err)
		}

		issuedAsset := IssuedAsset{
			Data: issuedAssetData,
			Info: assetInfo,
		}

		*ia = append(*ia, issuedAsset)
	}

	return nil
}

func (ad *IssuedAssetData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ad.PublicKey)
	if err != nil {
		return fmt.Errorf("reading issued asset public key: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return fmt.Errorf("reading issued asset type: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Name)
	if err != nil {
		return fmt.Errorf("reading issued asset name: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfDecimalPlaces)
	if err != nil {
		return fmt.Errorf("reading issued asset number of decimal places: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.UnitOfMeasurement)
	if err != nil {
		return fmt.Errorf("reading issued asset unit of measurement: %w", err)
	}
	return nil
}

/* Possessed asset */

type PossessedAssetData struct {
	PublicKey             [32]byte
	Type                  byte
	Padding               [1]int8
	ManagingContractIndex uint16
	IssuanceIndex         uint32
	NumberOfUnits         int64
	OwnedAsset            OwnedAssetData
}

type PossessedAsset struct {
	Data PossessedAssetData
	Info AssetInfo
}

type PossessedAssets []PossessedAsset

func (pa *PossessedAssets) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != PossessedAssetsResponse {
			return fmt.Errorf("Invalid header type, expected %d, found %d", PossessedAssetsResponse, header.Type)
		}

		var possessedAssetData PossessedAssetData
		err = possessedAssetData.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling possessed asset data: %w", err)
		}

		var assetInfo AssetInfo
		err = assetInfo.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("reading possessed asset info: %w", err)
		}

		possessedAsset := PossessedAsset{
			Data: possessedAssetData,
			Info: assetInfo,
		}

		*pa = append(*pa, possessedAsset)
	}

	return nil
}

func (ad *PossessedAssetData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ad.PublicKey)
	if err != nil {
		return fmt.Errorf("reading asset data: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return fmt.Errorf("reading asset type: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Padding)
	if err != nil {
		return fmt.Errorf("reading asset padding: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.ManagingContractIndex)
	if err != nil {
		return fmt.Errorf("reading asset managing contract index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.IssuanceIndex)
	if err != nil {
		return fmt.Errorf("reading asset issuance index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfUnits)
	if err != nil {
		return fmt.Errorf("reading asset number of units: %w", err)
	}

	err = ad.OwnedAsset.UnmarshallBinary(r)
	if err != nil {
		return fmt.Errorf("reading owned asset: %w", err)
	}

	return nil
}

/* Owned Asset */

type OwnedAssetData struct {
	PublicKey             [32]byte
	Type                  byte
	Padding               [1]int8
	ManagingContractIndex uint16
	IssuanceIndex         uint32
	NumberOfUnits         int64
	IssuedAsset           IssuedAssetData
}

type OwnedAsset struct {
	Data OwnedAssetData
	Info AssetInfo
}

type OwnedAssets []OwnedAsset

func (oa *OwnedAssets) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != OwnedAssetsResponse {
			return fmt.Errorf("Invalid header type, expected %d, found %d", OwnedAssetsResponse, header.Type)
		}

		var ownedAssetData OwnedAssetData
		err = ownedAssetData.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling owned asset data: %w", err)
		}

		var assetInfo AssetInfo
		err = assetInfo.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("reading owned asset info: %w", err)
		}

		ownedAsset := OwnedAsset{
			Data: ownedAssetData,
			Info: assetInfo,
		}

		*oa = append(*oa, ownedAsset)
	}

	return nil
}

func (ad *OwnedAssetData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ad.PublicKey)
	if err != nil {
		return fmt.Errorf("reading asset public key: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return fmt.Errorf("reading asset type: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Padding)
	if err != nil {
		return fmt.Errorf("reading asset padding: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.ManagingContractIndex)
	if err != nil {
		return fmt.Errorf("reading asset managing contract index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.IssuanceIndex)
	if err != nil {
		return fmt.Errorf("reading asset issuance index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfUnits)
	if err != nil {
		return fmt.Errorf("reading asset number of units: %w", err)
	}

	err = ad.IssuedAsset.UnmarshallBinary(r)
	if err != nil {
		return fmt.Errorf("reading issued asset: %w", err)
	}

	return nil
}

// issuance

type AssetIssuanceData struct {
	PublicKey             [32]byte
	Type                  byte
	Name                  [7]int8
	NumberOfDecimalPlaces int8
	UnitOfMeasurement     [7]int8
}

type AssetIssuance struct {
	Asset         AssetIssuanceData // can be reused
	Tick          uint32
	UniverseIndex uint32
}

type AssetIssuances []AssetIssuance

func (ia *AssetIssuances) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != RespondAssets {
			return fmt.Errorf("Invalid header type, expected %d, found %d", RespondAssets, header.Type)
		}

		var issuedAssetData AssetIssuanceData
		err = issuedAssetData.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling issued asset data: %w", err)
		}

		var tick uint32
		err = binary.Read(r, binary.LittleEndian, &tick)
		if err != nil {
			return fmt.Errorf("reading asset tick: %w", err)
		}

		var universeIndex uint32
		err = binary.Read(r, binary.LittleEndian, &universeIndex)
		if err != nil {
			return fmt.Errorf("reading asset universe index: %w", err)
		}

		issuedAsset := AssetIssuance{
			Asset:         issuedAssetData,
			Tick:          tick,
			UniverseIndex: universeIndex,
		}

		*ia = append(*ia, issuedAsset)
	}

	return nil
}

func (ad *AssetIssuanceData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ad.PublicKey)
	if err != nil {
		return fmt.Errorf("reading issued asset public key: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return fmt.Errorf("reading issued asset type: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Name)
	if err != nil {
		return fmt.Errorf("reading issued asset name: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfDecimalPlaces)
	if err != nil {
		return fmt.Errorf("reading issued asset number of decimal places: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.UnitOfMeasurement)
	if err != nil {
		return fmt.Errorf("reading issued asset unit of measurement: %w", err)
	}
	return nil
}

// ownership

type AssetOwnershipData struct {
	PublicKey             [32]byte
	Type                  byte
	Padding               [1]int8
	ManagingContractIndex uint16
	IssuanceIndex         uint32
	NumberOfUnits         int64
}

type AssetOwnership struct {
	Asset         AssetOwnershipData
	Tick          uint32
	UniverseIndex uint32
}

type AssetOwnerships []AssetOwnership

func (oa *AssetOwnerships) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != RespondAssets {
			return fmt.Errorf("Invalid header type, expected %d, found %d", RespondAssets, header.Type)
		}

		var assetOwnershipData AssetOwnershipData
		err = assetOwnershipData.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling owned asset data: %w", err)
		}

		var tick uint32
		err = binary.Read(r, binary.LittleEndian, &tick)
		if err != nil {
			return fmt.Errorf("reading asset tick: %w", err)
		}

		var universeIndex uint32
		err = binary.Read(r, binary.LittleEndian, &universeIndex)
		if err != nil {
			return fmt.Errorf("reading asset universe index: %w", err)
		}

		assetOwnership := AssetOwnership{
			Asset:         assetOwnershipData,
			Tick:          tick,
			UniverseIndex: universeIndex,
		}

		*oa = append(*oa, assetOwnership)
	}

	return nil
}

func (ad *AssetOwnershipData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ad.PublicKey)
	if err != nil {
		return fmt.Errorf("reading asset public key: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return fmt.Errorf("reading asset type: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Padding)
	if err != nil {
		return fmt.Errorf("reading asset padding: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.ManagingContractIndex)
	if err != nil {
		return fmt.Errorf("reading asset managing contract index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.IssuanceIndex)
	if err != nil {
		return fmt.Errorf("reading asset issuance index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfUnits)
	if err != nil {
		return fmt.Errorf("reading asset number of units: %w", err)
	}

	return nil
}

type AssetPossessionData struct {
	PublicKey             [32]byte
	Type                  byte
	Padding               [1]int8
	ManagingContractIndex uint16
	OwnershipIndex        uint32
	NumberOfUnits         int64
}

type AssetPossession struct {
	Asset         AssetPossessionData
	Tick          uint32
	UniverseIndex uint32
}

type AssetPossessions []AssetPossession

func (pa *AssetPossessions) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return fmt.Errorf("reading header: %w", err)
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != RespondAssets {
			return fmt.Errorf("Invalid header type, expected %d, found %d", RespondAssets, header.Type)
		}

		var possessedAssetData AssetPossessionData
		err = possessedAssetData.UnmarshallBinary(r)
		if err != nil {
			return fmt.Errorf("unmarshalling possessed asset data: %w", err)
		}

		var tick uint32
		err = binary.Read(r, binary.LittleEndian, &tick)
		if err != nil {
			return fmt.Errorf("reading asset tick: %w", err)
		}

		var universeIndex uint32
		err = binary.Read(r, binary.LittleEndian, &universeIndex)
		if err != nil {
			return fmt.Errorf("reading asset universe index: %w", err)
		}

		possessedAsset := AssetPossession{
			Asset:         possessedAssetData,
			Tick:          tick,
			UniverseIndex: universeIndex,
		}

		*pa = append(*pa, possessedAsset)
	}

	return nil
}

func (ad *AssetPossessionData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &ad.PublicKey)
	if err != nil {
		return fmt.Errorf("reading asset public key: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return fmt.Errorf("reading asset type: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Padding)
	if err != nil {
		return fmt.Errorf("reading asset padding: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.ManagingContractIndex)
	if err != nil {
		return fmt.Errorf("reading asset managing contract index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.OwnershipIndex)
	if err != nil {
		return fmt.Errorf("reading asset ownership index: %w", err)
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfUnits)
	if err != nil {
		return fmt.Errorf("reading asset number of units: %w", err)
	}

	return nil
}
