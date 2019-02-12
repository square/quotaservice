package mysqlpersister

import (
	"errors"
	"sort"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
	"github.com/jmoiron/sqlx"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/logging"
	qsc "github.com/square/quotaservice/protos/config"
)

type MysqlPersister struct {
	latestVersion int
	db            *sqlx.DB
	m             *sync.RWMutex

	watcher         chan struct{}
	shutdown        chan struct{}
	fetcherShutdown chan struct{}

	configs map[int]*qsc.ServiceConfig
}

type configRow struct {
	Version int    `db:"Version"`
	Config  string `db:"Config"`
}

type Connector interface {
	Connect() (*sqlx.DB, error)
}

func New(c Connector, pollingInterval time.Duration) (*MysqlPersister, error) {
	db, err := c.Connect()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("SELECT 1 FROM quotaservice LIMIT 1")
	if err != nil {
		return nil, errors.New("table quotaservice does not exist")
	}

	mp := &MysqlPersister{
		db:              db,
		configs:         make(map[int]*qsc.ServiceConfig),
		m:               &sync.RWMutex{},
		watcher:         make(chan struct{}),
		shutdown:        make(chan struct{}),
		fetcherShutdown: make(chan struct{}),
		latestVersion:   -1,
	}

	if _, err := mp.pullConfigs(); err != nil {
		return nil, err
	}

	go mp.configFetcher(pollingInterval)

	return mp, nil
}

func (mp *MysqlPersister) configFetcher(pollingInterval time.Duration) {
	defer func() {
		close(mp.fetcherShutdown)
	}()

	for {
		select {
		case <-time.After(pollingInterval):
			if newConf, err := mp.pullConfigs(); err != nil {
				logging.Printf("Received an error trying to fetch config updates: %s", err)
			} else if newConf {
				mp.notifyWatcher()
			}
		case <-mp.shutdown:
			logging.Print("Received shutdown signal, shutting down mysql watcher")
			return
		}
	}
}

// pullConfigs checks the database for new configs and returns true if there is a new config
func (mp *MysqlPersister) pullConfigs() (bool, error) {
	mp.m.RLock()
	v := mp.latestVersion
	mp.m.RUnlock()

	q, args, err := sq.
		Select("Version", "Config").
		From("quotaservice").
		Where("Version > ?", v).
		OrderBy("Version ASC").ToSql()
	if err != nil {
		return false, err
	}

	var rows []*configRow
	err = mp.db.Select(&rows, q, args...)
	if err != nil {
		return false, err
	}

	// No new configs, exit
	if len(rows) == 0 {
		return false, nil
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

	return true, nil
}

func (mp *MysqlPersister) notifyWatcher() {
	mp.watcher <- struct{}{}
}

// PersistAndNotify persists a marshalled configuration passed in.
func (mp *MysqlPersister) PersistAndNotify(_ string, c *qsc.ServiceConfig) error {
	b, err := proto.Marshal(c)
	q, args, err := sq.Insert("quotaservice").Columns("Version", "Config").Values(c.GetVersion(), string(b)).ToSql()
	if err != nil {
		return err
	}

	_, err = mp.db.Exec(q, args...)
	if err != nil {
		return err
	}

	return nil
}

// ConfigChangedWatcher returns a channel that is notified whenever a new config is available.
func (mp *MysqlPersister) ConfigChangedWatcher() <-chan struct{} {
	return mp.watcher
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (mp *MysqlPersister) ReadPersistedConfig() (*qsc.ServiceConfig, error) {
	mp.m.RLock()
	defer mp.m.RUnlock()
	c := mp.configs[mp.latestVersion]
	if c == nil {
		return nil, errors.New("persister has a nil config")
	}
	c = config.CloneConfig(c)

	return c, nil
}

func (mp *MysqlPersister) ReadHistoricalConfigs() ([]*qsc.ServiceConfig, error) {
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

func (mp *MysqlPersister) Close() {
	close(mp.shutdown)
	<-mp.fetcherShutdown

	close(mp.watcher)
	err := mp.db.Close()
	if err != nil {
		logging.Printf("Could not terminate mysql connection: %v", err)
	} else {
		logging.Printf("Mysql persister shut down")
	}
}
