// Shared types for the importer program
package beanport

import "time"

// An import provider
type Provider = string

// A transaction that hasn't been confirmed
type PendingTransaction struct {
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

// The interface for a data importer
type Importer interface {
	Import() ([]*PendingTransaction, error)
}
