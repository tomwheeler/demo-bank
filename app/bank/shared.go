package banking

// InsufficientFundsError occurs when an account lacks the funds to
// successfully perform the requested operation.
type InsufficientFundsError struct {
	message string
}

func (e InsufficientFundsError) Error() string {
	return e.message
}
