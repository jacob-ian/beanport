# Beanport!

A CLI tool to import transaction data into a beancount ledger, written in Go.

## Financial Institution Support

| Institution | Provider |
| ----------- | -------- |
| Amex (CSV)  | amexcsv  |

## Get Started

### Requirements

- Beancount v2
- Go v1.22

### Installation

```bash
go install github.com/jacob-ian/beanport@latest
```

### Usage

```bash
beanport
```

#### Arguments

- `--provider`: The provider for the financial insitution
- `--input`: The input file
- `--output`: The output beancount file
- `--defaults`: The location of your beanport defaults file
- `--commodity`: The commodity of the account, e.g. AUD
- `--account`: The name of the account that owns the transactions
