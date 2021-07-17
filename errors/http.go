package errors

type HttpError struct {
    Status  int
    Message string
}

func (err HttpError) Error() string {
    return err.Message
}
