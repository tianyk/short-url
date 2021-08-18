package errors

import (
    "net/http"
)

type HttpError struct {
    Status  int
    Message string
    Err     error
}

func (err HttpError) Error() string {
    var message string
    if err.Message == "" {
        message = http.StatusText(err.Status)
    }

    if err.Err != nil {
        return message + ": " + err.Err.Error()
    } else {
        return message
    }
}
