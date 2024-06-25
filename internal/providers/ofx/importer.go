package ofx

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"

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
	for i, txn := range txns {
		var description string
		if txn.Payee != nil {
			description = txn.Payee.Name.String()
		} else if txn.Name.String() != "" {
			description = txn.Name.String()
		} else if txn.Memo.String() != "" {
			description = txn.Memo.String()
		} else {
			err = fmt.Errorf("No available description for FITID %v", txn.FiTID.String())
			break
		}

		if description[:5] == "VISA-" {
			description = description[5:]
		}

		amt, err := strconv.ParseFloat(txn.TrnAmt.String(), 64)
		if err != nil {
			err = errors.Join(fmt.Errorf("Could not parse transaction amount for FITID %v: '%s'", txn.FiTID.String(), description), err)
			break
		}

		date := txn.DtPosted.Time

		var ref string

		if txn.CheckNum.String() != "" {
			ref = txn.CheckNum.String()
		} else if txn.RefNum.String() != "" {
			ref = txn.RefNum.String()
		} else {
			input := fmt.Sprintf("%v:%s:%.2f", date.Format("2006-01-02"), description, amt)
			ref = fmt.Sprintf("%x", md5.Sum([]byte(input)))
		}

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

	if err != nil {
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
