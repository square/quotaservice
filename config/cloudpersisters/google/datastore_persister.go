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
// project, and creating a new key.
//
// The Datastore should further be configured with a namespace and an entity type, both of which should be passed in
// to the `New()` function when creating a new DatastoreConfigPersister. Do not create any entity instances by hand;
// they should only ever be created by this package.
//
// Once configurations are persisted, they can be viewed on the Google Cloud admin dashboard.
package google

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/config/internal"
	"github.com/square/quotaservice/logging"
	pb "github.com/square/quotaservice/protos/config"
)

// storedEntity stores the configuration as a serialized protobuf, and some metadata about the configuration.
// storedEntities use a named key, of the format "version:{version_int}", to make it efficient to retrieve a
// specific record based on version.
type storedEntity struct {
	Contents []byte
	Version  int32
	Date     time.Time
	User     string
	Hash     string
}

// DatastoreConfigPersister is a config persister that makes use of Google Cloud's Datastore service.
type DatastoreConfigPersister struct {
	projectId string
	namespace string
	entity    string
	client    *datastore.Client
	*internal.Notifier
	version     int
	newVersions chan int
}

func (p *DatastoreConfigPersister) PersistAndNotify(oldHash string, cfg *pb.ServiceConfig) error {
	// TODO(manik) Optimistic version check with oldHash

	b, e := proto.Marshal(cfg)
	if e != nil {
		return e
	}
	// Persist...
	s := &storedEntity{Contents: b,
		Version: cfg.Version,
		Date:    time.Unix(cfg.Date, 0),
		User:    cfg.User,
		Hash:    config.HashConfigBytes(b)}

	// TODO(manik) datastore key should be the hash, not version?
	k := datastore.NameKey(p.entity, fmt.Sprintf("version:%v", cfg.Version), nil)
	k.Namespace = p.namespace

	_, e = p.client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		existing := &storedEntity{}
		e := tx.Get(k, existing)
		if e != nil && e != datastore.ErrNoSuchEntity {
			return e
		}

		if e == nil {
			if existing.Hash != s.Hash {
				// Hashes don't match, likely to be a bug.
				return fmt.Errorf("Attempting to write configuration with version %v and hash %v. Datastore already contains a configuration with the same version, with hash %v.", cfg.Version, s.Hash, existing.Hash)
			}

			// This version already exists. Do not overwrite.
			logging.Debugf("Version %v already exists; not clobbering.", cfg.Version)
			return nil
		}

		_, e = tx.Put(k, s)
		if e != nil {
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
	_, s, e := p.getLatest(false)
	if e != nil {
		return nil, e
	}

	p.newVersions <- int(s.Version)

	return config.UnmarshalBytes(s.Contents)
}

func (p *DatastoreConfigPersister) getLatest(keyOnly bool) (*datastore.Key, *storedEntity, error) {
	var entities []*storedEntity
	q := datastore.NewQuery(p.entity).
		Namespace(p.namespace).
		Order("-Version").
		Limit(1)

	if keyOnly {
		q = q.KeysOnly()
	}

	keys, e := p.client.GetAll(context.Background(), q, &entities)

	if e != nil {
		return nil, nil, e
	}

	if len(keys) != 1 {
		return nil, nil, fmt.Errorf("expected 1 result, got %v result(s)", len(keys))
	}

	return keys[0], entities[0], nil
}

func (p *DatastoreConfigPersister) ReadHistoricalConfigs() ([]*pb.ServiceConfig, error) {
	var entities []*storedEntity
	var e error

	if _, e = p.client.GetAll(context.Background(),
		datastore.NewQuery(p.entity).
			Namespace(p.namespace).
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
			k, _, e := p.getLatest(true)
			if e != nil {
				logging.Warnf("Caught error %v when polling Google Datastore", e)
			} else {
				logging.Infof("Latest version is %v", versionOf(k))
				if versionOf(k) > p.version {
					p.Notify()
				}
			}
		case v := <-p.newVersions:
			p.version = v
		}
	}
}

func versionOf(k *datastore.Key) int {
	parts := strings.Split(k.Name, ":")
	i, _ := strconv.ParseInt(parts[1], 10, 64)
	return int(i)
}

func New(projectId, credentialsFile, namespace, entity string, pollingDuration time.Duration) (*DatastoreConfigPersister, error) {
	ctx := context.Background()
	o := option.WithServiceAccountFile(credentialsFile)
	client, err := datastore.NewClient(ctx, projectId, o)

	if err != nil {
		return nil, err
	}

	p := &DatastoreConfigPersister{
		projectId:   projectId,
		namespace:   namespace,
		entity:      entity,
		client:      client,
		Notifier:    internal.NewNotifier(),
		newVersions: make(chan int)}

	go p.poll(pollingDuration)

	return p, nil
}
