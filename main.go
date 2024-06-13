package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacob-ian/beanport/internal/beanport"
	"github.com/jacob-ian/beanport/internal/providers/amexcsv"
	"github.com/jacob-ian/beanport/internal/tui"
)

type Config struct {
	Account          string
	InputFilePath    string
	OutputFilePath   string
	DefaultsFilePath string
	Provider         beanport.Provider
	Commodity        string
}

func main() {
	cfg, err := getArgs()
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(cfg.InputFilePath)
	if err != nil {
		panic("Could not read file: " + err.Error())
	}

	var importer beanport.Importer

	if cfg.Provider == amexcsv.Provider {
		importer = amexcsv.NewImporter(data, &amexcsv.ImporterConfig{
			Account:         cfg.Account,
			NegativeAmounts: true,
			Commodity:       cfg.Commodity,
		})
	}

	defaults, err := beanport.UserDefaultsFromFile(cfg.DefaultsFilePath)
	if err != nil {
		panic("Could not load defaults: " + err.Error())
	}

	app := tui.New(tui.Config{
		OutputFilePath: cfg.OutputFilePath,
		Importer:       importer,
		Defaults:       defaults,
	})

	setupInterruptHandler(app)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func getArgs() (Config, error) {
	var provider beanport.Provider
	flag.StringVar(&provider, "provider", "", "amexcsv")
	var input string
	flag.StringVar(&input, "input", "", "amex.csv")
	var output string
	flag.StringVar(&output, "output", "", "~/finances/amex.beancount")
	var commodity string
	flag.StringVar(&commodity, "commodity", "AUD", "AUD")
	var account string
	flag.StringVar(&account, "account", "", "The name of the statement's account")
	var defaults string
	flag.StringVar(&defaults, "defaults", "beanport.yaml", "The defaults for beanport")

	flag.Parse()

	if provider == "" {
		return Config{}, errors.New("'provider' is required")
	}

	if input == "" {
		return Config{}, errors.New("'input' is required")
	}

	if account == "" {
		return Config{}, errors.New("'account' is required")
	}

	if commodity == "" {
		return Config{}, errors.New("'commodity' is required")
	}

	return Config{
		InputFilePath:    input,
		Provider:         provider,
		Account:          account,
		OutputFilePath:   output,
		Commodity:        commodity,
		DefaultsFilePath: defaults,
	}, nil
}

func setupInterruptHandler(app *tui.Application) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nCtrl+C Pressed")
		if err := app.SafelyExit(); err != nil {
			fmt.Printf("Could not safely exit: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Safely exited. Bye!")
		os.Exit(0)
	}()
}
