package banking

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// BankClient allows a caller to invoke operations (such as Withdraw
// and Deposit) provided by a banking service available via a network.
type BankClient struct {
	host string
	port int
}

// NewBankClient creates a BankClient and returns a pointer to it.
func NewBankClient(host string, port int) *BankClient {
	client := BankClient{
		host: host,
		port: port,
	}

	return &client
}

// GetName returns the name of the bank that the client will access
func (client *BankClient) GetName() (string, error) {
	base := "http://%s:%d/name"
	url := fmt.Sprintf(base, client.host, client.port)

	content, err := callService(url)
	if err != nil {
		fmt.Printf("Error retrieving name: %v\n", err)
		return "", err
	}

	_, name, found := strings.Cut(content, "=")
	if !found {
		return "", fmt.Errorf("failed to parse name from service response: %s", content)
	}

	return name, nil
}

// GetBalance returns the current account balance
func (client *BankClient) GetBalance() (int, error) {
	base := "http://%s:%d/balance"
	url := fmt.Sprintf(base, client.host, client.port)

	content, err := callService(url)
	if err != nil {
		fmt.Printf("Error retrieving balance: %v\n", err)
		return -1, err
	}

	_, balanceString, _ := strings.Cut(content, "=")
	balance, err := strconv.Atoi(balanceString)
	if err != nil {
		fmt.Printf("failed to parse balance from service response: %v\n", err)
		return -1, err
	}

	return balance, nil
}

// Deposit calls the banking service, requesting that it adds the
// specified amount to the balance. The idempotency key is used to
// identify duplicate requests. This returns the transaction ID
// if successful or an error if it was not.
func (client *BankClient) Deposit(amount int, idempotencyKey string) (string, error) {
	base := "http://%s:%d/deposit?amount=%d&idempotency-key=%s"
	url := fmt.Sprintf(base, client.host, client.port, amount, url.QueryEscape(idempotencyKey))

	content, err := callService(url)
	if err != nil {
		fmt.Printf("Error making deposit: %v\n", err)
		return "", err
	}

	_, transactionID, found := strings.Cut(content, "=")
	if !found {
		return "", fmt.Errorf("failed to parse ID from service response: %s", content)
	}

	return transactionID, nil
}

// Withdraw removes the specified amount from the balance. The
// idempotency key is used to identify duplicate requests. This
// returns a transaction ID if successful or will return an
// error if the amount is invalid (either negative or greater
// than the current balance).
func (client *BankClient) Withdraw(amount int, idempotencyKey string) (string, error) {
	base := "http://%s:%d/withdraw?amount=%d&idempotency-key=%s"
	url := fmt.Sprintf(base, client.host, client.port, amount, url.QueryEscape(idempotencyKey))

	content, err := callService(url)
	if err != nil {
		fmt.Printf("Error making withdrawal: %v\n", err)
		return "", err
	}

	_, transactionID, found := strings.Cut(content, "=")
	if !found {
		return "", fmt.Errorf("failed to parse ID from service response: %s", content)
	}

	return transactionID, nil
}

// IsServiceRunning returns true if the service is available, false otherwise
func (client *BankClient) IsServiceRunning() bool {
	base := "http://%s:%d/balance"
	url := fmt.Sprintf(base, client.host, client.port)
	_, err := http.Get(url)
	return err == nil
}

// utility function for making calls to the banking service
// Input is a valid URL (with URL-escaped parameters)
// Output is the response as a string, or an error
func callService(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := string(body)

	status := resp.StatusCode
	if status >= 400 {
		// Expose a specific type of business-level error so that it
		// could be defined as non-retryable in a RetryPolicy
		if strings.Contains(content, "INSUFFICIENT_FUNDS") {
			re := regexp.MustCompile(`INSUFFICIENT_FUNDS:\s(.*)`)
			matches := re.FindStringSubmatch(content)
			return "", InsufficientFundsError{message: matches[1]}
		}

		// some other type of error, such as a malformed request
		message := fmt.Sprintf("HTTP Error %d: %s", status, content)

		return "", errors.New(message)
	}

	return content, nil
}
