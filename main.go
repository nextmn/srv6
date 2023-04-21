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

	sr "github.com/louisroyer/nextmn-srv6/runtime"
	"github.com/urfave/cli/v2"
)

func initSignals() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	func(_ os.Signal) {}(<-cancelChan)
	sr.Exit()
	os.Exit(0)
}

func main() {
	log.SetPrefix("[nextmn-SRv6] ")
	var config string
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
				Destination: &config,
				Required:    true,
				DefaultText: "not set",
			},
		},
		Action: func(c *cli.Context) error {
			err := sr.ParseConf(config)
			if err != nil {
				fmt.Println("Error loading config, exiting…")
				os.Exit(1)
			}
			err = sr.Run()
			if err != nil {
				fmt.Println("Error while running, exiting…")
				log.Fatal(err)
				os.Exit(2)
			}
			return nil
		},
	}
	go initSignals()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
