// Copyright 2017 Alexander Palaistras. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.
//
//go:build !php5
// +build !php5

package php

// #cgo CFLAGS: -Iinclude/php -Isrc/php
// #cgo LDFLAGS: -lphp
import "C"
