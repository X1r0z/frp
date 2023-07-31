// Copyright 2018 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package visitor

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/transport"
	"github.com/fatedier/frp/pkg/util/xlog"
)

type Manager struct {
	clientCfg config.ClientCommonConf
	cfgs      map[string]config.VisitorConf
	visitors  map[string]Visitor
	helper    Helper

	checkInterval time.Duration

	mu  sync.RWMutex
	ctx context.Context

	stopCh chan struct{}
}

func NewManager(
	ctx context.Context,
	runID string,
	clientCfg config.ClientCommonConf,
	connectServer func() (net.Conn, error),
	msgTransporter transport.MessageTransporter,
) *Manager {
	m := &Manager{
		clientCfg:     clientCfg,
		cfgs:          make(map[string]config.VisitorConf),
		visitors:      make(map[string]Visitor),
		checkInterval: 10 * time.Second,
		ctx:           ctx,
		stopCh:        make(chan struct{}),
	}
	m.helper = &visitorHelperImpl{
		connectServerFn: connectServer,
		msgTransporter:  msgTransporter,
		transferConnFn:  m.TransferConn,
		runID:           runID,
	}
	return m
}

func (vm *Manager) Run() {
	xl := xlog.FromContextSafe(vm.ctx)

	ticker := time.NewTicker(vm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-vm.stopCh:
			xl.Info("gracefully shutdown visitor manager")
			return
		case <-ticker.C:
			vm.mu.Lock()
			for _, cfg := range vm.cfgs {
				name := cfg.GetBaseConfig().ProxyName
				if _, exist := vm.visitors[name]; !exist {
					xl.Info("try to start visitor [%s]", name)
					_ = vm.startVisitor(cfg)
				}
			}
			vm.mu.Unlock()
		}
	}
}

func (vm *Manager) Close() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	for _, v := range vm.visitors {
		v.Close()
	}
	select {
	case <-vm.stopCh:
	default:
		close(vm.stopCh)
	}
}

// Hold lock before calling this function.
func (vm *Manager) startVisitor(cfg config.VisitorConf) (err error) {
	xl := xlog.FromContextSafe(vm.ctx)
	name := cfg.GetBaseConfig().ProxyName
	visitor := NewVisitor(vm.ctx, cfg, vm.clientCfg, vm.helper)
	err = visitor.Run()
	if err != nil {
		xl.Warn("start error: %v", err)
	} else {
		vm.visitors[name] = visitor
		xl.Info("start visitor success")
	}
	return
}

func (vm *Manager) Reload(cfgs map[string]config.VisitorConf) {
	xl := xlog.FromContextSafe(vm.ctx)
	vm.mu.Lock()
	defer vm.mu.Unlock()

	delNames := make([]string, 0)
	for name, oldCfg := range vm.cfgs {
		del := false
		cfg, ok := cfgs[name]
		if !ok || !reflect.DeepEqual(oldCfg, cfg) {
			del = true
		}

		if del {
			delNames = append(delNames, name)
			delete(vm.cfgs, name)
			if visitor, ok := vm.visitors[name]; ok {
				visitor.Close()
			}
			delete(vm.visitors, name)
		}
	}
	if len(delNames) > 0 {
		xl.Info("visitor removed: %v", delNames)
	}

	addNames := make([]string, 0)
	for name, cfg := range cfgs {
		if _, ok := vm.cfgs[name]; !ok {
			vm.cfgs[name] = cfg
			addNames = append(addNames, name)
			_ = vm.startVisitor(cfg)
		}
	}
	if len(addNames) > 0 {
		xl.Info("visitor added: %v", addNames)
	}
}

// TransferConn transfers a connection to a visitor.
func (vm *Manager) TransferConn(name string, conn net.Conn) error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	v, ok := vm.visitors[name]
	if !ok {
		return fmt.Errorf("visitor [%s] not found", name)
	}
	return v.AcceptConn(conn)
}

type visitorHelperImpl struct {
	connectServerFn func() (net.Conn, error)
	msgTransporter  transport.MessageTransporter
	transferConnFn  func(name string, conn net.Conn) error
	runID           string
}

func (v *visitorHelperImpl) ConnectServer() (net.Conn, error) {
	return v.connectServerFn()
}

func (v *visitorHelperImpl) TransferConn(name string, conn net.Conn) error {
	return v.transferConnFn(name, conn)
}

func (v *visitorHelperImpl) MsgTransporter() transport.MessageTransporter {
	return v.msgTransporter
}

func (v *visitorHelperImpl) RunID() string {
	return v.runID
}
