// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/maniksurtani/quotaservice/test/helpers"
	"github.com/samuel/go-zookeeper/zk"
)

var servers []string

func TestMain(m *testing.M) {
	t, err := zk.StartTestCluster(1, nil, nil)
	helpers.PanicError(err)

	defer t.Stop()
	servers = make([]string, 1)
	servers[0] = fmt.Sprintf("localhost:%d", t.Servers[0].Port)

	createExistingNode("/existing")
	createExistingNode("/historic")
	r := m.Run()

	os.Exit(r)
}

func TestNewPathError(t *testing.T) {
	p, err := NewZkConfigPersister("/LOCAL/quotaservice", servers)

	if err == nil {
		t.Error("Should have received error on new because path does not exit")
		p.(*ZkConfigPersister).Close()
	}
}

func TestNew(t *testing.T) {
	p, err := NewZkConfigPersister("/quotaservice", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	select {
	case <-p.ConfigChangedWatcher():
	// this is good
	default:
		t.Error("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	cfgArray, err := ioutil.ReadAll(cfg)
	helpers.CheckError(t, err)

	if len(cfgArray) > 0 {
		t.Errorf("Received non-empty cfg on new node: %+v", cfgArray)
	}
}

func TestNewExisting(t *testing.T) {
	p, err := NewZkConfigPersister("/existing", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	select {
	case <-p.ConfigChangedWatcher():
	// this is good
	default:
		t.Fatal("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	newConfig, err := Unmarshal(cfg)
	helpers.CheckError(t, err)

	if newConfig.Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func TestSetAndNotify(t *testing.T) {
	p, err := NewZkConfigPersister("/existing", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	<-p.ConfigChangedWatcher()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["foo"] = NewDefaultNamespaceConfig("foo")

	r, err := Marshal(cfg)
	helpers.CheckError(t, err)

	p.PersistAndNotify(r)

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	newConfig, err := Unmarshal(ioCfg)
	helpers.CheckError(t, err)

	if newConfig.Namespaces["foo"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}

	cfg.Namespaces["bar"] = NewDefaultNamespaceConfig("bar")

	r, err = Marshal(cfg)
	helpers.CheckError(t, err)

	p.PersistAndNotify(r)

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err = p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	newConfig, err = Unmarshal(ioCfg)
	helpers.CheckError(t, err)

	if newConfig.Namespaces["bar"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func TestHistoricalConfigs(t *testing.T) {
	p, err := NewZkConfigPersister("/historic", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	<-p.ConfigChangedWatcher()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["foo"] = NewDefaultNamespaceConfig("foo")

	r, err := Marshal(cfg)
	helpers.CheckError(t, err)

	p.PersistAndNotify(r)

	// There is no signal for children being updated,
	// so we just have to sleep for a bit here
	time.Sleep(time.Second * 1)

	cfgs, err := p.ReadHistoricalConfigs()
	helpers.CheckError(t, err)

	if len(cfgs) != 1 {
		t.Fatalf("Historical configs are not correct: %+v", cfgs)
	}

	newConfig, err := Unmarshal(cfgs[0])
	helpers.CheckError(t, err)

	if newConfig.Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func createExistingNode(path string) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	helpers.PanicError(err)

	defer conn.Close()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")

	reader, err := Marshal(cfg)
	helpers.PanicError(err)

	bytes, err := ioutil.ReadAll(reader)
	helpers.PanicError(err)

	_, err = conn.Create(path, bytes, 0, zk.WorldACL(zk.PermAll))
	helpers.PanicError(err)
}
