// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"

	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/ctrl"
	tasks "github.com/nextmn/srv6/internal/tasks"
	tasks_api "github.com/nextmn/srv6/internal/tasks/api"
)

type Setup struct {
	config   *config.SRv6Config
	tasks    tasks_api.Registry
	registry app_api.Registry
}

func NewSetup(config *config.SRv6Config) *Setup {
	return &Setup{
		config:   config,
		tasks:    tasks.NewRegistry(),
		registry: NewRegistry(),
	}
}

// Add tasks to setup
func (s *Setup) AddTasks() {
	debug := false
	if s.config.Debug != nil {
		if *s.config.Debug {
			debug = true
		}
	}
	// user hooks
	var preInitHook, preExitHook *string
	var postInitHook, postExitHook *string
	if s.config.Hooks != nil {
		preInitHook = s.config.Hooks.PreInitHook
		preExitHook = s.config.Hooks.PreExitHook
		postInitHook = s.config.Hooks.PostInitHook
		postExitHook = s.config.Hooks.PostExitHook
	}
	// 0.1 pre-hooks
	s.tasks.Register(tasks.NewMultiHook("hook.pre.init", preInitHook, "hook.post.exit", postExitHook))

	httpPort := "80" // default http port
	if s.config.HTTPPort != nil {
		httpPort = *s.config.HTTPPort
	}
	httpURI := "http://"
	if s.config.HTTPAddress.Is6() {
		httpURI = httpURI + "[" + s.config.HTTPAddress.String() + "]:" + httpPort
	} else {
		httpURI = httpURI + s.config.HTTPAddress.String() + ":" + httpPort
	}
	httpAddr := fmt.Sprintf("[%s]:%s", s.config.HTTPAddress, httpPort)

	// 0.2 http server

	rr := ctrl.NewRulesRegistry()
	s.tasks.Register(tasks.NewHttpServerTask("ctrl.rest-api", httpAddr, rr))

	// 0.3 controller registry
	if s.config.Locator != nil {
		s.tasks.Register(tasks.NewControllerRegistryTask("ctrl.registry", s.config.ControllerURI, s.config.BackboneIP, *s.config.Locator, httpURI, s.registry))
		// 0.4 database
		s.tasks.Register(tasks.NewDBTask("database", s.registry))
	}

	// 1.  ifaces
	// 1.1 iface linux (type dummy)
	s.tasks.Register(tasks.NewTaskDummyIface("iproute2.iface.linux", constants.IFACE_LINUX))
	// 1.2 ifaces golang-srv6-* (tun via water)
	for i, e := range s.config.Endpoints.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.tun.golang-srv6/%s", e.Prefix)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_SRV6_PREFIX, i)
		s.tasks.Register(tasks.NewTaskTunIface(t_name, iface_name, s.registry))
	}
	// 1.3 ifaces golang-gtp4-* (tun via water)
	for i, h := range s.config.Headends.FilterWithBehavior(config.ProviderNextMN, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn.tun.golang-gtp4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_GTP4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskTunIface(t_name, iface_name, s.registry))
	}
	for i, h := range s.config.Headends.FilterWithBehavior(config.ProviderNextMNWithController, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn-ctrl.tun.golang-gtp4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_GTP4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskTunIface(t_name, iface_name, s.registry))
	}
	// 1.4 ifaces golang-ipv4-* (tun via water)
	for i, h := range s.config.Headends.FilterWithoutBehavior(config.ProviderNextMN, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn.tun.golang-ipv4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_IPV4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskTunIface(t_name, iface_name, s.registry))
	}
	for i, h := range s.config.Headends.FilterWithoutBehavior(config.ProviderNextMNWithController, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn-ctrl.tun.golang-ipv4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_IPV4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskTunIface(t_name, iface_name, s.registry))
	}

	// 2.  ip routes
	// 2.1 blackhole route (ipv6)
	s.tasks.Register(tasks.NewTaskBlackhole("iproute2.route.nextmn-ipv6.blackhole", constants.RT_TABLE_NEXTMN_IPV6))
	// 2.2 blackhole route (ipv4)
	s.tasks.Register(tasks.NewTaskBlackhole("iproute2.route.nextmn-ipv4.blackhole", constants.RT_TABLE_NEXTMN_IPV4))

	// 3.  endpoints + headends
	// 3.1 linux headends
	if s.config.LinuxHeadendSetSourceAddress != nil {
		s.tasks.Register(tasks.NewTaskLinuxHeadendSetSourceAddress("linux.headend.set-source-address", *s.config.LinuxHeadendSetSourceAddress))
	}
	for _, h := range s.config.Headends.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("linux.headend/%s", h.Name)
		s.tasks.Register(tasks.NewTaskLinuxHeadend(t_name, h, constants.RT_TABLE_NEXTMN_IPV4, constants.IFACE_LINUX))
	}
	// 3.1 linux endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("linux.endpoint/%s", e.Prefix)
		s.tasks.Register(tasks.NewTaskLinuxEndpoint(t_name, e, constants.RT_TABLE_NEXTMN_IPV6, constants.IFACE_LINUX))
	}
	// 3.2 nextmn endpoints
	for i, e := range s.config.Endpoints.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.endpoint/%s", e.Prefix)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_SRV6_PREFIX, i)
		s.tasks.Register(tasks.NewTaskNextMNEndpoint(t_name, e, constants.RT_TABLE_NEXTMN_IPV6, iface_name, s.registry, debug))
	}
	// 3.3 nextmn ipv4 headends
	for i, h := range s.config.Headends.FilterWithoutBehavior(config.ProviderNextMN, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn.headend.ipv4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_IPV4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskNextMNHeadend(t_name, h, constants.RT_TABLE_NEXTMN_IPV4, iface_name, s.registry, debug))
	}
	// 3.4 nextmn gtp4 headends
	for i, h := range s.config.Headends.FilterWithBehavior(config.ProviderNextMN, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn.headend.gtp4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_GTP4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskNextMNHeadend(t_name, h, constants.RT_TABLE_NEXTMN_IPV4, iface_name, s.registry, debug))
	}
	// 3.5 nextmn-ctrl ipv4 headends
	for i, h := range s.config.Headends.FilterWithoutBehavior(config.ProviderNextMNWithController, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn-ctrl.headend.ipv4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_IPV4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskNextMNHeadendWithCtrl(t_name, h, rr, constants.RT_TABLE_NEXTMN_IPV4, iface_name, s.registry, debug))
	}
	// 3.6 nextmn-ctrl gtp4 headends
	for i, h := range s.config.Headends.FilterWithBehavior(config.ProviderNextMNWithController, config.H_M_GTP4_D) {
		t_name := fmt.Sprintf("nextmn-ctrl.headend.gtp4/%s", h.Name)
		iface_name := fmt.Sprintf("%s%d", constants.IFACE_GOLANG_GTP4_PREFIX, i)
		s.tasks.Register(tasks.NewTaskNextMNHeadendWithCtrl(t_name, h, rr, constants.RT_TABLE_NEXTMN_IPV4, iface_name, s.registry, debug))
	}

	// 4.  ip rules
	// 4.1 rule to rttable nextmn-srv6
	if s.config.Locator != nil {
		s.tasks.Register(tasks.NewTaskIP6Rule("iproute2.rule.nextmn-srv6", *s.config.Locator, constants.RT_TABLE_NEXTMN_IPV6))
	}
	// 4.2 rule to rttable nextmn-gtp4
	if s.config.GTP4HeadendPrefix != nil {
		s.tasks.Register(tasks.NewTaskIP4Rule("iproute2.rule.nextmn-gtp4", *s.config.GTP4HeadendPrefix, constants.RT_TABLE_NEXTMN_IPV4))
	}
	// 4.3 rule to rttable nextmn-ipv4
	if s.config.IPV4HeadendPrefix != nil {
		s.tasks.Register(tasks.NewTaskIP4Rule("iproute2.rule.nextmn-ipv4", *s.config.IPV4HeadendPrefix, constants.RT_TABLE_NEXTMN_IPV4))
	}

	// user hooks
	s.tasks.Register(tasks.NewMultiHook("hook.post.init", postInitHook, "hook.pre.exit", preExitHook))
}

// Init
func (s *Setup) Init() error {
	return s.tasks.RunInit()
}

// Exit
func (s *Setup) Exit() {
	// This function may be called at any time,
	// and a maximum of exit tasks must be run,
	// even if previous one resulted in errors.
	s.tasks.RunExit()

}

// Run
func (s *Setup) Run() error {
	if err := s.Init(); err != nil {
		return err
	}
	select {}
}
