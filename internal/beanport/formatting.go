package beanport

import "fmt"

// Formats a pending transaction with a placeholder
// for the opposite account
func FormatPending(txn *PendingTransaction) string {
	return fmt.Sprintf(
		"%s * \"%s REF:%s\"\n\t%s\t%.2f %s\n\t{}\t%.2f %s",
		txn.Date.Format("2006-01-02"),
		txn.Description,
		txn.Reference,
		txn.Account,
		txn.Amount,
		txn.Commodity,
		-txn.Amount,
		txn.Commodity,
	)
}

// Formats a transaction for beancount
func FormatTransaction(txn *Transaction) string {
	return fmt.Sprintf(
		"%s * \"%s REF:%s\"\n\t%s\t%.2f %s\n\t%s\t%.2f %s\n\n",
		txn.Date.Format("2006-01-02"),
		txn.Description, txn.Reference,
		txn.Account,
		txn.Amount,
		txn.Commodity,
		txn.OppositeAccount,
		-txn.Amount,
		txn.Commodity,
	)
}
