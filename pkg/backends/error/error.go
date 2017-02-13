/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

// Package error describes errors in remco backends
package error

import "errors"

// BackendError contains an error message and the name of the backend that produced the error
type BackendError struct {
	Backend string
	Message string
}

// Error is for the error interface
func (e BackendError) Error() string {
	return e.Message
}

// ErrNilConfig is returned if Connect is called on a nil Config
var ErrNilConfig = errors.New("config is nil")
