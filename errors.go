// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"errors"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
)

// APIError is implemented by every error this package returns. It mirrors the
// Ruby exception hierarchy that a binding layer exposes as
//
//	OpenStack::Error < StandardError
//	  OpenStack::NotFound   < OpenStack::Error
//	  OpenStack::AuthError  < OpenStack::Error
//	  OpenStack::Forbidden  < OpenStack::Error
//	  OpenStack::Conflict   < OpenStack::Error
//	  OpenStack::BadRequest < OpenStack::Error
//
// Every concrete type below satisfies APIError, so a consumer can either type
// switch, use the Is* predicates, or (in Ruby) rescue the mapped class.
type APIError interface {
	error
	// Status returns the HTTP status code carried by the underlying API
	// response, or 0 when the error did not originate from one.
	Status() int
	// Unwrap exposes the wrapped gophercloud error for errors.Is/As.
	Unwrap() error
}

// baseError is embedded by every concrete error type. It is the Go analogue of
// the common OpenStack::Error superclass.
type baseError struct {
	message string
	status  int
	cause   error
}

// Error implements the error interface.
func (e *baseError) Error() string { return e.message }

// Status returns the HTTP status code, or 0 if not from an API response.
func (e *baseError) Status() int { return e.status }

// Unwrap returns the wrapped gophercloud error.
func (e *baseError) Unwrap() error { return e.cause }

// Error is the generic error returned when no more specific type applies (for
// example a transport failure or a 5xx response). It is the base of the tree.
type Error struct{ baseError }

// NotFoundError maps gophercloud's 404 responses (OpenStack::NotFound).
type NotFoundError struct{ baseError }

// AuthError maps gophercloud's 401 responses and authentication failures
// (OpenStack::AuthError).
type AuthError struct{ baseError }

// ForbiddenError maps gophercloud's 403 responses (OpenStack::Forbidden).
type ForbiddenError struct{ baseError }

// ConflictError maps gophercloud's 409 responses (OpenStack::Conflict).
type ConflictError struct{ baseError }

// BadRequestError maps gophercloud's 400 responses (OpenStack::BadRequest).
type BadRequestError struct{ baseError }

// IsNotFound reports whether err is (or wraps) a NotFoundError.
func IsNotFound(err error) bool {
	var e *NotFoundError
	return errors.As(err, &e)
}

// IsAuth reports whether err is (or wraps) an AuthError.
func IsAuth(err error) bool {
	var e *AuthError
	return errors.As(err, &e)
}

// IsForbidden reports whether err is (or wraps) a ForbiddenError.
func IsForbidden(err error) bool {
	var e *ForbiddenError
	return errors.As(err, &e)
}

// IsConflict reports whether err is (or wraps) a ConflictError.
func IsConflict(err error) bool {
	var e *ConflictError
	return errors.As(err, &e)
}

// IsBadRequest reports whether err is (or wraps) a BadRequestError.
func IsBadRequest(err error) bool {
	var e *BadRequestError
	return errors.As(err, &e)
}

// mapError translates any error returned by gophercloud into this package's
// typed error tree. HTTP status codes 404/401/403/409/400 map onto their
// dedicated types; everything else (transport errors, 5xx, marshalling
// failures) becomes a generic *Error. A nil error maps to nil.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	base := baseError{message: err.Error(), cause: err}
	switch {
	case gophercloud.ResponseCodeIs(err, http.StatusNotFound):
		base.status = http.StatusNotFound
		return &NotFoundError{base}
	case gophercloud.ResponseCodeIs(err, http.StatusUnauthorized):
		base.status = http.StatusUnauthorized
		return &AuthError{base}
	case gophercloud.ResponseCodeIs(err, http.StatusForbidden):
		base.status = http.StatusForbidden
		return &ForbiddenError{base}
	case gophercloud.ResponseCodeIs(err, http.StatusConflict):
		base.status = http.StatusConflict
		return &ConflictError{base}
	case gophercloud.ResponseCodeIs(err, http.StatusBadRequest):
		base.status = http.StatusBadRequest
		return &BadRequestError{base}
	default:
		var codeErr gophercloud.ErrUnexpectedResponseCode
		if errors.As(err, &codeErr) {
			base.status = codeErr.Actual
		}
		return &Error{base}
	}
}

// authError wraps a low-level authentication failure as an AuthError regardless
// of the underlying status, so Connection.Connect always surfaces a typed
// OpenStack::AuthError to Ruby.
func authError(err error) error {
	if err == nil {
		return nil
	}
	if mapped := mapError(err); IsAuth(mapped) {
		return mapped
	}
	return &AuthError{baseError{message: err.Error(), cause: err}}
}
