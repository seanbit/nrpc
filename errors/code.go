package errors

// ErrUnknownCode is a string code representing an unknown error
// This will be used when no error code is sent by the handler
const ErrUnknownCode = "PIT-000"

// ErrInternalCode is a string code representing an internal NRpc error
const ErrInternalCode = "PIT-500"

// ErrNotFoundCode is a string code representing a not found related error
const ErrNotFoundCode = "PIT-404"

// ErrBadRequestCode is a string code representing a bad request related error
const ErrBadRequestCode = "PIT-400"

const ErrRequestTimeoutCode = "PIT-408"
const ErrTooManyRequestCode = "PIT-429"

// ErrClientClosedRequest is a string code representing the client closed request error
const ErrClientClosedRequest = "PIT-499"

// ErrClosedRequest is a string code representing the closed request error
const ErrClosedRequest = "PIT-498"

// ErrServiceUnavailable is a string code representing the service under maintenance
const ErrServiceUnavailable = "PIT-503"
