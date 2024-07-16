// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package ctrl

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/srv6/internal/database"
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	rule, err := rr.db.GetRule(iduuid)
	if err != nil {
		log.Printf("Could not get rule from database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not get rule from database"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (rr *RulesRegistry) GetRules(c *gin.Context) {
	rules, err := rr.db.GetRules()
	if err != nil {
		log.Printf("Could not get all rules from database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not get all rules from database"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (rr *RulesRegistry) DeleteRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.DeleteRule(iduuid)
	if err != nil {
		log.Printf("Could not delete rule in the database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not delete rule in the database"})
		return
	}
	c.Status(http.StatusNoContent) // successful deletion
}

func (rr *RulesRegistry) EnableRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.EnableRule(iduuid)
	if err != nil {
		log.Printf("Could not enable rule in the database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not enable rule in the database"})
		return
		//TODO: check if rule not found
	}
	c.Status(http.StatusNoContent)
}

func (rr *RulesRegistry) DisableRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	err = rr.db.DisableRule(iduuid)
	if err != nil {
		log.Printf("Could not disable rule in the database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not disable rule in the database"})
		return
		//TODO: check if rule not found
	}
	c.Status(http.StatusNoContent)
}

// Post a new rule
func (rr *RulesRegistry) PostRule(c *gin.Context) {
	var rule jsonapi.Rule
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not deserialize", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	id, err := rr.db.InsertRule(rule)
	if err != nil {
		log.Printf("Could not insert rule in the database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to insert rule"})
	}
	c.Header("Location", fmt.Sprintf("/rules/%s", id))
	c.JSON(http.StatusCreated, rule)
}
