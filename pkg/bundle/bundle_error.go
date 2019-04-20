// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import "fmt"

const (
	errorTypeContentID  = "CONTENT_ID"
	errorTypeSource     = "SOURCE"
	errorTypeFormat     = "FORMAT"
	errorTypeExtraction = "EXTRACTION"
)

type bundleError struct {
	cause     error
	errorType string
}

func (err *bundleError) Error() string {
	return fmt.Sprintf("[%s] bundle error: %v", err.errorType, err.cause)
}

func (err *bundleError) GetCause() error {
	return err.cause
}

func (err *bundleError) GetErrorType() string {
	return err.errorType
}

func newBundleError(err error, errorType string) *bundleError {
	return &bundleError{
		cause:     err,
		errorType: errorType,
	}
}
