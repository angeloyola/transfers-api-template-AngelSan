package models

import "transfers-api/internal/enums"

type Transfer struct {
	ID         string
	SenderID   string
	ReceiverID string
	Currency   enums.Currency
	Amount     float64
	State      string // TODO: replace with enums.State
}
