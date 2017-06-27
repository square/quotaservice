// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"
	"time"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/test/helpers"
)

func TestWithNoRpcs(t *testing.T) {
	helpers.ExpectingPanic(t, func() {
		New(&MockBucketFactory{}, &config.MemoryConfigPersister{}, NewReaperConfigForTests(), 0)
	})
}

func TestValidServer(t *testing.T) {
	s := New(&MockBucketFactory{}, config.NewMemoryConfigPersister(), NewReaperConfigForTests(), 0, &MockEndpoint{}).(*server)
	_, err := s.Start()
	helpers.CheckError(t, err)
	stopServer(t, s)
}

func TestUpdateConfig(t *testing.T) {
	p := config.NewMemoryConfigPersister()
	s := New(&MockBucketFactory{}, p, NewReaperConfigForTests(), 0, &MockEndpoint{}).(*server)

	originalConfig := config.NewDefaultServiceConfig()
	originalConfig.Version = 2
	originalConfig.Date = time.Now().Unix() - 10
	marshalledConfig, err := config.Marshal(originalConfig)

	if err != nil {
		t.Fatal("Error when updating config", err)
	}

	helpers.CheckError(t, p.PersistAndNotify(marshalledConfig))

	_, err = s.Start()
	helpers.CheckError(t, err)
	defer stopServer(t, s)

	newConfig := config.NewDefaultServiceConfig()

	if err := s.UpdateConfig(newConfig, "test"); err != nil {
		t.Fatal("Error when updating config", err)
	}

	start := time.Now()

	for s.Configs().Version == originalConfig.Version {
		if time.Since(start) > time.Second {
			t.Fatal("Timeout waiting for config to change!")
		}

		time.Sleep(time.Millisecond * 5)
	}

	cfg := s.Configs()

	if cfg.User != "test" {
		t.Errorf("User %+v does not match passed in user \"test\"", s.cfgs.User)
	}

	if cfg.Version != 3 {
		t.Errorf("Version %+v does not match current version: 3", s.cfgs.Version)
	}

	if cfg.Date <= originalConfig.Date {
		t.Errorf("Date %+v was not updated from %+v", s.cfgs.Date, originalConfig.Date)
	}
}

func TestTooManyTokensRequested(t *testing.T) {
	cfg := config.NewDefaultServiceConfig()
	nsc := config.NewDefaultNamespaceConfig("dummy")
	bc := config.NewDefaultBucketConfig("dummy")
	bc.MaxTokensPerRequest = 5
	helpers.CheckError(t, config.AddBucket(nsc, bc))
	helpers.CheckError(t, config.AddNamespace(cfg, nsc))

	s := New(&MockBucketFactory{}, config.NewMemoryConfig(cfg), NewReaperConfigForTests(), 0, &MockEndpoint{}).(*server)
	_, err := s.Start()
	helpers.CheckError(t, err)
	defer stopServer(t, s)

	w, _, e := s.Allow("dummy", "dummy", 1, 0, false)
	if e != nil {
		t.Fatal("Wasn't expecting an error to s.Allow()", e)
	}

	if w > 0 {
		t.Fatalf("Wait time should be 0, not %v", w)
	}

	w, _, e = s.Allow("dummy", "dummy", 10, 0, false)
	if e == nil {
		t.Fatal("Expecting an error to s.Allow()", e)
	}

	if e.(QuotaServiceError).Reason != ER_TOO_MANY_TOKENS_REQUESTED {
		t.Fatalf("Expected Reason to be %v but was %v", ER_TOO_MANY_TOKENS_REQUESTED, e.(QuotaServiceError).Reason)
	}
}

func stopServer(t *testing.T, s *server) {
	_, err := s.Stop()
	helpers.CheckError(t, err)
}
