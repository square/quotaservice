// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package google implements a ConfigPersister making use of Google Cloud's Datastore to store configuration.
// See https://cloud.google.com/datastore/ for more details.
//
// Configuration
// -------------
// This package expects a Google Cloud account set up, and should be initialized with a JSON file containing credentials
// with which to connect to Google Cloud. Google Cloud Datastore should be enabled on your account. A JSON credentials
// file can be created by browsing to https://console.cloud.google.com/iam-admin/serviceaccounts and selecting your
// project, and creating a new key. Service accounts should have "Cloud Datastore Owner" permissions.
//
// The Datastore should further be configured with a namespace and an entity type, both of which should be passed in
// to the `New()` function when creating a new `DatastoreConfigPersister`. Do not create any entity instances by hand;
// they should only ever be created by this package.
//
// Once configurations are persisted, they can be viewed on the Google Cloud admin dashboard.
package google

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/proto"
	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/logging"
	pb "github.com/square/quotaservice/protos/config"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

const (
	// persistedConfigurationKind is the "Kind" used to describe persistedConfiguration instances.
	// See https://console.cloud.google.com/datastore
	persistedConfigurationKind = "PersistedConfiguration"

	// latestPointerKind is the "Kind" used to describe latestPointer instances.
	// See https://console.cloud.google.com/datastore
	latestPointerKind = "LatestPointer"
)

// persistedConfiguration stores the configuration as a serialized protobuf, and some metadata about the configuration.
// persistedConfigurations use a named key, of the format "hash:{hash}", to make it efficient to retrieve a
// specific record based on hash.
type persistedConfiguration struct {
	Contents []byte
	Version  int32
	Date     time.Time
	User     string
	Hash     string
}

// latestPointer stores the hash of the latest configuration such that the most recent persistedConfiguration can be
// looked up my inspecting the latestPointer. There should only ever be a single latestPointer instance in existence.
type latestPointer struct {
	Hash string
}

// DatastoreConfigPersister is a config persister that makes use of Google Cloud's Datastore service.
type DatastoreConfigPersister struct {
	projectId        string
	dsNamespace      string
	client           *datastore.Client
	version          int32
	newVersions      chan int32
	latestPointerKey *datastore.Key
	*config.Notifier
}

func (p *DatastoreConfigPersister) PersistAndNotify(oldHash string, cfg *pb.ServiceConfig) error {
	b, e := proto.Marshal(cfg)
	if e != nil {
		return e
	}
	// Persist...
	s := &persistedConfiguration{Contents: b,
		Version: cfg.Version,
		Date:    time.Unix(cfg.Date, 0),
		User:    cfg.User,
		Hash:    config.HashConfigBytes(b)}

	k := p.keyFrom(s.Hash)

	_, e = p.client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		existing := &persistedConfiguration{}
		e := tx.Get(k, existing)
		if e != nil && e != datastore.ErrNoSuchEntity {
			return e
		}

		if e == nil { // Key exists!
			if existing.Version != s.Version {
				// Versions don't match, likely to be a bug.
				return fmt.Errorf("trying to write configuration with version %v and hash %v. Datastore already contains a configuration with the same hash, with version %v", cfg.Version, s.Hash, existing.Version)
			}

			// This version already exists. Do not overwrite.
			logging.Printf("Version %v Hash %v already exists; not clobbering.", cfg.Version, s.Hash)
			return nil
		}

		// Optimistic version check
		var currentLatestPointer *latestPointer
		e = tx.Get(p.latestPointerKey, currentLatestPointer)
		if e != nil {
			panic(e) // TODO(remove panics)
			return e
		}

		if currentLatestPointer.Hash != oldHash && currentLatestPointer.Hash == "uninitialized" {
			return fmt.Errorf("optimistic version check failure - expecting latestHash to be %v but was %v", oldHash, currentLatestPointer.Hash)
		}

		// Store config
		_, e = tx.Put(k, s)
		if e != nil {
			panic(e) // TODO(remove panics)
			return e
		}

		// Store latestPointer
		_, e = tx.Put(p.latestPointerKey, &latestPointer{Hash: s.Hash})
		if e != nil {
			panic(e) // TODO(remove panics)
			return e
		}

		return nil
	})

	if e != nil {
		return e
	}

	// ... and notify.
	p.Notify()

	return nil
}

func (p *DatastoreConfigPersister) ConfigChangedWatcher() <-chan struct{} {
	return p.Notifier.Watcher
}

func (p *DatastoreConfigPersister) ReadPersistedConfig() (*pb.ServiceConfig, error) {
	s, e := p.getLatest()
	if e != nil {
		return nil, e
	}

	p.newVersions <- s.Version

	return config.UnmarshalBytes(s.Contents)
}

func (p *DatastoreConfigPersister) getLatest() (*persistedConfiguration, error) {
	tx, err := p.client.NewTransaction(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to start read-only transaction: %v", err)
	}

	// read-only, so rollback on exiting function
	defer tx.Rollback()

	// Lookup pointer
	var lp *latestPointer
	err = tx.Get(p.latestPointerKey, lp)
	if err != nil {
		return nil, err
	}

	var pc *persistedConfiguration
	err = tx.Get(p.keyFrom(lp.Hash), pc)
	if err != nil {
		return nil, err
	}

	return pc, nil
}

func (p *DatastoreConfigPersister) ReadHistoricalConfigs() ([]*pb.ServiceConfig, error) {
	var entities []*persistedConfiguration
	var e error

	if _, e = p.client.GetAll(context.Background(),
		datastore.NewQuery(persistedConfigurationKind).
			Namespace(p.dsNamespace).
			Order("-Version"),
		&entities); e != nil {
		return nil, e
	}

	res := make([]*pb.ServiceConfig, len(entities))

	for i, t := range entities {
		res[i], e = config.UnmarshalBytes(t.Contents)
		if e != nil {
			return nil, e
		}
	}

	return res, nil
}

func (p *DatastoreConfigPersister) poll(pollingDuration time.Duration) {
	t := time.NewTicker(pollingDuration)
	for {
		select {
		case <-t.C:
			pc, e := p.getLatest()
			if e != nil {
				logging.Printf("Caught error %v when polling Google Datastore", e)
			} else {
				logging.Printf("Latest hash %v version %v", pc.Hash, pc.Version)
				if pc.Version > p.version {
					p.Notify()
				}
			}
		case v := <-p.newVersions:
			p.version = v
		}
	}
}

// checkOrCreateLatestPointer attempts to ensure integrity of the datastore by ensuring only a single latestPointer
// exists. If no latestPointer exists, and no configuration entities exist either, then a latestPointer is created with
// an empty hash. Otherwise an error is returned.
func (p *DatastoreConfigPersister) checkOrCreateLatestPointer() error {
	_, err := p.client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		var lp *latestPointer
		err := tx.Get(p.latestPointerKey, lp)
		if err != nil {
			// Pointer doesn't exist; attempt to create.
			// TODO(manik): only do this if there are no configuration entries stored as well.
			_, err = tx.Put(p.latestPointerKey, &latestPointer{Hash: "uninitialized"})
			if err != nil {
				return fmt.Errorf("problems initializing latestPointer: %v", err)
			}
		}

		return nil
	})

	return err
}

func (p *DatastoreConfigPersister) keyFrom(hash string) *datastore.Key {
	k := datastore.NameKey(persistedConfigurationKind, fmt.Sprintf("hash:%v", hash), nil)
	k.Namespace = p.dsNamespace

	return k
}

func New(projectId, credentialsFile, dsNamespace string, pollingDuration time.Duration) (*DatastoreConfigPersister, error) {
	o := option.WithServiceAccountFile(credentialsFile)
	client, err := datastore.NewClient(context.Background(), projectId, o)

	if err != nil {
		return nil, err
	}

	latestPointerKey := datastore.NameKey(latestPointerKind, "LATEST", nil)
	latestPointerKey.Namespace = dsNamespace

	p := &DatastoreConfigPersister{
		projectId:        projectId,
		dsNamespace:      dsNamespace,
		client:           client,
		latestPointerKey: latestPointerKey,
		Notifier:         config.NewNotifier(),
		newVersions:      make(chan int32)}

	err = p.checkOrCreateLatestPointer()
	if err != nil {
		return nil, err
	}

	go p.poll(pollingDuration)

	return p, nil
}
