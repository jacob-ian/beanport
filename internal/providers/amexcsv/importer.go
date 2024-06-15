// An AMEX CSV importer
package amexcsv

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jacob-ian/beanport/internal/beanport"
)

const (
	Provider beanport.Provider = "amexcsv"
)

type ImporterConfig struct {
	Account         string
	NegativeAmounts bool
	Commodity       string
}

type amexImporter struct {
	data   []byte
	config *ImporterConfig
}

func (ai *amexImporter) Import() ([]*beanport.PendingTransaction, error) {
	columns := []string{
		"Date",
		"Date Processed",
		"Description",
		"Amount",
		"Foreign Spend Amount",
		"Commission",
		"Exchange Rate",
		"Additional Information",
		"Appears On Your Statement As",
		"Address",
		"Town/City",
		"Postcode",
		"Country",
		"Reference",
	}

	if len(ai.data) == 0 {
		return nil, errors.New("Empty input file")
	}

	var headerb []byte
	var data []byte
	for i, b := range ai.data {
		if b != '\n' {
			continue
		}
		headerb = ai.data[0:i]
		data = ai.data[i+1:]
		break
	}

	if bytes.Compare(headerb, []byte(strings.Join(columns, ","))) != 0 {
		return nil, errors.New("Header mismatch, please check columns")
	}

	var rows [][]byte
	rows = append(rows, []byte{})

	i := 0
	j := 0
	t := len(data) - 1
	quotes := false

	for t > 0 {
		t--

		// Create new row if matching \n__/__/____,
		if data[j] == '\n' && data[j+3] == '/' && data[j+6] == '/' && data[j+11] == ',' {
			rows = append(rows, []byte{})
			i++
			j++
			continue
		}

		// Characters to replace with a space
		if data[j] == '\n' || data[j] == '\t' {
			rows[i] = append(rows[i], ' ')
			j++
			continue
		}

		// Characters to remove
		if data[j] == '\'' {
			j++
			continue
		}

		// Check if there is a quote capture
		if data[j] == '"' {
			quotes = !quotes
			j++
			continue
		}

		// Don't add commas if in quote capture
		if quotes && data[j] == ',' {
			j++
			continue
		}

		rows[i] = append(rows[i], data[j])
		j++
	}

	var txns []*beanport.PendingTransaction
	var parseErr error
	for i, row := range rows {
		split := strings.Split(string(row), ",")
		date, err := time.Parse("02/01/2006", split[0])
		if err != nil {
			parseErr = err
			break
		}
		description := split[2]
		amount, err := strconv.ParseFloat(split[3], 64)
		if err != nil {
			parseErr = err
			break
		}
		reference := split[13]
		txns = append(txns, &beanport.PendingTransaction{
			Index:       i,
			Account:     ai.config.Account,
			Date:        date,
			Description: description,
			Amount:      0 - amount,
			Reference:   reference,
			Commodity:   ai.config.Commodity,
		})
	}

	if parseErr != nil {
		return nil, parseErr
	}

	return txns, nil
}

func NewImporter(data []byte, config *ImporterConfig) *amexImporter {
	return &amexImporter{
		data:   data,
		config: config,
	}
}
