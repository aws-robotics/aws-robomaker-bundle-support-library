package s3

// ReadError represents an error while
// attempting to read from AWS S3
type ReadError struct {
	err error
}

func (e *ReadError) Error() string {
	return e.err.Error()
}
