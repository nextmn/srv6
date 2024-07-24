// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package logger

import "github.com/sirupsen/logrus"

type logFormatter struct {
	logrus.TextFormatter
	prefix []byte
}

// Customized log formatter
func NewLogFormatter(prefix string) *logFormatter {
	return &logFormatter{
		TextFormatter: logrus.TextFormatter{
			ForceColors:            true,
			FullTimestamp:          true,
			DisableTimestamp:       false,
			DisableLevelTruncation: true,
		},
		prefix: []byte("[" + prefix + "] "),
	}
}

func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	l, e := f.TextFormatter.Format(entry)
	return append(f.prefix, l...), e
}
