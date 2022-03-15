// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"github.com/urfave/cli/v2"
)

const DefaultKeystorePath = "./keys"

var (
	ConfigPath = &cli.StringFlag{
		Name:  "C",
		Usage: "Path to configfile",
		Value: "./conf_payer.toml",
	}
)
