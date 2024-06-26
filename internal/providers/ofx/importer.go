package ofx

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aclindsa/ofxgo"
	"github.com/jacob-ian/beanport/internal/beanport"
)

const (
	Provider beanport.Provider = "ofx"
)

type ImporterConfig struct {
	Commodity string
	Account   string
}

type Importer struct {
	data      []byte
	commodity string
	account   string
}

func (imp *Importer) Import() ([]*beanport.PendingTransaction, error) {
	res, err := ofxgo.ParseResponse(bytes.NewReader(imp.data))
	if err != nil {
		return nil, errors.Join(errors.New("Couldn't parse OFX"), err)
	}

	if len(res.Bank) != 1 {
		return nil, errors.New("OFX error: too many bank messages")
	}

	root, ok := res.Bank[0].(*ofxgo.StatementResponse)
	if !ok {
		return nil, errors.Join(errors.New("OFX error: not a statement response"), err)
	}

	txns := root.BankTranList.Transactions
	var output []*beanport.PendingTransaction
	var parseErr error

	for i, txn := range txns {
		description, err := description(txn)
		if err != nil {
			parseErr = errors.Join(fmt.Errorf("Could not generate description"), err)
			break
		}

		amt, err := strconv.ParseFloat(txn.TrnAmt.String(), 64)
		if err != nil {
			parseErr = errors.Join(fmt.Errorf("Could not parse transaction amount for FITID %v: '%s'", txn.FiTID.String(), description), err)
			break
		}

		if amt == 0.00 {
			continue
		}

		ref := reference(txn, description, amt)

		output = append(output, &beanport.PendingTransaction{
			Index:       i,
			Account:     imp.account,
			Date:        txn.DtPosted.Time,
			Description: description,
			Amount:      amt,
			Reference:   ref,
			Commodity:   imp.commodity,
		})
	}

	if parseErr != nil {
		return nil, errors.Join(errors.New("Could not read transactions"), err)
	}

	return output, nil
}

func NewImporter(data []byte, config *ImporterConfig) *Importer {
	return &Importer{
		data:      data,
		commodity: config.Commodity,
		account:   config.Account,
	}
}
