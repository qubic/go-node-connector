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

	/* We don't read siblings as it doesn't seem to be implemented on node side yet. */
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

func (pd *PossessedAssetData) UnmarshallBinary(r io.Reader) error {

	err := binary.Read(r, binary.LittleEndian, &pd.PublicKey)
	if err != nil {
		return errors.Wrap(err, "reading asset data")
	}

	err = binary.Read(r, binary.LittleEndian, &pd.Type)
	if err != nil {
		return errors.Wrap(err, "reading asset type")
	}

	err = binary.Read(r, binary.LittleEndian, &pd.Padding)
	if err != nil {
		return errors.Wrap(err, "reading asset padding")
	}

	err = binary.Read(r, binary.LittleEndian, &pd.ManagingContractIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset managing contract index")
	}

	err = binary.Read(r, binary.LittleEndian, &pd.IssuanceIndex)
	if err != nil {
		return errors.Wrap(err, "reading asset issuance index")
	}

	err = binary.Read(r, binary.LittleEndian, &pd.NumberOfUnits)
	if err != nil {
		return errors.Wrap(err, "reading asset number of units")
	}

	err = pd.OwnedAsset.UnmarshallBinary(r)
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
