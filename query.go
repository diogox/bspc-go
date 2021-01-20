package bspc

import (
	"encoding/json"
	"fmt"
	"strings"
)

type queryResponseResolver func(payload []byte) error

// TODO: godoc
// TODO: restrict this somehow so that it only accepts structs?
//   Maybe I can have an interface implemented by all the package's types, and that is required here.
func ToStruct(res interface{}) queryResponseResolver {
	return func(payload []byte) error {
		if err := json.Unmarshal(payload, &res); err != nil {
			return err
		}

		return nil
	}
}

func ToIDSlice(res *[]ID) queryResponseResolver {
	return func(payload []byte) error {
		lines := strings.Split(string(payload), "\n")
		for _, l := range lines {
			if l == "" {
				continue
			}

			id, err := hexToID(l)
			if err != nil {
				return fmt.Errorf("failed to convert hex iD into ID type: %v", err)
			}

			*res = append(*res, id)
		}

		return nil
	}
}