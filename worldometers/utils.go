package worldometers

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func trimCommas(v string) string {
	return strings.ReplaceAll(v, ",", "")
}

func parseUint(dataItem string) (result uint64, err error) {
	res := strings.TrimSpace(dataItem)
	if dataItem == "" || dataItem == "N/A" {
		return result, nil
	}
	result, err = strconv.ParseUint(trimCommas(res), 10, 64)
	if err != nil {
		return result, errors.Wrapf(err, "failed to parse uint %v", res)
	}
	return result, nil
}

func parseFloat(dataItem string) (result float64, err error) {
	res := strings.TrimSpace(dataItem)
	if dataItem == "" || dataItem == "N/A" {
		return result, nil
	}
	result, err = strconv.ParseFloat(trimCommas(res), 64)
	if err != nil {
		return result, errors.Wrapf(err, "failed to parse float %v", res)
	}
	return result, nil
}
