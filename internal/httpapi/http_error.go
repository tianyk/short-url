package httpapi

import (
	"net/http"
)

type HTTPError struct {
	Status  int
	Message string
	Err     error
}

func (err HTTPError) Error() string {
	message := err.Message
	if message == "" {
		message = http.StatusText(err.Status)
	}

	if err.Err != nil {
		return message + ": " + err.Err.Error()
	}

	return message
}
