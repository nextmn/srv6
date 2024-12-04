// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package ctrl_api

import (
	"github.com/gin-gonic/gin"
)

type RulesRegistryHTTP interface {
	GetRule(c *gin.Context)
	GetRules(c *gin.Context)
	DeleteRule(c *gin.Context)
	EnableRule(c *gin.Context)
	DisableRule(c *gin.Context)
	SwitchRule(c *gin.Context)
	PostRule(c *gin.Context)
}
