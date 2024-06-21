// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package ctrl

import (
	"fmt"
	"net/http"
	"net/netip"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/nextmn/json-api/jsonapi"
)

// A RulesRegistry contains rules for an headend
type RulesRegistry struct {
	sync.RWMutex
	rules jsonapi.RuleMap
}

func NewRulesRegistry() *RulesRegistry {
	return &RulesRegistry{
		rules: make(jsonapi.RuleMap),
	}
}

func (rr *RulesRegistry) Action(dstIp netip.Addr) (jsonapi.Action, error) {
	rr.RLock()
	defer rr.RUnlock()
	for _, r := range rr.rules {
		if !r.Enabled {
			continue
		}
		if r.Match.DstIpPrefix.Contains(dstIp) {
			return r.Action, nil
		}
	}
	return jsonapi.Action{}, fmt.Errorf("Not found")
}

func (rr *RulesRegistry) GetRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	rr.Lock()
	defer rr.Unlock()
	if val, ok := rr.rules[iduuid]; ok {
		c.JSON(http.StatusOK, val)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "rule not found"})
}

func (rr *RulesRegistry) GetRules(c *gin.Context) {
	c.JSON(http.StatusOK, rr.rules)
}

func (rr *RulesRegistry) DeleteRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	rr.Lock()
	defer rr.Unlock()
	if _, exists := rr.rules[iduuid]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "rule not found"})
		return
	}
	delete(rr.rules, iduuid)
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
	rr.Lock()
	defer rr.Unlock()
	if val, ok := rr.rules[iduuid]; ok {
		val.Enabled = true
		rr.rules[iduuid] = val // rules is not a map of pointers
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "rule not found"})
}

func (rr *RulesRegistry) DisableRule(c *gin.Context) {
	id := c.Param("uuid")
	iduuid, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad uuid", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	rr.Lock()
	defer rr.Unlock()
	if val, ok := rr.rules[iduuid]; ok {
		val.Enabled = false
		rr.rules[iduuid] = val // rules is not a map of pointers
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "rule not found"})
}

// Post a new rule
func (rr *RulesRegistry) PostRule(c *gin.Context) {
	var rule jsonapi.Rule
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not deserialize", "error": fmt.Sprintf("%v", err)})
		return
	}
	c.Header("Cache-Control", "no-cache")
	rr.Lock()
	defer rr.Unlock()

	// TODO: perform checks

	id, err := uuid.NewV4()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate UUID"})
	}
	for {
		if _, exists := rr.rules[id]; !exists {
			break
		} else {
			id, err = uuid.NewV4()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate UUID"})
			}
		}
	}
	rr.rules[id] = rule
	c.Header("Location", fmt.Sprintf("/rules/%s", id))
	c.JSON(http.StatusCreated, rr.rules[id])
}
