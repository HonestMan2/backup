// Copyright (c) 2018 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2017 The go-matrix Authors

package params

const (
	// These are the multipliers for man denominations.
	// Example: To get the wei value of an amount in 'douglas', use
	//    new(big.Int).Mul(value, big.NewInt(params.Douglas))
	Wei      = 1
	Ada      = 1e3
	Babbage  = 1e6
	Shannon  = 1e9
	Szabo    = 1e12
	Finney   = 1e15
	Ether    = 1e18
	Einstein = 1e21
	Douglas  = 1e42
)
