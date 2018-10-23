// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2017 The go-matrix Authors

//go:generate go-bindata -nometadata -o assets.go -pkg tracers -ignore ((tracers)|(assets)).go ./...
//go:generate gofmt -s -w assets.go

// Package tracers contains the actual JavaScript tracer assets.
package tracers
