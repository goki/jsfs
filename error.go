// Copyright (c) 2023, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js

package fs

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"syscall/js"

	"github.com/hack-pad/hackpadfs"
	"github.com/pkg/errors"
)

// JsError converts the given error value into a JS error value.
func JSError(err error, message string, args ...js.Value) js.Value {
	if err == nil {
		return js.Null()
	}

	errMessage := errors.Wrap(err, message).Error()
	for _, arg := range args {
		errMessage += fmt.Sprintf("\n%v", arg)
	}

	return js.ValueOf(map[string]interface{}{
		"message": js.ValueOf(errMessage),
		"code":    js.ValueOf(GetErrType(err, errMessage)),
	})
}

// GetErrType returns the JS type of the given error.
func GetErrType(err error, debugMessage string) string {
	if err, ok := err.(interface{ Code() string }); ok {
		return err.Code()
	}
	if err := errors.Unwrap(err); err != nil {
		return GetErrType(err, debugMessage)
	}
	switch err {
	case io.EOF, exec.ErrNotFound:
		return "ENOENT"
	}
	switch {
	case errors.Is(err, hackpadfs.ErrClosed):
		return "EBADF" // if it was already closed, then the file descriptor was invalid
	case errors.Is(err, hackpadfs.ErrNotExist):
		return "ENOENT"
	case errors.Is(err, hackpadfs.ErrExist):
		return "EEXIST"
	case errors.Is(err, hackpadfs.ErrIsDir):
		return "EISDIR"
	case errors.Is(err, hackpadfs.ErrPermission):
		return "EPERM"
	default:
		log.Printf("error: jsfs.GetErrType: unknown error type: (%T) %+v\n\n%s\n", err, err, debugMessage)
		return "EPERM"
	}
}

func WrapError(err error, code string) error {
	return &codeError{err, code}
}

type codeError struct {
	error
	code string
}

func (ce *codeError) Code() string { return ce.code }

func ErrBadFileNumber(fd uint64) error {
	return WrapError(fmt.Errorf("bad file number %d", fd), "EBADF")
}

func ErrBadFile(identifier string) error {
	return WrapError(fmt.Errorf("bad file %q", identifier), "EBADF")
}