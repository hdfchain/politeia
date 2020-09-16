// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/hdfchain/hdfd/chaincfg/v3"
	"github.com/hdfchain/hdfd/wire"
	"github.com/hdfchain/hdfdata/netparams"
)

// activeNetParams is a pointer to the parameters specific to the
// currently active hdfchain network.
var activeNetParams = &mainNetParams

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	*chaincfg.Params
	WalletRPCServerPort string
}

// mainNetParams contains parameters specific to the main network
// (wire.MainNet).  NOTE: The RPC port is intentionally different than the
// reference implementation because hdfd does not handle wallet requests.  The
// separate wallet process listens on the well-known port and forwards requests
// it does not handle on to hdfd.  This approach allows the wallet process
// to emulate the full reference implementation RPC API.
//noinspection ALL
var mainNetParams = params{
	Params:              &chaincfg.MainNetParams,
	WalletRPCServerPort: netparams.MainNetParams.GRPCServerPort,
}

// testNet3Params contains parameters specific to the test network (version 0)
// (wire.TestNet).  NOTE: The RPC port is intentionally different than the
// reference implementation - see the mainNetParams comment for details.
//noinspection ALL
var testNet3Params = params{
	Params:              &chaincfg.TestNet3Params,
	WalletRPCServerPort: netparams.TestNet3Params.GRPCServerPort,
}

// simNetParams contains parameters specific to the simulation test network
// (wire.SimNet).
//noinspection ALL
var simNetParams = params{
	Params:              &chaincfg.SimNetParams,
	WalletRPCServerPort: netparams.SimNetParams.GRPCServerPort,
}

// netName returns the name used when referring to a hdfchain network.  At the
// time of writing, hdfd currently places blocks for testnet version 0 in the
// data and log directory "testnet", which does not match the Name field of the
// chaincfg parameters.  This function can be used to override this directory name
// as "testnet" when the passed active network matches wire.TestNet.
//
// A proper upgrade to move the data and log directories for this network to
// "testnet" is planned for the future, at which point this function can be
// removed and the network parameter's name used instead.
//noinspection ALL
func netName(chainParams *params) string {
	switch chainParams.Net {
		case wire.TestNet3:
			return "testnet3"
		default:
			return chainParams.Name
	}
}
