package ofx

import (
	"fmt"
	"strings"

	"github.com/aclindsa/ofxgo"
)

// Gets a description from an OFX transaction
func description(txn ofxgo.Transaction) (string, error) {
	if txn.Payee != nil {
		return fmt.Sprintf("%s %s", txn.Payee.Name.String(), txn.Payee.State), nil
	}
	if txn.Name.String() != "" {
		return txn.Name.String(), nil
	}
	if txn.Memo.String() != "" {
		return descriptionFromMemo(txn.Memo.String()), nil
	}
	return "", fmt.Errorf("No available description properties for FITID %v", txn.FiTID.String())
}

// Convert a bank transaction memo to a description
func descriptionFromMemo(memo string) string {
	description := memo
	if description[:5] == "VISA-" {
		description = description[5:]
	}

	fields := strings.Fields(description)
	if len(fields) == 0 {
		return description
	}

	var output []string
	for _, f := range fields {
		if strings.Contains(f, "Ref") {
			continue
		}
		if strings.Contains(f, "Apple Pay") {
			continue
		}
		output = append(output, f)
	}

	return strings.Join(output, " ")
}
