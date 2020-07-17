package libstoragemgmt

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	errors "github.com/libstorage/libstoragemgmt-golang/errors"
)

// udsPath returns the lsm unix domain file path
func udsPath() string {
	if p := os.Getenv(udsPathVarName); len(p) > 0 {
		return p
	}
	return udsPathDefault
}

func contains(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}

func emptySliceIfNil(provided []string) []string {
	if provided != nil {
		return provided
	}
	return make([]string, 0)
}

func handleSnapshotOptArg(args map[string]interface{}, ss *FileSystemSnapShot) {

	if ss != nil {
		args["snapshot"] = *ss
	} else {
		args["snapshot"] = nil
	}
}

func handleSearch(args map[string]interface{}, search []string) bool {
	rc := true

	switch num := len(search); num {
	case 0:
		args["search_key"] = nil
		args["search_value"] = nil
	case 2:
		args["search_key"] = search[0]
		args["search_value"] = search[1]
	default:
		rc = false
	}
	return rc
}

func validateInitID(initID string, initType InitiatorType) error {
	if initType == InitiatorTypeWwpn {
		matched, err := regexp.Match("^(0x|0X)?([0-9A-Fa-f]{2})(([\\.:\\-])?[0-9A-Fa-f]{2}){7}$", []byte(initID))
		if err != nil {
			return err
		}
		if !matched {
			return &errors.LsmError{
				Code: errors.InvalidArgument,
				Message: fmt.Sprintf(
					"initID invalid for InitiatorTypeWwpn: %s", initID)}
		}

	} else if initType == InitiatorTypeIscsiIqn {
		if !strings.HasPrefix(initID, "iqn") && !strings.HasPrefix(initID, "eui") && !strings.HasPrefix(initID, "naa") {
			return &errors.LsmError{
				Code: errors.InvalidArgument,
				Message: fmt.Sprintf(
					"initID invalid for InitiatorTypeIscsiIqn: %s", initID)}
		}

	} else {
		return &errors.LsmError{
			Code:    errors.InvalidArgument,
			Message: fmt.Sprintf("invalid initType: %d", initType)}
	}
	return nil
}
