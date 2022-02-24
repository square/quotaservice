// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"io"
	"os"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/square/quotaservice/config/zkhelpers"
	pb "github.com/square/quotaservice/protos/config"
	"github.com/square/quotaservice/test/helpers"
	"github.com/square/quotaservice/test/zktestserver"
)

var servers []string

func TestMain(m *testing.M) {
	t, err := zktestserver.StartTestCluster(1, nil, nil)
	helpers.PanicError(err)

	defer func() { _ = t.Stop() }()
	servers = make([]string, 1)
	servers[0] = fmt.Sprintf("localhost:%d", t.Servers[0].Port)

	os.Exit(m.Run())
}

func TestNewPathError(t *testing.T) {
	conn := setup()

	p, err := NewZkConfigPersisterWithConnection("/LOCAL/quotaservice", conn)

	if err == nil {
		defer p.Close()
		t.Error("Should have received error on new because path does not exit")
	}
}

func TestNew(t *testing.T) {
	conn := setup()

	p, err := NewZkConfigPersisterWithConnection("/quotaservice", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	select {
	case <-p.ConfigChangedWatcher():
	// this is good
	default:
		t.Error("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	if cfg != nil {
		t.Error("Received non-nil cfg on new node")
	}
}

func TestNewExisting(t *testing.T) {
	conn := setup("/existing")

	p, err := NewZkConfigPersisterWithConnection("/existing", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	select {
	case <-p.ConfigChangedWatcher():
	// this is good
	default:
		t.Fatal("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	if cfg.Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", cfg)
	}
}

func TestSetExisting(t *testing.T) {
	conn := setup("/existing")

	p, err := NewZkConfigPersisterWithConnection("/existing", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	select {
	case <-p.ConfigChangedWatcher():
	// this is good
	default:
		t.Fatal("Config channel should not be empty!")
	}

	cfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)
	helpers.CheckError(t, p.PersistAndNotify("", cfg))
}

func TestSetAndNotify(t *testing.T) {
	conn := setup("/existing")

	p, err := NewZkConfigPersisterWithConnection("/existing", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	<-p.ConfigChangedWatcher()

	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["foo"] = NewDefaultNamespaceConfig("foo")
	cfg.Version++

	helpers.CheckError(t, p.PersistAndNotify("", cfg))

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	if ioCfg.Namespaces["foo"] == nil {
		t.Errorf("Config is not valid: %+v", ioCfg)
	}

	cfg.Namespaces["bar"] = NewDefaultNamespaceConfig("bar")
	cfg.Version++
	helpers.CheckError(t, p.PersistAndNotify("", cfg))

	select {
	case <-p.ConfigChangedWatcher():
	case <-time.After(time.Second * 1):
		t.Fatalf("Did not receive notification!")
	}

	ioCfg, err = p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	if ioCfg.Namespaces["bar"] == nil {
		t.Errorf("Config is not valid: %+v", ioCfg)
	}
}

func TestHistoricalConfigs(t *testing.T) {
	conn := setup("/historic")

	p, err := NewZkConfigPersisterWithConnection("/historic", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	waitOrTimeout(p.ConfigChangedWatcher(), time.Minute)

	cfgs, err := p.ReadHistoricalConfigs()
	helpers.CheckError(t, err)

	if len(cfgs) != 1 {
		t.Fatalf("Historical configs are not correct: %+v", cfgs)
	}

	if cfgs[0].Namespaces["test"] == nil {
		t.Errorf("Config is not valid: %+v", cfgs[0])
	}
}

func TestReadingStaleVersions(t *testing.T) {
	conn := setup("/conflicting")

	p, err := NewZkConfigPersisterWithConnection("/conflicting", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	origCfg := NewDefaultServiceConfig()
	origCfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")
	origCfg.Version = 200

	persistOrPanic(p, origCfg)

	// Two changes
	cfg1 := CloneConfig(origCfg)
	cfg1.Namespaces["test_new"] = NewDefaultNamespaceConfig("test_new")
	cfg1.Version = origCfg.Version + 1

	cfg2 := CloneConfig(origCfg)
	cfg2.Namespaces["test_newer"] = NewDefaultNamespaceConfig("test_newer")
	cfg2.Version = origCfg.Version + 2

	// Persist in quick succession
	persistOrPanic(p, cfg1)
	persistOrPanic(p, cfg2)

	// Wait for callbacks to re-read persisted configs
	waitOrTimeout(p.ConfigChangedWatcher(), time.Minute)
	waitOrTimeout(p.ConfigChangedWatcher(), time.Minute)
	waitOrTimeout(p.ConfigChangedWatcher(), time.Minute)

	latest, err := p.ReadPersistedConfig()
	helpers.CheckError(t, err)

	if latest.Version != 202 {
		t.Fatalf("Expected latest version = 202, got %v", latest.Version)
	}
}

// TestConcurrentUpdate attempts to recreate a case where two concurrent readers read a version, attempt an update to
// the next version, and both succeed, creating potential for lost changes.
func TestConcurrentUpdate(t *testing.T) {
	conn := setup("/lost_changes")

	p, err := NewZkConfigPersisterWithConnection("/lost_changes", conn)
	helpers.CheckError(t, err)

	defer p.Close()

	origCfg := NewDefaultServiceConfig()
	origCfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")
	origCfg.Version = 200

	persistOrPanic(p, origCfg)

	// Both cfg1 and cfg2 are configs derived from origConfig, but diverging
	cfg1 := CloneConfig(origCfg)
	cfg1.Namespaces["test_new"] = NewDefaultNamespaceConfig("test_new")
	cfg1.Version = origCfg.Version + 1

	persistOrPanic(p, cfg1)

	cfg2 := CloneConfig(origCfg)
	cfg2.Namespaces["test_newer"] = NewDefaultNamespaceConfig("test_newer")
	cfg2.Version = origCfg.Version + 1

	err = p.PersistAndNotify("", cfg2)

	// TODO(manik) clobbering is currently possible, this test fails. Uncomment assertion below and commit along with fix.
	//helpers.ExpectingError(t, err)
}

func createExistingNode(zkConn *zk.Conn, path string) {
	cfg := NewDefaultServiceConfig()
	cfg.Namespaces["test"] = NewDefaultNamespaceConfig("test")

	storeInZkOrPanic(path, zkConn, cfg)
}

func persistOrPanic(p ConfigPersister, cfg *pb.ServiceConfig) {
	helpers.PanicError(p.PersistAndNotify("", cfg))
}

func storeInZkOrPanic(pathPrefix string, conn *zk.Conn, cfg *pb.ServiceConfig) (zkPath string) {
	r := marshallOrPanic(cfg)

	bytes, err := ioutil.ReadAll(r)
	helpers.PanicError(err)

	key := HashConfig(cfg)

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

func waitOrTimeout(c <-chan struct{}, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	select {
	case <-c:
	case <-ticker.C:
	}
}

// setup performs pre-flight setup, including cleaning out stale state in Zookeeper. This should be called at the start
// of every test. The Zookeeper client connection returned can be used during the test, and must be closed at the end
// of the test.
func setup(nodesToCreate ...string) *zk.Conn {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	helpers.PanicError(err)

	// Empty out the contents of Zookeeper
	err = zkhelpers.DeleteRecursively(conn, zkhelpers.DefaultRoot)
	helpers.PanicError(err)

	for _, node := range nodesToCreate {
		createExistingNode(conn, node)
	}

	return conn
}
