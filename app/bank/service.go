package banking

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// BankingService represents account operations that a specific bank
// allows one to invoke over a network connection.
type BankingService struct {
	bank   *Bank
	port   int
	server *http.Server
}

// NewBankingService creates a new BankingService and returns a
// pointer to it.
func NewBankingService(bank *Bank, port int) *BankingService {
	svc := BankingService{
		bank:   bank,
		port:   port,
		server: &http.Server{Addr: fmt.Sprintf(":%d", port)},
	}

	return &svc
}

func (svc *BankingService) balanceHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "SUCCESS: balance=%d", svc.bank.GetBalance())
}

func (svc *BankingService) nameHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "SUCCESS: name=%s", svc.bank.GetName())
}

func (svc *BankingService) depositHandler(w http.ResponseWriter, r *http.Request) {
	amountParams, hasAmountParam := r.URL.Query()["amount"]
	if !hasAmountParam {
		http.Error(w, "ERROR: MISSING_AMOUNT_PARAM", http.StatusBadRequest)
		return
	}

	amount, err := strconv.Atoi(amountParams[0])
	if err != nil || amount < 1 {
		http.Error(w, "ERROR: INVALID_AMOUNT", http.StatusBadRequest)
		return
	}

	var idempotencyKey string
	idempotencyKeyParams, hasIdempotencyKeyParam := r.URL.Query()["idempotency-key"]
	if hasIdempotencyKeyParam {
		idempotencyKey = idempotencyKeyParams[0]
	}

	// TODO maybe expose specific errors such as insufficient funds,
	// but for now I'll just expose as a generic deposit failure
	txID, err := svc.bank.Deposit(amount, idempotencyKey)
	if err != nil {
		message := fmt.Sprintf("ERROR: DEPOSIT_FAIL: %v", err)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("SUCCESS: DEPOSIT_COMPLETE: transaction-id=%s", txID))
}

func (svc *BankingService) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	amountParams, hasAmountParam := r.URL.Query()["amount"]
	if !hasAmountParam {
		http.Error(w, "ERROR: MISSING_AMOUNT_PARAM", http.StatusBadRequest)
		return
	}

	amount, err := strconv.Atoi(amountParams[0])
	if err != nil || amount < 1 {
		http.Error(w, "ERROR: INVALID_AMOUNT", http.StatusBadRequest)
		return
	}

	var idempotencyKey string
	idempotencyKeyParams, hasIdempotencyKeyParam := r.URL.Query()["idempotency-key"]
	if hasIdempotencyKeyParam {
		idempotencyKey = idempotencyKeyParams[0]
	}

	txID, err := svc.bank.Withdraw(amount, idempotencyKey)
	if err != nil {
		message := fmt.Sprintf("ERROR: WITHDRAW_FAIL: %v", err)
		// Expose insufficient funds error so that the client will recognize
		// it and return it to the caller as a specific typed error
		if strings.Contains(err.Error(), "insufficient funds:") {
			// extract the details, which follows the colon
			re := regexp.MustCompile(`:\s(.*)`)
			matches := re.FindStringSubmatch(err.Error())
			message = fmt.Sprintf("ERROR: INSUFFICIENT_FUNDS: %s", matches[1])
		}

		http.Error(w, message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("SUCCESS: WITHDRAW_COMPLETE: transaction-id=%s", txID))
}

// Start starts the BankingService, allowing it to handle client reqests
func (svc *BankingService) Start() error {
	http.HandleFunc("/balance", svc.balanceHandler)
	http.HandleFunc("/name", svc.nameHandler)
	http.HandleFunc("/withdraw", svc.withdrawHandler)
	http.HandleFunc("/deposit", svc.depositHandler)

	return svc.server.ListenAndServe()
}

// Shutdown stops the BankingService, preventing it from handling client reqests
func (svc *BankingService) Shutdown() error {
	log.Printf("Shut down requested for '%s' banking service", svc.bank.GetName())

	return svc.server.Shutdown(context.Background())
}
