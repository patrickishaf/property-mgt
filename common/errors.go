package common

const (
	ErrCodeExpired          string = "code expired"
	ErrDBInsertFailed       string = "failed to insert into db"
	ErrDBOperationFailed    string = "failed to perform db operation"
	ErrDBRecordNotFound     string = "record not found in the db"
	ErrGeneric              string = "something went wrong"
	ErrInvalidCredentials   string = "invalid credentials"
	ErrInvalidRequest       string = "invalid request"
	ErrMethodNotImplemented string = "method not implemented"
	ErrNoRowsInResultSet    string = "no rows in result set"
	ErrRequestRateLimited   string = "request rate limited"
	ErrSocketClosed         string = "websocket: close 1001 (going away)"
)
