// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	srv6_app "github.com/nextmn/srv6/internal/app"
	srv6_config "github.com/nextmn/srv6/internal/config"
	"github.com/urfave/cli/v2"
)

// Handler for os signals
func initSignals(setup *srv6_app.Setup) {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	func(_ os.Signal) {}(<-cancelChan)
	if setup != nil {
		setup.Exit()
	}
	os.Exit(0)
}

// Entrypoint
func main() {
	log.SetPrefix("[nextmn-SRv6] ")
	var config_file string
	var setup *srv6_app.Setup
	app := &cli.App{
		Name:                 "NextMN-SRv6",
		Usage:                "Experimental implementation of SRv6 SIDs for MUP",
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{Name: "Louis Royer"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Load configuration from `FILE`",
				Destination: &config_file,
				Required:    true,
				DefaultText: "not set",
			},
		},
		Action: func(c *cli.Context) error {
			conf, err := srv6_config.ParseConf(config_file)
			if err != nil {
				fmt.Println("Error loading config, exiting…")
				os.Exit(1)
			}

			setup = srv6_app.NewSetup(conf)
			if err := setup.Run(); err != nil {
				fmt.Println("Error while running, exiting…")
				log.Fatal(err)
				os.Exit(2)
			}
			return nil
		},
	}
	go initSignals(setup)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
