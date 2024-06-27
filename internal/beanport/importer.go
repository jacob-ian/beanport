// The shared types for beanport
package beanport

// An import provider
type Provider = string

// The interface for a data importer
type Importer interface {
	Import() ([]*PendingTransaction, error)
}
