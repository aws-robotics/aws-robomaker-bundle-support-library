// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle_support

import "fmt"

const (
	ERROR_TYPE_CONTENT_ID = "CONTENT_ID"
	ERROR_TYPE_SOURCE     = "SOURCE"
	ERROR_TYPE_FORMAT     = "FORMAT"
	ERROR_TYPE_EXTRACTION = "EXTRACTION"
)

type BundleError struct {
	cause     error
	errorType string
}

func (err *BundleError) Error() string {
	return fmt.Sprintf("[%s] bundle error: %v", err.errorType, err.cause)
}

func (err *BundleError) GetCause() error {
	return err.cause
}

func (err *BundleError) GetErrorType() string {
	return err.errorType
}

func NewBundleError(err error, errorType string) *BundleError {
	return &BundleError{
		cause:     err,
		errorType: errorType,
	}
}
