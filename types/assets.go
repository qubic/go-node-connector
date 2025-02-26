package types

import (
	"encoding/binary"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "reading asset tick")
	}

	err = binary.Read(r, binary.LittleEndian, &ai.UniverseIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset universe index")
	}

	err = binary.Read(r, binary.LittleEndian, &ai.Siblings)
	if err != nil {
		return errors.Wrap(err, "reading asset siblings")
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
			return errors.Wrap(err, "reading header")
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != IssuedAssetsResponse {
			return errors.Errorf("Invalid header type, expected %d, found %d", IssuedAssetsResponse, header.Type)
		}

		var issuedAssetData IssuedAssetData
		err = issuedAssetData.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "unmarshalling issued asset data")
		}

		var assetInfo AssetInfo
		err = assetInfo.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "reading issued asset info")
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
		return errors.Wrap(err, "reading issued asset public key")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return errors.Wrap(err, "reading issued asset type")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Name)
	if err != nil {
		return errors.Wrap(err, "reading issued asset name")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfDecimalPlaces)
	if err != nil {
		return errors.Wrap(err, "reading issued asset number of decimal places")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.UnitOfMeasurement)
	if err != nil {
		return errors.Wrap(err, "reading issued asset unit of measurement")
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
			return errors.Wrap(err, "reading header")
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != PossessedAssetsResponse {
			return errors.Errorf("Invalid header type, expected %d, found %d", PossessedAssetsResponse, header.Type)
		}

		var possessedAssetData PossessedAssetData
		err = possessedAssetData.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "unmarshalling possessed asset data")
		}

		var assetInfo AssetInfo
		err = assetInfo.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "reading possessed asset info")
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
		return errors.Wrap(err, "reading asset data")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return errors.Wrap(err, "reading asset type")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Padding)
	if err != nil {
		return errors.Wrap(err, "reading asset padding")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.ManagingContractIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset managing contract index")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.IssuanceIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset issuance index")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfUnits)
	if err != nil {
		return errors.Wrap(err, "reading asset number of units")
	}

	err = ad.OwnedAsset.UnmarshallBinary(r)
	if err != nil {
		return errors.Wrap(err, "reading owned asset")
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
			return errors.Wrap(err, "reading header")
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != OwnedAssetsResponse {
			return errors.Errorf("Invalid header type, expected %d, found %d", OwnedAssetsResponse, header.Type)
		}

		var ownedAssetData OwnedAssetData
		err = ownedAssetData.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "unmarshalling owned asset data")
		}

		var assetInfo AssetInfo
		err = assetInfo.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "reading owned asset info")
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
		return errors.Wrap(err, "reading asset data")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Type)
	if err != nil {
		return errors.Wrap(err, "reading asset type")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.Padding)
	if err != nil {
		return errors.Wrap(err, "reading asset padding")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.ManagingContractIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset managing contract index")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.IssuanceIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset issuance index")
	}

	err = binary.Read(r, binary.LittleEndian, &ad.NumberOfUnits)
	if err != nil {
		return errors.Wrap(err, "reading asset number of units")
	}

	err = ad.IssuedAsset.UnmarshallBinary(r)
	if err != nil {
		return errors.Wrap(err, "reading issued asset")
	}

	return nil
}

// TODO new asset requests are very similar to old one but typically have one property less. How to handle this?

// issuance

type AssetIssuance struct {
	Asset         IssuedAssetData
	Tick          uint32
	UniverseIndex uint32
}

type AssetIssuances []AssetIssuance

func (ia *AssetIssuances) UnmarshallFromReader(r io.Reader) error {
	for {
		var header RequestResponseHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return errors.Wrap(err, "reading header")
		}

		if header.Type == EndResponse {
			break
		}

		if header.Type != RespondAssets {
			return errors.Errorf("Invalid header type, expected %d, found %d", RespondAssets, header.Type)
		}

		var issuedAssetData IssuedAssetData
		err = issuedAssetData.UnmarshallBinary(r)
		if err != nil {
			return errors.Wrap(err, "unmarshalling issued asset data")
		}

		var tick uint32
		err = binary.Read(r, binary.LittleEndian, &tick)
		if err != nil {
			return errors.Wrap(err, "reading asset tick")
		}

		var universeIndex uint32
		err = binary.Read(r, binary.LittleEndian, &universeIndex)
		if err != nil {
			return errors.Wrap(err, "reading asset universe index")
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

// ownership

// IssuedAssetData can be reused. OwnedAssetData and PossessedAssetData have incorrect structure

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
	tick          uint32
	universeIndex uint32
}

type AssetOwnerships []AssetOwnership

type AssetPossessionData struct {
	PublicKey             [32]byte
	Type                  byte
	Padding               [1]int8
	ManagingContractIndex uint16
	IssuanceIndex         uint32
	NumberOfUnits         int64
}

type AssetPossession struct {
	Asset         AssetPossessionData
	tick          uint32
	universeIndex uint32
}

type AssetPossessions []AssetPossession
