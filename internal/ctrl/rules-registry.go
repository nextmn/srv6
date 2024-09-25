// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package ctrl

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/srv6/internal/database"
	"github.com/sirupsen/logrus"
)

// A RulesRegistry contains rules for an headend
type RulesRegistry struct {
	db *database.Database
}

func NewRulesRegistry(db *database.Database) *RulesRegistry {
	return &RulesRegistry{
		db: db,
	}
}

func (rr *RulesRegistry) GetRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		logrus.WithError(err).Error("Bad UUID")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "bad uuid", Error: err})
		return
	}
	c.Header("Cache-Control", "no-cache")
	rule, err := rr.db.GetRule(c, iduuid)
	if err != nil {
		logrus.WithError(err).Error("Could not get rule from database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not get rule from database", Error: err})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (rr *RulesRegistry) GetRules(c *gin.Context) {
	rules, err := rr.db.GetRules(c)
	if err != nil {
		logrus.WithError(err).Error("Could not get all rules from database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not get all rules from database", Error: err})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (rr *RulesRegistry) DeleteRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		logrus.WithError(err).Error("Bad UUID")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "bad uuid", Error: err})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.DeleteRule(c, iduuid)
	if err != nil {
		logrus.WithError(err).Error("Could not delete rule in the database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not delete rule in the database", Error: err})
		return
	}
	c.Status(http.StatusNoContent) // successful deletion
}

func (rr *RulesRegistry) EnableRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		logrus.WithError(err).Error("Bad UUID")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "bad uuid", Error: err})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.EnableRule(c, iduuid)
	if err != nil {
		logrus.WithError(err).Error("Could not enable rule in the database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not enable rule in the database", Error: err})
		return
		//TODO: check if rule not found
	}
	c.Status(http.StatusNoContent)
}

func (rr *RulesRegistry) DisableRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		logrus.WithError(err).Error("Bad UUID")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "bad uuid", Error: err})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.DisableRule(c, iduuid)
	if err != nil {
		logrus.WithError(err).Error("Could not disable rule in the database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not disable rule in the database", Error: err})
		return
		//TODO: check if rule not found
	}
	c.Status(http.StatusNoContent)
}

func (rr *RulesRegistry) SwitchRule(c *gin.Context) {
	idEnable := c.Param("enable_uuid")
	idDisable := c.Param("disable_uuid")
	iduuidEnable, err := uuid.FromString(idEnable)
	if err != nil {
		logrus.WithError(err).Error("Bad UUID")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "bad uuid", Error: err})
		return
	}
	iduuidDisable, err := uuid.FromString(idDisable)
	if err != nil {
		logrus.WithError(err).Error("Bad UUID")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "bad uuid", Error: err})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.SwitchRule(c, iduuidEnable, iduuidDisable)
	if err != nil {
		logrus.WithError(err).Error("Could not Switch rule in the database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not switch rule in the database", Error: err})
		return
		//TODO: check if rule not found
	}
	c.Status(http.StatusNoContent)
}

// Post a new rule
func (rr *RulesRegistry) PostRule(c *gin.Context) {
	var rule jsonapi.Rule
	if err := c.BindJSON(&rule); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	c.Header("Cache-Control", "no-cache")
	id, err := rr.db.InsertRule(c, rule)
	if err != nil {
		logrus.WithError(err).Error("Could not insert rule in the database")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "failed to insert rule", Error: err})
		return
	}
	c.Header("Location", fmt.Sprintf("/rules/%s", id))
	c.JSON(http.StatusCreated, rule)
}
