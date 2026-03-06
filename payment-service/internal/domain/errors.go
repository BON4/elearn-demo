package domain

import "errors"

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrPaymentAlreadyFinal     = errors.New("payment already in final state")
	ErrRefundNotAllowed        = errors.New("refund not allowed")
)
