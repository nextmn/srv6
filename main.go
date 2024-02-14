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
func initSignals(ch chan *srv6_app.Setup) {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	func(_ os.Signal) {}(<-cancelChan)
	select {
	case setup := <-ch:
		setup.Exit()
	default:
		break
	}
	os.Exit(0)
}

// Entrypoint
func main() {
	log.SetPrefix("[nextmn-SRv6] ")
	var config_file string
	ch := make(chan *srv6_app.Setup, 1)
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
				fmt.Println("Error loading config, exiting…:", err)
				os.Exit(1)
			}

			setup := srv6_app.NewSetup(conf)
			go func(cha chan *srv6_app.Setup, s *srv6_app.Setup) {
				cha <- s
			}(ch, setup)
			setup.AddTasks()
			if err := setup.Run(); err != nil {
				fmt.Println("Error while running, exiting…:", err)
				setup.Exit()
				os.Exit(2)
			}
			return nil
		},
	}
	go initSignals(ch)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
