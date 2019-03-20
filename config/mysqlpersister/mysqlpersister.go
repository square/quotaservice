package mysqlpersister

import (
	"database/sql"
	"errors"
	"github.com/square/quotaservice/config/internal"
	"sort"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/logging"
	qsc "github.com/square/quotaservice/protos/config"
)

var ErrDuplicateConfig = errors.New("config with provided version number already exists")

const (
	mysqlErrDuplicateEntry = 1062
)

type MysqlPersister struct {
	latestVersion int
	db            *sql.DB
	m             *sync.RWMutex

	notifier        *internal.Notifier
	shutdown        chan struct{}
	fetcherShutdown chan struct{}

	configs map[int]*qsc.ServiceConfig
}

type configRow struct {
	Version int    `db:"Version"`
	Config  string `db:"Config"`
}

type Connector interface {
	Connect() (*sql.DB, error)
}

func New(c Connector, pollingInterval time.Duration) (*MysqlPersister, error) {
	logging.Trace("Connecting to MySQL")
	db, err := c.Connect()
	if err != nil {
		return nil, err
	}
	logging.Trace("Connecting to MySQL: OK")

	logging.Trace("Verifying table exists")
	q, args, err := sq.Select("1").From("quotaservice").Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(q, args...)
	if err != nil {
		return nil, errors.New("table quotaservice does not exist")
	}
	logging.Trace("Verifying table exists: OK")

	mp := &MysqlPersister{
		db:              db,
		configs:         make(map[int]*qsc.ServiceConfig),
		m:               &sync.RWMutex{},
		notifier:        internal.NewNotifier(),
		shutdown:        make(chan struct{}),
		fetcherShutdown: make(chan struct{}),
		latestVersion:   -1,
	}

	logging.Info("Pulling configs from MySQL")
	if _, err := mp.pullConfigs(); err != nil {
		return nil, err
	}

	mp.m.RLock()
	v := mp.latestVersion
	mp.m.RUnlock()
	logging.Infof("Pulling configs from MySQL: OK; Latest Version: %v", v)

	mp.notifyWatcher()

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
				logging.Warnf("Received an error trying to fetch config updates: %s", err)
			} else if newConf {
				logging.Debug("New config(s) found in MySQL")
				mp.notifyWatcher()
			}
		case <-mp.shutdown:
			logging.Debug("Received shutdown signal, shutting down mysql watcher")
			return
		}
	}
}

// pullConfigs checks the database for new configs and returns true if there is a new config
func (mp *MysqlPersister) pullConfigs() (bool, error) {
	mp.m.RLock()
	v := mp.latestVersion
	mp.m.RUnlock()

	logging.Tracef("Fetching configs later than %v", v)
	q, args, err := sq.
		Select("Version", "Config").
		From("quotaservice").
		Where("Version > ?", v).
		OrderBy("Version ASC").ToSql()
	if err != nil {
		return false, err
	}

	rows, err := mp.db.Query(q, args...)
	if err != nil {
		return false, err
	}
	logging.Tracef("Fetching configs later than %v: OK", v)

	rowCount := 0
	maxVersion := -1
	for rows.Next() {
		rowCount++

		var r configRow
		err := rows.Scan(&r.Version, &r.Config)
		if err != nil {
			return false, err
		}

		var c qsc.ServiceConfig
		err = proto.Unmarshal([]byte(r.Config), &c)
		if err != nil {
			logging.Warnf("Could not unmarshal config version %v, error: %s", r.Version, err)
			continue
		}

		mp.m.Lock()
		mp.configs[r.Version] = &c
		mp.m.Unlock()

		maxVersion = r.Version
	}

	if rowCount == 0 {
		logging.Debug("No versions later than %v found", v)
		return false, nil
	}

	logging.Info("Upgrading from version %v to %v", v, maxVersion)

	mp.m.Lock()
	mp.latestVersion = maxVersion
	mp.m.Unlock()

	return true, nil
}

func (mp *MysqlPersister) notifyWatcher() {
	logging.Trace("Notifying config watcher")
	mp.notifier.Notify()
}

// PersistAndNotify persists a marshalled configuration passed in.
func (mp *MysqlPersister) PersistAndNotify(_ string, c *qsc.ServiceConfig) error {
	logging.Info("Persisting version %v", c.GetVersion())
	b, err := proto.Marshal(c)
	q, args, err := sq.Insert("quotaservice").Columns("Version", "Config").Values(c.GetVersion(), string(b)).ToSql()
	if err != nil {
		return err
	}

	_, err = mp.db.Exec(q, args...)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == mysqlErrDuplicateEntry {
			return ErrDuplicateConfig
		}

		return err
	}

	logging.Infof("Persisting version %v: OK", c.GetVersion())
	return nil
}

// ConfigChangedWatcher returns a channel that is notified whenever a new config is available.
func (mp *MysqlPersister) ConfigChangedWatcher() <-chan struct{} {
	return mp.notifier.Watcher
}

// ReadPersistedConfig provides a config previously persisted.
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

// ReadHistoricalConfigs returns an array of previously persisted configs
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
	logging.Debug("Shutting down MySQL persister")
	close(mp.shutdown)
	<-mp.fetcherShutdown

	close(mp.notifier.Watcher)
	err := mp.db.Close()
	if err != nil {
		logging.Errorf("Could not terminate mysql connection: %v", err)
	} else {
		logging.Debug("Shutting down MySQL persister: OK")
	}
}
