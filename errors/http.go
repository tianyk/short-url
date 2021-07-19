package errors

type HttpError struct {
    Status  int
    Message string
    Err     error
}

func (err HttpError) Error() string {
    if err.Err != nil {
        return err.Message + ": " + err.Err.Error()
    } else {
        return err.Message
    }
}
