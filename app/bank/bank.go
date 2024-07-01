package banking

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Bank represents an institution that offers basic financial accounts
// to customers. For the sake of simplicity in this demo, a given bank
// only has a single customer. This source file contains the business
// logic for managing the account and persisting the balance across
// sessions. The service.go file contains logic for exposing methods
// for account management over a network through a basic HTTP API.
// Implementation note: At present, idempotency keys are only retained
// for the lifetime of the service; they are not persisted during the
// current session and therefore values used then are not available for
// checking in the next session.
type Bank struct {
	name         string
	balance      int
	requests     map[string]string // idempotency keys => transaction IDs
	requestsLock sync.Mutex 
}

// NewBank returns a Bank instance for the named account. The balance
// for that account will be the same as in the previous session, or
// if there was no previous session, it will be zero (in which case
// you might call the Deposit method to provide initial funding).
func NewBank(name string) *Bank {
	bank := Bank{
		name:     name,
		balance:  0,
		requests: make(map[string]string),
	}

	previousBalance, err := bank.load()
	if err != nil {
		log.Printf("ERROR: Failed to load account data from previous session: %v\n", err)
	}

	bank.balance = previousBalance

	return &bank
}

// GetName returns the name used when creating the instance
func (bank *Bank) GetName() string {
	return bank.name
}

// GetBalance returns the current account balance
func (bank *Bank) GetBalance() int {
	return bank.balance
}

// Deposit adds the specified amount to the balance. The
// idempotency key is used to identify duplicate requests.
// This returns the transaction ID if successful or will return an
// error if the amount is invalid (zero or negative).
func (bank *Bank) Deposit(amount int, idempotencyKey string) (string, error) {
	if amount < 1 {
		return "", fmt.Errorf("Invalid amount - %d", amount)
	}

	// check idempotency key, only process deposit if it's unique. If it's a
	// duplicate, return the transaction ID from the original deposit.
	if idempotencyKey != "" {
		previousTxID, keyExists := bank.requests[idempotencyKey]
		if keyExists {
			msg := "Duplicate request for idempotency key '%s', returning txID: '%s'"
			log.Printf(msg, idempotencyKey, previousTxID)
			return previousTxID, nil
		}
	}

	bank.balance = bank.balance + amount
	txID := generateTransactionID("D", 10)
	bank.requestsLock.Lock()
	bank.requests[idempotencyKey] = txID
	bank.requestsLock.Unlock()

	err := bank.save()
	if err != nil {
		log.Printf("ERROR: could not save account data following deposit: %v\n", err)
		return "", err
	}

	log.Printf("Deposited $%d into '%s' account (ID: %s)", amount, bank.name, txID)
	return txID, nil
}

// Withdraw removes the specified amount from the balance. The
// idempotency key is used to identify duplicate requests. This
// returns a transaction ID if successful or will return an
// error if the amount is invalid (either negative or greater
// than the current balance).
func (bank *Bank) Withdraw(amount int, idempotencyKey string) (string, error) {
	if amount < 1 {
		return "", fmt.Errorf("Invalid amount: %d", amount)
	}

	if amount > bank.balance {
		msg := "insufficient funds: withdrawal amount $%d exceeds balance $%d"
		return "", fmt.Errorf(msg, amount, bank.balance)
	}

	// check idempotency key, only process withdrawal if it's unique. If it's a
	// duplicate, return the transaction ID from the original withdrawal.
	if idempotencyKey != "" {
		previousTxID, keyExists := bank.requests[idempotencyKey]
		if keyExists {
			msg := "Duplicate request for idempotency key '%s', returning txID: '%s'"
			log.Printf(msg, idempotencyKey, previousTxID)
			return previousTxID, nil
		}
	}

	bank.balance = bank.balance - amount
	txID := generateTransactionID("W", 10)
	bank.requestsLock.Lock()
	bank.requests[idempotencyKey] = txID
	bank.requestsLock.Unlock()

	err := bank.save()
	if err != nil {
		log.Printf("ERROR: could not save account data following withdrawal: %v\n", err)
		return "", err
	}

	log.Printf("Withdrew $%d from '%s' account (ID: %s)", amount, bank.name, txID)
	return txID, nil
}

// Generates a transaction ID with the specified prefix and of
// the specified length.
func generateTransactionID(prefix string, length int) string {
	randChars := make([]byte, length)
	for i := range randChars {
		allowedChars := "0123456789"
		randChars[i] = allowedChars[rand.Intn(len(allowedChars))]
	}

	return prefix + string(randChars)
}

// GetDataPath returns the path of the file where account data is persisted
func (bank *Bank) GetDataPath() string {
	// TODO: strip non-alpha characters from string
	fileName := fmt.Sprintf("bank-%s.dat", strings.ToLower(bank.name))

	// if possible, locate the file in the same directory as this source
	dir, err := filepath.Abs("./")
	if err == nil {
		fileName = filepath.Join(dir, fileName)
	}

	return fileName
}

// Load the account balance from the previous session. It returns
// the balance from the previous session that was stored to disk
// (and returns zero if no previous session data file exists).
// It returns an error if the data file exists, but could not be
// loaded for some reason (such as the data file being corrupted).
func (bank *Bank) load() (int, error) {
	// load initial data from a JSON file in current directory
	dataFileName := bank.GetDataPath()

	var balance = 0
	if _, err := os.Stat(dataFileName); err == nil {
		// data from previous session exists, load it
		log.Printf("Loading '%s' account data from file '%s'\n", bank.name, dataFileName)

		db, err := os.ReadFile(dataFileName)
		if err != nil {
			log.Printf("ERROR: problem loading file '%s': %v\n", dataFileName, err)
			return -1, err
		}

		err = json.Unmarshal([]byte(db), &balance)
		if err != nil {
			log.Printf("ERROR: could not unmarshal account info: %v\n", err)
			return -1, err
		}
	}

	return balance, nil
}

// Save the current account balance to disk so that it can be
// loaded in a future session. This returns an error if the data
// file could not be written for some reason.
func (bank *Bank) save() error {
	dataFileName := bank.GetDataPath()

	log.Printf("Writing account info to database '%s'\n", dataFileName)
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(bank.balance)

	data, _ := json.MarshalIndent(bank.balance, "", "  ")
	file, err := os.Create(dataFileName)
	if err != nil {
		log.Fatalf("failed to write account data to DB: %v", err)
	}
	defer file.Close()

	file.Write(data)
	file.Sync()
	log.Printf("Finished writing account info to database '%s'\n", dataFileName)

	return err
}
