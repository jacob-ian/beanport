package tui

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"jacobmatthe.ws/beanport/internal/beanport"
)

type session struct {
	file     *os.File
	Previous []*beanport.Transaction
}

func (sess *session) encodeLine(txn *beanport.Transaction) []byte {
	var encoded []byte
	base64.RawStdEncoding.Encode(encoded, []byte(fmt.Sprintf(
		"v1\t%s\t%s\t%.2f\t%s\t%s\t%s",
		txn.Date.Format("2006-01-06"),
		txn.Description,
		txn.Amount,
		txn.Reference,
		txn.Account,
		txn.OppositeAccount,
	)))
	encoded = append(encoded, '\n')
	return encoded
}

func (sess *session) decodeLine(line []byte) (*beanport.Transaction, error) {
	var decoded []byte
	_, err := base64.RawStdEncoding.Decode(decoded, line)
	if err != nil {
		return nil, err
	}

	str := string(decoded)
	if str[:1] != "v1" {
		return nil, errors.New("Mismatched session version")
	}

	parts := strings.Split("\t", str)

	date, err := time.Parse("2006-01-06", parts[1])
	if err != nil {
		return nil, errors.New("Could not parse date")
	}

	amt, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, errors.New("Could not parse amount")
	}

	return &beanport.Transaction{
		PendingTransaction: beanport.PendingTransaction{
			Date:        date,
			Description: parts[2],
			Amount:      amt,
			Reference:   parts[4],
			Account:     parts[5],
		},
		OppositeAccount: parts[6],
	}, nil
}

func (sess *session) Save(txn *beanport.Transaction) error {
	encoded := sess.encodeLine(txn)
	_, err := sess.file.Write(encoded)
	return err
}

func (app *Application) session() (*session, []*beanport.Transaction, error) {
	file, err := os.OpenFile(fmt.Sprintf("%s.tmp", app.outputFilePath), os.O_CREATE, os.ModeExclusive)
	if err != nil {
		return nil, nil, err
	}

	var contents []byte
	_, err = file.Read(contents)
	if err != nil {
		return nil, nil, err
	}

	session := &session{
		file: file,
	}

	lines := bytes.Split(contents, []byte{'\n'})
	if len(lines) == 0 {
		return session, nil, nil
	}

	var previous []*beanport.Transaction = make([]*beanport.Transaction, len(lines))
	for _, line := range lines {
		txn, err := session.decodeLine(line)
		if err != nil {
			app.logger.Error("Skipping transaction", "error", err.Error())
			continue
		}
		previous = append(previous, txn)
	}

	return session, previous, nil
}
