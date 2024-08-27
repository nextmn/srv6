// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nextmn/srv6/internal/app"
	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/logger"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	logger.Init("NextMN-Srv6")
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
			conf, err := config.ParseConf(config_file)
			if err != nil {
				logrus.WithContext(ctx).WithError(err).Fatal("Error loading config, exiting…")
			}
			if conf.Logger != nil {
				logrus.SetLevel(conf.Logger.Level)
			}

			if err := app.NewSetup(conf).Run(ctx); err != nil {
				logrus.WithError(err).Fatal("Error while running, exiting…")
			}
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
