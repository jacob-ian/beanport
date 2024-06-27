// Shared types for the importer program
package beanport

import "time"

// A transaction that hasn't been confirmed
type PendingTransaction struct {
	Index       int
	Date        time.Time
	Description string
	Amount      float64
	Reference   string
	Account     string
	Commodity   string
}

// A transaction that has been assigned an account
type Transaction struct {
	PendingTransaction
	OppositeAccount string
}
