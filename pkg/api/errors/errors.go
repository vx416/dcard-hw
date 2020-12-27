package errors

import "net/http"

var (
	// 400 ~
	ErrTooManyRequestsError     = NewHttpErr(http.StatusTooManyRequests)
	ErrUnprocessableEntityError = NewHttpErr(http.StatusUnprocessableEntity)

	// 500 ~
	ErrInternalServerError = NewHttpErr(http.StatusInternalServerError)
)
