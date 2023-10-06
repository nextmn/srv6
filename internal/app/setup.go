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
	tasks "github.com/nextmn/srv6/internal/tasks"
	tasks_api "github.com/nextmn/srv6/internal/tasks/api"
)

type Setup struct {
	config   *config.SRv6Config
	tasks    map[string]tasks_api.Task
	registry app_api.Registry
}

func NewSetup(config *config.SRv6Config) *Setup {
	return &Setup{
		config:   config,
		tasks:    make(map[string]tasks_api.Task),
		registry: NewRegistry(),
	}
}

// Add tasks to setup
func (s *Setup) AddTasks() {
	// 0.  user hooks
	// 0.1 pre-hooks
	s.tasks["hook.pre"] = tasks.NewMultiHook(s.config.IPRoute2.PreInitHook, s.config.IPRoute2.PreExitHook)
	// 0.2 post-hooks
	s.tasks["hook.post"] = tasks.NewMultiHook(s.config.IPRoute2.PostInitHook, s.config.IPRoute2.PostExitHook)

	// 1.  ifaces
	// 1.1 iface linux-srv6 (type dummy)
	s.tasks["iproute2.iface.linux-srv6"] = tasks.NewTaskDummyIface(constants.IFACE_LINUX_SRV6)
	// 1.2 iface golang-srv6 (tun via water)
	s.tasks["nextmn.tun.golang-srv6"] = tasks.NewTaskTunIface(constants.IFACE_GOLANG_SRV6, s.registry)
	// 1.3 iface golang-gtp4 (tun via water)
	s.tasks["nextmn.tun.golang-gtp4"] = tasks.NewTaskTunIface(constants.IFACE_GOLANG_GTP4, s.registry)

	// 2.  ip routes
	// 2.1 blackhole route (srv6)
	s.tasks["iproute2.route.nextmn-srv6.blackhole"] = tasks.NewTaskBlackhole(constants.RT_TABLE_NEXTMN_SRV6)
	// 2.2 blackhole route (gtp4)
	s.tasks["iproute2.route.nextmn-gtp4.blackhole"] = tasks.NewTaskBlackhole(constants.RT_TABLE_NEXTMN_GTP4)

	// 3.  endpoints + headends
	// 3.1 linux-srv6 headends
	for _, h := range s.config.Headends.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("iproute2.headend.linux-srv6/%s", h.Name)
		s.tasks[t_name] = tasks.NewFakeTask() // FIXME
		//s.tasks[t_name] = tasks.NewTaskLinuxHeadend(e, constants.RT_TABLE_MAIN, constants.IFACE_LINUX_SRV6)
	}
	// 3.1 linux-srv6 endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("iproute2.endpoint.linux-srv6/%s", e.Sid)
		s.tasks[t_name] = tasks.NewTaskLinuxEndpoint(e, constants.RT_TABLE_NEXTMN_SRV6, constants.IFACE_LINUX_SRV6)
	}
	// 3.2 nextmn-srv6 endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.endpoint.srv6/%s", e.Sid)
		s.tasks[t_name] = tasks.NewFakeTask() //FIXME
		//s.tasks[t_name] = tasks.NewTaskGolangSRv6Endpoint(e, constants.RT_TABLE_NEXTMN_SRV6, contants.IFACE_GOLANGE_SRV6, s.registry)
	}
	// 3.3 nextmn-gtp4 headends
	for _, h := range s.config.Headends.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.headend.gtp4/%s", h.Name)
		s.tasks[t_name] = tasks.NewFakeTask() // FIXME
		//s.tasks[t_name] = tasks.NewTaskGolangHeadend(e, constants.RT_TABLE_GTP4, constants.IFACE_GOLANG_GTP4, s.registry)
	}

	// 4.  ip rules
	// 4.1 rule to rttable nextmn-srv6
	if s.config.Locator != nil {
		s.tasks["iproute2.rule.nextmn-srv6"] = tasks.NewTaskIP6Rule(*s.config.Locator, constants.RT_TABLE_NEXTMN_SRV6)
	}
	// 4.2 rule to rttable nextmn-gtp4
	if s.config.GTP4HeadendPrefix != nil {
		s.tasks["iproute2.rule.nextmn-gtp4"] = tasks.NewTaskIP4Rule(*s.config.GTP4HeadendPrefix, constants.RT_TABLE_NEXTMN_GTP4)
	}
}

// Runs init task by name
func (s *Setup) RunInitTask(name string) error {
	if s.tasks[name] != nil {
		if s.tasks[name].State() {
			// nothing to do
			return nil
		}
		if err := s.tasks[name].RunInit(); err != nil {
			return fmt.Errorf("[Failed] %s.init: %s", name, err)
		}
		fmt.Printf("[OK] %s.init%s\n", name)
		return nil
	}
	return fmt.Errorf("Unknown task: %s", name)
}

// Runs exist task by name
func (s *Setup) RunExitTask(name string) error {
	if s.tasks[name] != nil {
		if !s.tasks[name].State() {
			// nothing to do
			return nil
		}
		if err := s.tasks[name].RunExit(); err != nil {
			return fmt.Errorf("[Failed] %s.exit: %s", name, err)
		}
		fmt.Printf("[OK] %s.exit\n", name)
		return nil
	}
	return fmt.Errorf("Unknown task: %s", name)
}

// Init
func (s *Setup) Init() error {
	// 0. user pre-hook
	if err := s.RunInitTask("hook.pre"); err != nil {
		return err
	}

	// 1.  ifaces
	// 1.1 iface linux-srv6 (type dummy)
	if err := s.RunInitTask("iproute2.iface.linux-srv6"); err != nil {
		return err
	}
	// 1.2 iface golang-srv6 (tun via water)
	if err := s.RunInitTask("nextmn.tun.golang-srv6"); err != nil {
		return err
	}
	// 1.3 iface golang-gtp4 (tun via water)
	if err := s.RunInitTask("nextmn.tun.golang-gtp4"); err != nil {
		return err
	}

	// 2.  ip routes
	// 2.1 blackhole route (srv6)
	if err := s.RunInitTask("iproute2.route.nextmn-srv6.blackhole"); err != nil {
		return err
	}
	// 2.2 blackhole route (gtp4)
	if err := s.RunInitTask("iproute2.route.nextmn-gtp4.blackhole"); err != nil {
		return err
	}

	// 3.  endpoints + headends
	// 3.1 linux-srv6 headends
	for _, h := range s.config.Headends.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("iproute2.headend.linux-srv6/%s", h.Name)
		if err := s.RunInitTask(t_name); err != nil {
			return err
		}
	}
	// 3.2 linux-srv6 endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("iproute2.endpoint.linux-srv6/%s", e.Sid)
		if err := s.RunInitTask(t_name); err != nil {
			return err
		}
	}
	// 3.3 nextmn-srv6 endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.endpoint.srv6/%s", e.Sid)
		if err := s.RunInitTask(t_name); err != nil {
			return err
		}
	}
	// 3.4 nextmn-gtp4 headends
	for _, h := range s.config.Headends.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.headend.gtp4/%s", h.Name)
		if err := s.RunInitTask(t_name); err != nil {
			return err
		}
	}

	// 4.  ip rules
	// 4.1 rule to rttable nextmn-srv6
	if err := s.RunInitTask("iproute2.rule.nextmn-srv6"); err != nil {
		return err
	}
	// 4.2 rule to rttable nextmn-gtp4
	if err := s.RunInitTask("iproute2.rule.nextmn-gtp4"); err != nil {
		return err
	}

	// 5. user post-hook
	if err := s.RunInitTask("hook-post"); err != nil {
		return err
	}

	return nil
}

// Exit
func (s *Setup) Exit() {
	// This function may be called at any time,
	// and a maximum of exit tasks must be run,
	// even if previous one resulted in errors.

	// 0. user pre-hook
	if err := s.RunExitTask("hook-pre"); err != nil {
		fmt.Println(err)
	}
	// 1.  ip rules
	// 1.1 rule to rttable nextmn-gtp4
	if err := s.RunExitTask("iproute2.rule.nextmn-gtp4"); err != nil {
		fmt.Println(err)
	}
	// 1.2 rule to rttable nextmn-srv6
	if err := s.RunExitTask("iproute2.rule.nextmn-srv6"); err != nil {
		fmt.Println(err)
	}

	// 2  endpoints + headends
	// 2. golang-gtp4 headends
	for _, h := range s.config.Headends.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.headend.gtp4/%s", h.Name)
		if err := s.RunExitTask(t_name); err != nil {
			fmt.Println(err)
		}
	}
	// 2.2 golang-srv6 endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderNextMN) {
		t_name := fmt.Sprintf("nextmn.endpoint.srv6/%s", e.Sid)
		if err := s.RunExitTask(t_name); err != nil {
			fmt.Println(err)
		}
	}
	// 2.3 linux-srv6 endpoints
	for _, e := range s.config.Endpoints.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("iproute2.endpoint.linux-srv6/%s", e.Sid)
		if err := s.RunExitTask(t_name); err != nil {
			fmt.Println(err)
		}
	}
	// 2.3 linux-srv6 headends
	for _, h := range s.config.Headends.Filter(config.ProviderLinux) {
		t_name := fmt.Sprintf("iproute2.headend.linux-srv6/%s", h.Name)
		if err := s.RunExitTask(t_name); err != nil {
			fmt.Println(err)
		}
	}

	// 3.  ip routes
	// 3.1 blackhole route (gtp4)
	if err := s.RunExitTask("iproute2.route.nextmn-gtp4.blackhole"); err != nil {
		fmt.Println(err)
	}
	// 3.2 blackhole route (srv6)
	if err := s.RunExitTask("iproute2.route.nextmn-srv6.blackhole"); err != nil {
		fmt.Println(err)
	}

	// 4.  ifaces
	// 4.1 iface golang-gtp4 (tun via water)
	if err := s.RunExitTask("nextmn.tun.golang-gtp4"); err != nil {
		fmt.Println(err)
	}
	// 4.2 iface golang-srv6 (tun via water)
	if err := s.RunExitTask("nextmn.tun.golang-srv6"); err != nil {
		fmt.Println(err)
	}
	// 4.3 iface linux-srv6 (type dummy)
	if err := s.RunExitTask("iproute2.iface.linux-srv6"); err != nil {
		fmt.Println(err)
	}

	// 5. user post-hook
	if err := s.RunExitTask("hook-post"); err != nil {
		fmt.Println(err)
	}
}

// Run
func (s *Setup) Run() error {
	s.Init()
	select {}
}
