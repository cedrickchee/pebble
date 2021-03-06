// Copyright 2019 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package errbase

import (
	"context"
	goErr "errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	pkgErr "github.com/pkg/errors"
)

// This file provides the library the ability to encode/decode
// standard error types.

// errors.errorString from base Go does not need an encoding
// function, because the base encoding logic in EncodeLeaf() is
// able to extract everything about it.

// we can then decode it exactly.
func decodeErrorString(_ context.Context, msg string, _ []string, _ proto.Message) error {
	return goErr.New(msg)
}

// errors.fundamental from github.com/pkg/errors cannot be encoded
// exactly because it includes a non-serializable stack trace
// object. In order to work with it, we encode it by dumping
// the stack trace in a safe reporting detail field, and decode
// it as an opaqueLeaf instance in this package.

func encodePkgFundamental(
	_ context.Context, err error,
) (msg string, safe []string, _ proto.Message) {
	msg = err.Error()
	iErr := err.(interface{ StackTrace() pkgErr.StackTrace })
	safeDetails := []string{fmt.Sprintf("%+v", iErr.StackTrace())}
	return msg, safeDetails, nil
}

// errors.withMessage from github.com/pkg/errors can be encoded
// exactly because it just has a message prefix. The base encoding
// logic in EncodeWrapper() is able to extract everything from it.

// we can then decode it exactly.
func decodeWithMessage(
	_ context.Context, cause error, msgPrefix string, _ []string, _ proto.Message,
) error {
	return pkgErr.WithMessage(cause, msgPrefix)
}

// errors.withStack from github.com/pkg/errors cannot be encoded
// exactly because it includes a non-serializable stack trace
// object. In order to work with it, we encode it by dumping
// the stack trace in a safe reporting detail field, and decode
// it as an opaqueWrapper instance in this package.

func encodePkgWithStack(
	_ context.Context, err error,
) (msgPrefix string, safe []string, _ proto.Message) {
	iErr := err.(interface{ StackTrace() pkgErr.StackTrace })
	safeDetails := []string{fmt.Sprintf("%+v", iErr.StackTrace())}
	return "" /* withStack does not have a message prefix */, safeDetails, nil
}

func init() {
	baseErr := goErr.New("")
	RegisterLeafDecoder(GetTypeKey(baseErr), decodeErrorString)

	pkgE := pkgErr.New("")
	RegisterLeafEncoder(GetTypeKey(pkgE), encodePkgFundamental)

	RegisterWrapperDecoder(GetTypeKey(pkgErr.WithMessage(baseErr, "")), decodeWithMessage)

	ws := pkgErr.WithStack(baseErr)
	RegisterWrapperEncoder(GetTypeKey(ws), encodePkgWithStack)
}
