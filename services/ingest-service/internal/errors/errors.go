package errors

// A LambdaError is an error that occurred during the execution of a Lambda function.
type LambdaError struct {
	StatusCode    int    `json:"code"`
	Message       string `json:"message"`
	OriginalError error  `json:"-"`
}

// Error returns the error message.
func (e LambdaError) Error() string {
	return e.Message
}

// NewLambdaError creates a new LambdaError.
func NewLambdaError(statusCode int, message string) LambdaError {
	return LambdaError{
		StatusCode: statusCode,
		Message:    message,
	}
}
