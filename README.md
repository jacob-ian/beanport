# Beanport!

A CLI tool to import transaction data into a beancount ledger, written in Go.

## Financial Institution Support

| Institution | Provider |
| ----------- | -------- |
| Amex (CSV)  | amexcsv  |

## Usage

### Requirements

- Beancount v2
- Go v1.22

```bash
beanport
```

- `--provider`: The provider for the financial insitution
- `--input`: The input file
- `--output`: The output beancount file
- `--defaults`: The location of your beanport defaults file
