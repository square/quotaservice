package zktestserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/go-zookeeper/zk"
)

type TestServer struct {
	Port int
	Path string
	Srv  *Server
}

type TestCluster struct {
	Path    string
	Servers []TestServer
}

func StartTestCluster(size int, stdout, stderr io.Writer) (*TestCluster, error) {
	tmpPath, err := ioutil.TempDir("", "gozk")
	if err != nil {
		return nil, err
	}
	success := false
	startPort := int(rand.Int31n(6000) + 10000)
	cluster := &TestCluster{Path: tmpPath}
	defer func() {
		if !success {
			cluster.Stop()
		}
	}()
	for serverN := 0; serverN < size; serverN++ {
		srvPath := filepath.Join(tmpPath, fmt.Sprintf("srv%d", serverN))
		if err := os.Mkdir(srvPath, 0700); err != nil {
			return nil, err
		}
		port := startPort + serverN*3
		cfg := ServerConfig{
			ClientPort: port,
			DataDir:    srvPath,
		}
		for i := 0; i < size; i++ {
			cfg.Servers = append(cfg.Servers, ServerConfigServer{
				ID:                 i + 1,
				Host:               "127.0.0.1",
				PeerPort:           startPort + i*3 + 1,
				LeaderElectionPort: startPort + i*3 + 2,
			})
		}
		cfgPath := filepath.Join(srvPath, "zoo.cfg")
		fi, err := os.Create(cfgPath)
		if err != nil {
			return nil, err
		}
		err = cfg.Marshall(fi)
		fi.Close()
		if err != nil {
			return nil, err
		}

		fi, err = os.Create(filepath.Join(srvPath, "myid"))
		if err != nil {
			return nil, err
		}
		_, err = fmt.Fprintf(fi, "%d\n", serverN+1)
		fi.Close()
		if err != nil {
			return nil, err
		}

		srv := &Server{
			ConfigPath: cfgPath,
			Stdout:     stdout,
			Stderr:     stderr,
		}
		if err := srv.Start(); err != nil {
			return nil, err
		}
		cluster.Servers = append(cluster.Servers, TestServer{
			Path: srvPath,
			Port: cfg.ClientPort,
			Srv:  srv,
		})
	}
	if err := cluster.waitForStart(10, time.Second); err != nil {
		return nil, err
	}
	success = true
	return cluster, nil
}

func (tc *TestCluster) Stop() error {
	for _, srv := range tc.Servers {
		srv.Srv.Stop()
	}
	defer os.RemoveAll(tc.Path)
	return tc.waitForStop(5, time.Second)
}

// waitForStart blocks until the cluster is up
func (tc *TestCluster) waitForStart(maxRetry int, interval time.Duration) error {
	// verify that the servers are up with SRVR
	serverAddrs := make([]string, len(tc.Servers))
	for i, s := range tc.Servers {
		serverAddrs[i] = fmt.Sprintf("127.0.0.1:%d", s.Port)
	}

	for i := 0; i < maxRetry; i++ {
		_, ok := zk.FLWSrvr(serverAddrs, time.Second)
		if ok {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("unable to verify health of servers")
}

// waitForStop blocks until the cluster is down
func (tc *TestCluster) waitForStop(maxRetry int, interval time.Duration) error {
	// verify that the servers are up with RUOK
	serverAddrs := make([]string, len(tc.Servers))
	for i, s := range tc.Servers {
		serverAddrs[i] = fmt.Sprintf("127.0.0.1:%d", s.Port)
	}

	var success bool
	for i := 0; i < maxRetry && !success; i++ {
		success = true
		for _, ok := range zk.FLWRuok(serverAddrs, time.Second) {
			if ok {
				success = false
			}
		}
		if !success {
			time.Sleep(interval)
		}
	}
	if !success {
		return fmt.Errorf("unable to verify servers are down")
	}
	return nil
}
