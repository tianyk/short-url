package errors

type PrivateScopeError struct {
    UrlId string
}

func (err PrivateScopeError) Error() string {
    return err.UrlId
}
