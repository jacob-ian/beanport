package ofx

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aclindsa/ofxgo"
)

// Gets a reference for an OFX transaction
// The reference is designed to prevent ledger duplicates
func reference(txn ofxgo.Transaction, description string, amount float64) string {
	if txn.CheckNum.String() != "" {
		return txn.CheckNum.String()
	}
	if txn.RefNum.String() != "" {
		return txn.RefNum.String()
	}

	if ref := referenceFromMemo(txn.Memo.String()); ref != "" {
		return ref
	}

	input := fmt.Sprintf("%v:%s:%.2f", txn.DtPosted.Time.Format("2006-01-02"), description, amount)
	hashed := fmt.Sprintf("%x", md5.Sum([]byte(input)))
	encoded := base64.StdEncoding.EncodeToString([]byte(hashed))
	return encoded
}

// Attempts to find a ref in the transaction memo
func referenceFromMemo(memo string) string {
	if memo == "" {
		return ""
	}

	fields := strings.Fields(memo)
	if len(fields) == 0 {
		return ""
	}

	reference := ""
	for _, f := range fields {
		exists := strings.Contains(f, "Ref")
		if !exists {
			continue
		}
		reference = f
		break
	}

	return reference
}
