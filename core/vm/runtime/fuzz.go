// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2017 The go-matrix Authors

// +build gofuzz

package runtime

// Fuzz is the basic entry point for the go-fuzz tool
// This returns 1 for valid parsable/runable code, 0
// for invalid opcode.
func Fuzz(input []byte) int {
	_, _, err := Execute(input, input, &Config{
		GasLimit: 3000000,
	})

	// invalid opcode
	if err != nil && len(err.Error()) > 6 && string(err.Error()[:7]) == "invalid" {
		return 0
	}

	return 1
}
