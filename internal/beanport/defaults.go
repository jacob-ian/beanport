package beanport

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type defaults struct {
	Accounts map[string][]string `yaml:"accounts"`
}

type UserDefaults struct {
	filePath string
	defaults
}

// Writes the users defaults to file
func (ud *UserDefaults) WriteToFile() error {
	f, err := os.OpenFile(ud.filePath, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return errors.Join(errors.New("Could not open defaults file"), err)
	}

	defer f.Close()

	b, err := yaml.Marshal(ud.defaults)
	if err != nil {
		return errors.Join(errors.New("Could not encode defaults"), err)
	}

	_, err = f.Write(b)
	if err != nil {
		return errors.Join(errors.New("Could not write defaults to file"), err)
	}

	return nil

}

// Checks stored default accounts for a particular vendor name
// returning the name of the account
func (ud *UserDefaults) CheckVendor(check string) (string, bool) {
	found := ""
	ok := false

outer:
	for account, vendors := range ud.Accounts {
		for _, vendor := range vendors {
			if check == vendor {
				ok = true
				found = account
				break outer
			}
		}
	}

	return found, ok
}

// Sets a vendors default account
func (ud *UserDefaults) SaveVendor(vendor string, account string) {
	if len(ud.Accounts[account]) == 0 {
		ud.Accounts[account] = []string{vendor}
		return
	}
	exists := false
	for _, v := range ud.Accounts[account] {
		if v == vendor {
			exists = true
			break
		}
	}
	if exists {
		return
	}
	ud.Accounts[account] = append(ud.Accounts[account], vendor)
}

// Import the user's defaults from a file
func UserDefaultsFromFile(path string) (*UserDefaults, error) {
	b, err := os.ReadFile(path)
	if err != nil && os.IsNotExist(err) {
		return &UserDefaults{
				filePath: path,
				defaults: defaults{
					Accounts: make(map[string][]string),
				},
			},
			nil
	} else if err != nil {
		return nil, err
	}

	defaults := defaults{
		Accounts: make(map[string][]string),
	}
	err = yaml.Unmarshal(b, &defaults)
	if err != nil {
		return nil, errors.Join(errors.New("Could not parse preferences file"), err)
	}

	return &UserDefaults{
		filePath: path,
		defaults: defaults,
	}, nil
}
