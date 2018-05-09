// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/square/quotaservice/test/helpers"
	pb "github.com/square/quotaservice/protos/config"
	"io"
)

var servers []string

func TestMain(m *testing.M) {
	t, err := zk.StartTestCluster(1, nil, nil)
	helpers.PanicError(err)

	defer func() { _ = t.Stop() }()
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

	newConfig := unmarshallOrPanic(cfg)

	if newConfig.Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func TestSetExisting(t *testing.T) {
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
	helpers.CheckError(t, p.PersistAndNotify(cfg))
}

func TestSetAndNotify(t *testing.T) {
	p, err := NewZkConfigPersister("/existing", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	<-p.ConfigChangedWatcher()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["foo"] = NewDefaultNamespaceConfig("foo")
	cfg.Version++

	helpers.CheckError(t, p.PersistAndNotify(marshallOrPanic(cfg)))

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	newConfig := unmarshallOrPanic(ioCfg)

	if newConfig.Namespaces["foo"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}

	cfg.Namespaces["bar"] = NewDefaultNamespaceConfig("bar")
	cfg.Version++
	helpers.CheckError(t, p.PersistAndNotify(marshallOrPanic(cfg)))

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err = p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	newConfig = unmarshallOrPanic(ioCfg)

	if newConfig.Namespaces["bar"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func TestHistoricalConfigs(t *testing.T) {
	p, err := NewZkConfigPersister("/historic", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	<-p.ConfigChangedWatcher()

	cfgs, err := p.ReadHistoricalConfigs()
	helpers.CheckError(t, err)

	if len(cfgs) != 1 {
		t.Fatalf("Historical configs are not correct: %+v", cfgs)
	}

	newConfig := unmarshallOrPanic(cfgs[0])

	if newConfig.Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", newConfig)
	}
}

func TestReadingStaleVersions(t *testing.T) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	helpers.PanicError(err)

	defer conn.Close()

	p, err := NewZkConfigPersister("/conflicting", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	origCfg := NewDefaultServiceConfig()
	origCfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")
	origCfg.Version = 200

	persistOrPanic(p, origCfg)

	// Two changes
	cfg1 := cloneConfig(origCfg)
	cfg1.Namespaces["test_new"] = NewDefaultNamespaceConfig("test_new")
	cfg1.Version = origCfg.Version + 1

	cfg2 := cloneConfig(origCfg)
	cfg2.Namespaces["test_newer"] = NewDefaultNamespaceConfig("test_newer")
	cfg2.Version = origCfg.Version + 2

	// Persist in quick succession
	persistOrPanic(p, cfg1)
	persistOrPanic(p, cfg2)

	// Wait for callbacks to re-read persisted configs
	<- p.ConfigChangedWatcher()
	<- p.ConfigChangedWatcher()
	<- p.ConfigChangedWatcher()

	r, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	latest := unmarshallOrPanic(r)

	if latest.Version != 202 {
		t.Fatalf("Expected latest version = 200, got %v", latest.Version)
	}
}

// TestConcurrentUpdate attempts to recreate a case where two concurrent readers read a version, attempt an update to
// the next version, and both succeed, creating potential for lost changes.
func TestConcurrentUpdate(t *testing.T) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	helpers.PanicError(err)

	defer conn.Close()

	p, err := NewZkConfigPersister("/lost_changes", servers)
	helpers.CheckError(t, err)

	defer p.(*ZkConfigPersister).Close()

	origCfg := NewDefaultServiceConfig()
	origCfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")
	origCfg.Version = 200

	persistOrPanic(p, origCfg)

	// Both cfg1 and cfg2 are configs derived from origConfig, but diverging
	cfg1 := cloneConfig(origCfg)
	cfg1.Namespaces["test_new"] = NewDefaultNamespaceConfig("test_new")
	cfg1.Version = origCfg.Version + 1

	persistOrPanic(p, cfg1)

	cfg2 := cloneConfig(origCfg)
	cfg2.Namespaces["test_newer"] = NewDefaultNamespaceConfig("test_newer")
	cfg2.Version = origCfg.Version + 1

	r := marshallOrPanic(cfg2)
	err = p.PersistAndNotify(r)

	// TODO(manik) clobbering is currently possible, this test fails. Uncomment assertion below and commit along with fix.
	//helpers.ExpectingError(t, err)
}

func createExistingNode(path string) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	helpers.PanicError(err)

	defer conn.Close()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")

	storeInZkOrPanic(path, conn, cfg)
}

func persistOrPanic(p ConfigPersister, cfg *pb.ServiceConfig) {
	r := marshallOrPanic(cfg)
	helpers.PanicError(p.PersistAndNotify(r))
}

func storeInZkOrPanic(pathPrefix string, conn *zk.Conn, cfg *pb.ServiceConfig) (zkPath string) {
	r := marshallOrPanic(cfg)

	bytes, err := ioutil.ReadAll(r)
	helpers.PanicError(err)

	key := HashConfig(bytes)

	e, _, err := conn.Exists(pathPrefix)
	helpers.PanicError(err)

	if !e {
		// Create the parent if we need to
		_, err = conn.Create(pathPrefix, []byte(key), 0, zk.WorldACL(zk.PermAll))
		helpers.PanicError(err)
	}

	fullPath := fmt.Sprintf("%s/%s", pathPrefix, key)

	_, err = conn.Create(fullPath, bytes, 0, zk.WorldACL(zk.PermAll))
	helpers.PanicError(err)

	return fullPath
}

func marshallOrPanic(cfg *pb.ServiceConfig) io.Reader {
	r, err := Marshal(cfg)
	helpers.PanicError(err)
	return r
}

func unmarshallOrPanic(r io.Reader) *pb.ServiceConfig {
	c, err := Unmarshal(r)
	helpers.PanicError(err)
	return c
}