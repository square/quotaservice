package mysqlpersister

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"sort"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/logging"
	qsc "github.com/square/quotaservice/protos/config"
)

const (
	pollingInterval = 1 * time.Second
)

type mysqlPersister struct {
	latestVersion int
	db            *sqlx.DB
	m             *sync.RWMutex

	watcher  chan struct{}
	shutdown chan struct{}

	activeFetchers *sync.WaitGroup

	configs map[int]*qsc.ServiceConfig
}

type configRow struct {
	Version int    `db:"Version"`
	Config  string `db:"Config"`
}

func New(dbUser, dbPass, dbHost string, dbPort int, dbName string) (config.ConfigPersister, error) {
	db, err := sqlx.Open("mysql",
		fmt.Sprintf("%s:%s@(%s:%v)/%s",
			dbUser,
			dbPass,
			dbHost,
			dbPort,
			dbName))

	if err != nil {
		return nil, err
	}

	_, err = db.Query("SELECT 1 FROM quotaservice LIMIT 1")
	if err != nil {
		return nil, errors.New("table quotaservice does not exist")
	}

	mp := &mysqlPersister{
		db:             db,
		configs:        make(map[int]*qsc.ServiceConfig),
		activeFetchers: &sync.WaitGroup{},
		m:              &sync.RWMutex{},
		watcher:        make(chan struct{}),
		shutdown:       make(chan struct{}),
		latestVersion:  -1,
	}

	mp.pullConfigs()

	mp.activeFetchers.Add(1)
	go mp.configFetcher()

	return mp, nil
}

func (mp *mysqlPersister) configFetcher() {
	defer mp.activeFetchers.Done()

	for {
		select {
		case <-time.After(pollingInterval):
			if mp.pullConfigs() {
				mp.notifyWatcher()
			}
		case <-mp.shutdown:
			logging.Print("Received shutdown signal, shutting down mysql watcher")
			return
		}
	}
}

// pullConfigs checks the database for new configs and returns true if there is a new config
func (mp *mysqlPersister) pullConfigs() bool {
	mp.m.RLock()
	v := mp.latestVersion
	mp.m.RUnlock()

	var rows []*configRow
	err := mp.db.Select(&rows, "SELECT Version, Config FROM quotaservice WHERE Version > ? ORDER BY Version ASC", v)
	if err != nil {
		logging.Printf("Received error from querying mysql for the latest configs mysql: %s", err)
		return false
	}

	// No new configs, exit
	if len(rows) == 0 {
		return false
	}

	maxVersion := -1
	for _, r := range rows {
		var c qsc.ServiceConfig
		err := proto.Unmarshal([]byte(r.Config), &c)
		if err != nil {
			logging.Printf("Could not unmarshal config version %v, error: %s", r.Version, err)
			continue
		}

		mp.m.Lock()
		mp.configs[r.Version] = &c
		mp.m.Unlock()

		maxVersion = r.Version
	}

	mp.m.Lock()
	mp.latestVersion = maxVersion
	mp.m.Unlock()

	return true
}

func (mp *mysqlPersister) notifyWatcher() {
	mp.watcher <- struct{}{}
}

// PersistAndNotify persists a marshalled configuration passed in.
func (mp *mysqlPersister) PersistAndNotify(_ string, c *qsc.ServiceConfig) error {
	b, err := proto.Marshal(c)
	_, err = mp.db.Query("INSERT INTO quotaservice (Version, Config) VALUES (?, ?)", c.GetVersion(), string(b))
	if err != nil {
		return err
	}

	return nil
}

// ConfigChangedWatcher returns a channel that is notified whenever a new config is available.
func (mp *mysqlPersister) ConfigChangedWatcher() <-chan struct{} {
	return mp.watcher
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (mp *mysqlPersister) ReadPersistedConfig() (*qsc.ServiceConfig, error) {
	mp.m.RLock()
	defer mp.m.RUnlock()
	c := mp.configs[mp.latestVersion]
	if c == nil {
		return nil, errors.New("persister has a nil config")
	}
	c = config.CloneConfig(c)

	return c, nil
}

func (mp *mysqlPersister) ReadHistoricalConfigs() ([]*qsc.ServiceConfig, error) {
	var configs []*qsc.ServiceConfig

	mp.m.RLock()
	defer mp.m.RUnlock()

	var versions []int
	for k := range mp.configs {
		versions = append(versions, k)
	}

	sort.Ints(versions)

	for _, v := range versions {
		configs = append(configs, config.CloneConfig(mp.configs[v]))
	}

	return configs, nil
}

func (mp *mysqlPersister) Close() {
	close(mp.shutdown)
	mp.activeFetchers.Wait()

	close(mp.watcher)
	err := mp.db.Close()
	if err != nil {
		logging.Printf("Could not terminate mysql connection: %v", err)
	} else {
		logging.Printf("Mysql persister shut down")
	}
}
