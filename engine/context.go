// Copyright 2015 Alexander Palaistras. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package engine

// #cgo CFLAGS: -I/usr/include/php -I/usr/include/php/main -I/usr/include/php/TSRM
// #cgo CFLAGS: -I/usr/include/php/Zend -Iinclude
//
// #include <stdlib.h>
// #include <main/php.h>
// #include "context.h"
import "C"

import (
	"fmt"
	"io"
	"net/http"
	"unsafe"
)

// Context represents an individual execution context.
type Context struct {
	// Output and Log are unbuffered writers used for regular and debug output,
	// respectively. If left unset, any data written into either by the calling
	// context will be lost.
	Output io.Writer
	Log    io.Writer

	// Header represents the HTTP headers set by current PHP context.
	Header http.Header

	context *C.struct__engine_context
	values  map[string]*Value
}

// NewContext creates a new execution context for the active engine and returns
// an error if the execution context failed to initialize at any point.
func NewContext() (*Context, error) {
	ctx := &Context{
		Header: make(http.Header),
		values: make(map[string]*Value),
	}

	ptr, err := C.context_new(unsafe.Pointer(ctx))
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize context for PHP engine")
	}

	ctx.context = ptr

	return ctx, nil
}

// Bind allows for binding Go values into the current execution context under
// a certain name. Bind returns an error if attempting to bind an invalid value
// (check the documentation for NewValue for what is considered to be a "valid"
// value).
func (c *Context) Bind(name string, val interface{}) error {
	v, err := NewValue(val)
	if err != nil {
		return err
	}

	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	C.context_bind(c.context, n, v.Ptr())
	c.values[name] = v

	return nil
}

// Exec executes a PHP script pointed to by filename in the current execution
// context, and returns an error, if any. Output produced by the script is
// written to the context's pre-defined io.Writer instance.
func (c *Context) Exec(filename string) error {
	f := C.CString(filename)
	defer C.free(unsafe.Pointer(f))

	_, err := C.context_exec(c.context, f)
	if err != nil {
		return fmt.Errorf("Error executing script '%s' in context", filename)
	}

	return nil
}

// Eval executes the PHP expression contained in script, and returns a Value
// containing the PHP value returned by the expression, if any. Any output
// produced is written context's pre-defined io.Writer instance.
func (c *Context) Eval(script string) (*Value, error) {
	// When PHP compiles code with a non-NULL return value expected, it simply
	// prepends a `return` call to the code, thus breaking simple scripts that
	// would otherwise work. Thus, we need to wrap the code in a closure, and
	// call it immediately.
	s := C.CString("call_user_func(function(){" + script + "});")
	defer C.free(unsafe.Pointer(s))

	result, err := C.context_eval(c.context, s)
	if err != nil {
		return nil, fmt.Errorf("Error executing script '%s' in context", script)
	}

	defer C.free(result)

	val, err := NewValueFromPtr(result)
	if err != nil {
		return nil, err
	}

	return val, nil
}

// Destroy tears down the current execution context along with any active value
// bindings for that context.
func (c *Context) Destroy() {
	if c.context == nil {
		return
	}

	for _, v := range c.values {
		v.Destroy()
	}

	c.values = nil

	C.context_destroy(c.context)
	c.context = nil
}