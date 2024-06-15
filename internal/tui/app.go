// An interactive TUI for sorting transactions into accounts
package tui

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/jacob-ian/beanport/internal/beanport"
)

type Config struct {
	// The path to the preferences file
	PreferencesFilePath string

	// The path of the file to output
	OutputFilePath string

	// The data importer
	Importer beanport.Importer

	// The users defaults
	Defaults *beanport.UserDefaults
}

type Application struct {
	outputFilePath string
	commodity      string
	importer       beanport.Importer
	logger         *slog.Logger
	defaults       *beanport.UserDefaults
}

// Creates a TUI for the user to review and sort the transactions
func New(config Config) *Application {
	return &Application{
		outputFilePath: config.OutputFilePath,
		importer:       config.Importer,
		logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
		defaults:       config.Defaults,
	}
}

// Safely exit the running TUI
func (app *Application) SafelyExit() error {
	fmt.Println("Saving defaults")
	return app.defaults.WriteToFile()
}

// Runs the sort and review TUI
func (app *Application) Run() error {
	imported, err := app.importer.Import()
	if err != nil {
		return errors.New("Import failed: " + err.Error())
	}

	bold := color.New().Add(color.Bold)
	bold.Printf("Found %v transactions.\n", len(imported))

	vendors := make(map[string][]*beanport.PendingTransaction)

	for _, txn := range imported {
		if len(vendors[txn.Description]) == 0 {
			vendors[txn.Description] = []*beanport.PendingTransaction{txn}
		} else {
			vendors[txn.Description] = append(vendors[txn.Description], txn)
		}
	}

	var complete []*beanport.Transaction
	manual := make(map[string][]*beanport.PendingTransaction)
	for vendor, txns := range vendors {
		account, ok := app.defaults.CheckVendor(vendor)
		if ok {
			for _, pending := range txns {
				complete = append(complete, &beanport.Transaction{
					PendingTransaction: *pending,
					OppositeAccount:    account,
				})
			}
			continue
		}
		manual[vendor] = txns
	}

	bold.Printf("Automatically identified %v transactions.\n", len(complete))

	if len(manual) > 0 {
		app.runManualAttribution(manual, complete)
	}

	bold.Printf("Outputting %v transactions to beancount file.\n", len(complete))

	ledger := ""
	for _, txn := range complete {
		ledger += beanport.FormatTransaction(txn)
	}

	outFile, err := os.OpenFile(app.outputFilePath, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return errors.Join(errors.New("Could not open output file"), err)
	}

	defer outFile.Close()

	_, err = outFile.Write([]byte(ledger))
	if err != nil {
		return errors.Join(errors.New("Could not write transactions to file"), err)
	}

	bold.Println("Formatting output with bean-format")
	cmd := exec.Command(fmt.Sprintf("bean-format %s", app.outputFilePath))
	err = cmd.Run()
	if err != nil {
		panic("Could not format: " + err.Error())
	}

	return nil
}

func (app *Application) runManualAttribution(manual map[string][]*beanport.PendingTransaction, complete []*beanport.Transaction) {
	bold := color.New().Add(color.Bold)
	bold.Printf("%v vendor(s) requiring manual attribution.\n", len(manual))
	fmt.Printf("Press return to begin...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	idx := 1
	total := len(manual)
	for vendor, txns := range manual {
		bold.Printf("\n%v/%v: \"%v\"\n", idx, total, vendor)

		for _, txn := range txns {
			fmt.Println()
			fmt.Println(beanport.FormatPending(txn))
		}

		fmt.Println()

		bold.Printf("Assign to Account: ")

		account, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		account = account[:len(account)-1]
		app.defaults.SaveVendor(vendor, account)

		for _, pending := range txns {
			txn := &beanport.Transaction{
				PendingTransaction: *pending,
				OppositeAccount:    account,
			}
			complete = append(complete, txn)
		}

		err = app.defaults.WriteToFile()
		if err != nil {
			app.logger.Warn("Could not save vendor to defaults file.\n")
		}

		idx++
	}
}
