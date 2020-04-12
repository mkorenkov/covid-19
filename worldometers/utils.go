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
	if dataItem != "" {
		result, err = strconv.ParseUint(trimCommas(dataItem), 10, 64)
		if err != nil {
			return result, errors.Wrapf(err, "failed to parse uint %v", dataItem)
		}
	}
	return result, nil
}

func parseFloat(dataItem string) (result float64, err error) {
	if dataItem != "" {
		result, err = strconv.ParseFloat(trimCommas(dataItem), 64)
		if err != nil {
			return result, errors.Wrapf(err, "failed to parse float %v", dataItem)
		}
	}
	return result, nil
}
