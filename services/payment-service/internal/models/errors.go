package models

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrAccountNotFound    = errors.New("payment account not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrEscrowNotFound     = errors.New("escrow not found")
	ErrEscrowNotHeld      = errors.New("escrow is not held")
	ErrForbiddenRequester = errors.New("requester cannot perform this payment action")
)
