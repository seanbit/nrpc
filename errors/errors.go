// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package errors

import "errors"

// Error is an error with a code, message and metadata
type Error struct {
	Code     string
	Message  string
	Metadata map[string]string
}

// NewError ctor
func NewError(err error, code string, metadata ...map[string]string) *Error {
	var nrpcErr *Error
	if ok := errors.As(err, &nrpcErr); ok {
		if len(metadata) > 0 {
			mergeMetadatas(nrpcErr, metadata[0])
		}
		return nrpcErr
	}

	e := &Error{
		Code:    code,
		Message: err.Error(),
	}
	if len(metadata) > 0 {
		e.Metadata = metadata[0]
	}
	return e

}

func (e *Error) Error() string {
	return e.Message
}

func mergeMetadatas(nrpcErr *Error, metadata map[string]string) {
	if nrpcErr.Metadata == nil {
		nrpcErr.Metadata = metadata
		return
	}

	for key, value := range metadata {
		nrpcErr.Metadata[key] = value
	}
}

// FromError returns the code of error.
// If error is nil, return empty string.
// If error is not a nrpc error, returns unkown code
func FromError(err error) string {
	if err == nil {
		return ""
	}

	nrpcErr, ok := err.(*Error)
	if !ok {
		return ErrUnknownCode
	}
	if nrpcErr == nil {
		return ErrUnknownCode
	}
	return nrpcErr.Code
}
