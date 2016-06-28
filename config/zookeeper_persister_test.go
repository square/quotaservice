// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var t zk.TestCluster
var servers []string

func TestMain(m *testing.M) {
	t, err := zk.StartTestCluster(1, nil, nil)

	if err != nil {
		panic(err)
	}

	defer t.Stop()
	servers = make([]string, 1)
	servers[0] = fmt.Sprintf("localhost:%d", t.Servers[0].Port)

	createExistingNode()
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

	if err != nil {
		t.Fatalf("Received error on create: %+v", err)
	}

	defer p.(*ZkConfigPersister).Close()

	select {
	case <-p.ConfigChangedWatcher():
		// this is good
	default:
		t.Error("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()

	if err != nil {
		t.Fatal("Received error on ReadPersistedConfig")
	}

	cfgArray, err := ioutil.ReadAll(cfg)

	if err != nil {
		t.Fatal("Received error on ReadAll")
	}

	if len(cfgArray) > 0 {
		t.Errorf("Received non-empty cfg on new node: %+v", cfgArray)
	}
}

func TestNewExisting(t *testing.T) {
	p, err := NewZkConfigPersister("/existing", servers)

	if err != nil {
		t.Fatalf("Received error on create: %+v", err)
	}

	defer p.(*ZkConfigPersister).Close()

	select {
	case <-p.ConfigChangedWatcher():
		// this is good
	default:
		t.Fatal("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()

	if err != nil {
		t.Fatalf("Received error on ReadPersistedConfig: %+v", err)
	}

	newConfig, err := Unmarshal(cfg)

	if err != nil {
		t.Fatalf("Received error on Unmarshal: %+v", err)
	}

	if newConfig.Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func TestSetAndNotify(t *testing.T) {
	p, err := NewZkConfigPersister("/existing", servers)

	if err != nil {
		t.Fatalf("Received error on create: %+v", err)
	}

	defer p.(*ZkConfigPersister).Close()
	<-p.ConfigChangedWatcher()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["foo"] = NewDefaultNamespaceConfig("foo")

	r, err := Marshal(cfg)

	if err != nil {
		t.Fatalf("Received error on marshal: %+v", err)
	}

	p.PersistAndNotify(r)

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err := p.ReadPersistedConfig()

	if err != nil {
		t.Fatalf("Received error on ReadPersistedConfig: %+v", err)
	}

	newConfig, err := Unmarshal(ioCfg)

	if err != nil {
		t.Fatalf("Received error on Unmarshal: %+v", err)
	}

	if newConfig.Namespaces["foo"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func createExistingNode() {
	conn, _, err := zk.Connect(servers, SessionTimeout)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")

	reader, err := Marshal(cfg)

	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(reader)

	if err != nil {
		panic(err)
	}

	_, err = conn.Create("/existing", bytes, 0, zk.WorldACL(zk.PermAll))

	if err != nil {
		panic(err)
	}
}
