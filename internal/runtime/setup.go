// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import (
	"fmt"

	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/tasks"
)

type Setup struct {
	Config *SRv6Config
	tasks  map[string]tasks.Task
}

// Add tasks to setup
func (s *Setup) AddTasks() {
	// 0. user pre-hook
	s.tasks["hook.pre"] = NewMultiHook(Config.IPRoute2.PreInitHook, Config.IPRoute2.PreExitHook)

	// 1.  iproute2 config
	// 1.1 proto nextmn
	//	s.tasks["iproute2.config.proto.nextmn") //TODO
	// 1.2 rttable nextmn-gtp4
	//	s.tasks["iproute2.config.rttable.nextmn-gtp4") //TODO
	// 1.3 rttable nextmn-srv6
	//	s.tasks["iproute2.config.rttable.nextmn-srv6") //TODO

	// 2.  ifaces
	// 2.1 iface linux-srv6 (type dummy)
	//	s.tasks["iproute2.iface.linux-srv6") //TODO
	// 2.2 iface golang-srv6 (tun via water)
	//	s.tasks["nextmn.tun.golang-srv6") //TODO
	// 2.3 iface golang-gtp4 (tun via water)
	//	s.tasks["nextmn.tun.golang-gtp4") //TODO

	// 3.  ip routes
	// 3.1 blackhole route (srv6)
	//	s.tasks["iproute2.route.nextmn-srv6.blackhole") //TODO
	// 3.2 blackhole route (gtp4)
	//	s.tasks["iproute2.route.nextmn-gtp4.blackhole") //TODO
	// 3.3 routes to linux-srv6 endpoints (= endpoints themself)
	//	s.tasks["iproute2.routes.linux-srv6") //TODO
	// 3.4 routes to nextmn-gtp4 endpoints + endpoints
	//	s.tasks["nextmn.endpoints.gtp4") //TODO
	// 3.5 routes to nextmn-srv6 endpoints + endpoints
	//	s.tasks["nextmn.endpoints.srv6") //TODO

	// 4.  ip rules
	// 4.1 rule to rttable nextmn-srv6
	//	s.tasks["iproute2.rule.nextmn-srv6") //TODO
	// 4.2 rule to rttable nextmn-gtp4
	//	s.tasks["iproute2.rule.nextmn-gtp4") //TODO

	// 5. user post-hook
	s.tasks["hook.post"] = NewMultiHook(Config.IPRoute2.PostInitHook, Config.IPRoute2.PostExitHook)

}

// Runs init task by name
func (s *Setup) RunInitTask(name string) error {
	if s.tasks[name] != nil {
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

	// 1.  iproute2 config
	// 1.1 proto nextmn
	if err := s.RunInitTask("iproute2.config.proto.nextmn"); err != nil {
		return err
	}
	// 1.2 rttable nextmn-gtp4
	if err := s.RunInitTask("iproute2.config.rttable.nextmn-gtp4"); err != nil {
		return err
	}
	// 1.3 rttable nextmn-srv6
	if err := s.RunInitTask("iproute2.config.rttable.nextmn-srv6"); err != nil {
		return err
	}

	// 2.  ifaces
	// 2.1 iface linux-srv6 (type dummy)
	if err := s.RunInitTask("iproute2.iface.linux-srv6"); err != nil {
		return err
	}
	// 2.2 iface golang-srv6 (tun via water)
	if err := s.RunInitTask("nextmn.tun.golang-srv6"); err != nil {
		return err
	}
	// 2.3 iface golang-gtp4 (tun via water)
	if err := s.RunInitTask("nextmn.tun.golang-gtp4"); err != nil {
		return err
	}

	// 3.  ip routes
	// 3.1 blackhole route (srv6)
	if err := s.RunInitTask("iproute2.route.nextmn-srv6.blackhole"); err != nil {
		return err
	}
	// 3.2 blackhole route (gtp4)
	if err := s.RunInitTask("iproute2.route.nextmn-gtp4.blackhole"); err != nil {
		return err
	}
	// 3.3 routes to linux-srv6 endpoints (= endpoints themself)
	if err := s.RunInitTask("iproute2.routes.linux-srv6"); err != nil {
		return err
	}
	// 3.4 routes to nextmn-gtp4 endpoints + endpoints
	if err := s.RunInitTask("nextmn.endpoints.gtp4"); err != nil {
		return err
	}
	// 3.5 routes to nextmn-srv6 endpoints + endpoints
	if err := s.RunInitTask("nextmn.endpoints.srv6"); err != nil {
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
func (s *Setup) Exit() error {
	// 0. user pre-hook
	if err := s.RunExitTask("hook-pre"); err != nil {
		return err
	}
	// 1.  ip rules
	// 1.1 rule to rttable nextmn-gtp4
	if err := s.RunExitTask("iproute2.rule.nextmn-gtp4"); err != nil {
		return err
	}
	// 1.2 rule to rttable nextmn-srv6
	if err := s.RunExitTask("iproute2.rule.nextmn-srv6"); err != nil {
		return err
	}

	// 2.  ip routes
	// 2. routes to golang-gtp4 endpoints + endpoints
	if err := s.RunExitTask("nextmn.endpoints.gtp4"); err != nil {
		return err
	}
	// 2.2 routes to golang-srv6 endpoints + endpoints
	if err := s.RunExitTask("nextmn.endpoints.srv6"); err != nil {
		return err
	}
	// 2.3 routes to linux-srv6 endpoints (= endpoints themself)
	if err := s.RunExitTask("iproute2.routes.linux-srv6"); err != nil {
		return err
	}
	// 2.4 blackhole route (gtp4)
	if err := s.RunExitTask("iproute2.route.nextmn-gtp4.blackhole"); err != nil {
		return err
	}
	// 2.5 blackhole route (srv6)
	if err := s.RunExitTask("iproute2.route.nextmn-srv6.blackhole"); err != nil {
		return err
	}

	// 3.  create ifaces
	// 3.1 iface golang-gtp4 (tun via water)
	if err := s.RunExitTask("nextmn.tun.golang-gtp4"); err != nil {
		return err
	}
	// 3.2 iface golang-srv6 (tun via water)
	if err := s.RunExitTask("nextmn.tun.golang-srv6"); err != nil {
		return err
	}
	// 3.3 iface linux-srv6 (type dummy)
	if err := s.RunExitTask("iproute2.iface.linux-srv6"); err != nil {
		return err
	}

	// 4.  iproute2 config
	// 4.1 rttable nextmn-srv6
	if err := s.RunExitTask("iproute2.config.rttable.nextmn-srv6"); err != nil {
		return err
	}
	// 4.2 rttable nextmn-gtp4
	if err := s.RunExitTask("iproute2.config.rttable.nextmn-gtp4"); err != nil {
		return err
	}
	// 4.3 proto nextmn
	if err := s.RunExitTask("iproute2.config.proto.nextmn"); err != nil {
		return err
	}

	// 5. user post-hook
	if err := s.RunExitTask("hook-post"); err != nil {
		return err
	}

	return nil
}
