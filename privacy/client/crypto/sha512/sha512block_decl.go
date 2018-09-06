// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x ppc64le

package sha512

//go:noescape

func block(dig *digest, p []byte)
