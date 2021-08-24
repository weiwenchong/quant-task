package util

import (
	"errors"
	"fmt"
)

func GetAssetCode(assetType int32, assetCode string) string {
	return fmt.Sprintf("%d_%s", assetType, assetCode)
}

func GetAssetHttpReqName(asset string) (string, error) {
	r := []rune(asset)
	if len(r) == 0 {
		return "", nil
	}
	if r[0] == '1' {
		r = append([]rune("SH"), r[2:]...)
	} else if r[0] == '2' {
		r = append([]rune("SZ"), r[2:]...)
	} else if r[0] == '3' {
		r = r[2:]
	} else if r[0] == '4' {
		r = r[2:]
	} else {
		return "", errors.New("invalid asset type")
	}
	return string(r), nil
}
