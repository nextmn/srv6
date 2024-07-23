// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	srv6_app "github.com/nextmn/srv6/internal/app"
	srv6_config "github.com/nextmn/srv6/internal/config"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetPrefix("[nextmn-SRv6] ")
	var config_file string
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
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
				log.Println("Error loading config, exiting…:", err)
				os.Exit(1)
			}

			if err := srv6_app.NewSetup(conf).Run(ctx); err != nil {
				log.Println("Error while running, exiting…:", err)
				os.Exit(2)
			}
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
