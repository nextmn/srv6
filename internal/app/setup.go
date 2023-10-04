// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"

	runtime "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/runtime"
	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/tasks"
	tasks_api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/tasks/api"
)

type Setup struct {
	config   *runtime.SRv6Config
	tasks    map[string]tasks_api.Task
	registry app_api.Registry
}

func NewSetup(config *runtime.SRv6Config) *Setup {
	return &Setup{
		config:   config,
		tasks:    make(map[string]tasks_api.Task),
		registry: NewRegistry(),
	}
}

// Add tasks to setup
func (s *Setup) AddTasks() {
	// 0. user pre-hook
	s.tasks["hook.pre"] = tasks.NewMultiHook(Config.IPRoute2.PreInitHook, Config.IPRoute2.PreExitHook)

	// 1.  ifaces
	// 1.1 iface linux-srv6 (type dummy)
	s.tasks["iproute2.iface.linux-srv6"] = tasks.NewTaskDummyIface(IFACE_LINUX_SRV6, s.registry)
	// 1.2 iface golang-srv6 (tun via water)
	s.tasks["nextmn.tun.golang-srv6"] = tasks.NewTaskTunIface(IFACE_GOLANG_SRV6, s.registry)
	// 1.3 iface golang-gtp4 (tun via water)
	s.tasks["nextmn.tun.golang-gtp4"] = tasks.NewTaskTunIface(IFACE_GOLANG_GTP4, s.registry)

	// 2.  ip routes
	// 2.1 blackhole route (srv6)
	s.tasks["iproute2.route.nextmn-srv6.blackhole"] = NewTaskBlackhole(runtime.RT_TABLE_NEXTMN_SRV6)
	// 2.2 blackhole route (gtp4)
	s.tasks["iproute2.route.nextmn-gtp4.blackhole"] = NewTaskBlackhole(runtime.RT_TABLE_NEXTMN_GTP4)

	// 3.  endpoints
	// 3.1 routes to linux-srv6 endpoints (= endpoints themself)
	s.tasks["iproute2.endpoints.linux-srv6"] = NewTaskLinuxEndpoints(Config.Endpoints.Filter(runtime.ProviderLinux), runtime.IFACE_LINUX_SRV6)
	// 3.2 routes to nextmn-srv6 endpoints + endpoints
	s.tasks["nextmn.endpoints.srv6"] = NewTaskLinuxEndpoints(Config.Endpoints.Filter(runtime.ProviderNextMNSRv6), runtime.IFACE_GOLANG_SRV6)
	// 3.3 routes to nextmn-gtp4 endpoints + endpoints
	s.tasks["nextmn.endpoints.gtp4"] = NewTaskLinuxEndpoints(Config.Endpoints.Filter(runtime.ProviderNextMNGTP4), runtime.IFACE_GOLANG_GTP4)

	// 4.  ip rules
	// 4.1 rule to rttable nextmn-srv6
	if Config.Locator != nil {
		s.tasks["iproute2.rule.nextmn-srv6"] = NewTaskIPRule(Config.Locator, runtime.RT_TABLE_NEXTMN_SRV6)
	}
	// 4.2 rule to rttable nextmn-gtp4
	if Config.IPv4HeadendPrefix != nil {
		s.tasks["iproute2.rule.nextmn-gtp4"] = NewTaskIPRule(Config.IPv4HeadendPrefix, runtime.RT_TABLE_NEXTMN_GTP4)
	}

	// 5. user post-hook
	s.tasks["hook.post"] = NewMultiHook(Config.IPRoute2.PostInitHook, Config.IPRoute2.PostExitHook)

}

// Runs init task by name
func (s *Setup) RunInitTask(name string) error {
	if s.tasks[name] != nil {
		if s.task[name].State() {
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
		if !s.task[name].State() {
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

	// 3.  endpoints
	// 3.1 routes to linux-srv6 endpoints (= endpoints themself)
	if err := s.RunInitTask("iproute2.endpoints.linux-srv6"); err != nil {
		return err
	}
	// 3.2 routes to nextmn-srv6 endpoints + endpoints
	if err := s.RunInitTask("nextmn.endpoints.srv6"); err != nil {
		return err
	}
	// 3.3 routes to nextmn-gtp4 endpoints + endpoints
	if err := s.RunInitTask("nextmn.endpoints.gtp4"); err != nil {
		return err
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

	// 2  endpoints
	// 2. routes to golang-gtp4 endpoints + endpoints
	if err := s.RunExitTask("nextmn.endpoints.gtp4"); err != nil {
		fmt.Println(err)
	}
	// 2.2 routes to golang-srv6 endpoints + endpoints
	if err := s.RunExitTask("nextmn.endpoints.srv6"); err != nil {
		fmt.Println(err)
	}
	// 2.3 routes to linux-srv6 endpoints (= endpoints themself)
	if err := s.RunExitTask("iproute2.endpoints.linux-srv6"); err != nil {
		fmt.Println(err)
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

	return nil
}

// Run
func (s *Setup) Run() error {
	s.Init()
	select {}
}
