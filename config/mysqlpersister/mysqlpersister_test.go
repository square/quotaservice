package mysqlpersister

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"

	qsc "github.com/square/quotaservice/protos/config"
)

var db *sqlx.DB
var port int64

const (
	databaseCreateStatement = "CREATE DATABASE quotaservice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	tableCreateStatement    = "CREATE TABLE quotaservice.quotaservice (ID BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, Version INT UNIQUE, Config BLOB);"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "5.6", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sqlx.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	_, err = db.Query(databaseCreateStatement)
	if err != nil {
		panic(err)
	}

	_, err = db.Query(tableCreateStatement)
	if err != nil {
		panic(err)
	}

	port, err = strconv.ParseInt(resource.GetPort("3306/tcp"), 10, 32)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func setup(require *r.Assertions, db *sqlx.DB) {
	_, err := db.Query("TRUNCATE TABLE quotaservice.quotaservice;")
	require.NoError(err)
}

func TestReadPersistedConfig(t *testing.T) {
	assert := a.New(t)
	require := r.New(t)

	setup(require, db)
	p, err := New("root", "secret", "localhost", int(port), "quotaservice")
	assert.NoError(err)

	c1234 := &qsc.ServiceConfig{
		Version: 1234,
	}

	assert.NoError(p.PersistAndNotify("", c1234))

	cPersisted, err := p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Nil(cPersisted)

	select {
	case <-time.After(2 * pollingInterval):
		assert.Fail("No notification received for new config")
	case <-p.ConfigChangedWatcher():
	}

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(c1234, cPersisted)

	c1233 := &qsc.ServiceConfig{
		Version: 1233,
	}

	assert.NoError(p.PersistAndNotify("", c1233))

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(c1234, cPersisted)

	select {
	case <-time.After(2 * pollingInterval):
		// Do nothing
	case <-p.ConfigChangedWatcher():
		assert.Fail("Watcher was notified when an old config was persisted")
	}

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(c1234, cPersisted)

	assert.Error(p.PersistAndNotify("", c1234))

	p.(*mysqlPersister).Close()
}

func TestFirstConfigVersion(t *testing.T) {
	assert := a.New(t)
	require := r.New(t)

	setup(require, db)
	p, err := New("root", "secret", "localhost", int(port), "quotaservice")
	assert.NoError(err)

	u := "I'm a test!"
	cTest := &qsc.ServiceConfig{
		Version: 0,
		User:    u,
	}

	assert.NoError(p.PersistAndNotify("", cTest))

	cPersisted, err := p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Nil(cPersisted)

	select {
	case <-time.After(2 * pollingInterval):
		assert.Fail("No notification received for new config")
	case <-p.ConfigChangedWatcher():
	}

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(cTest, cPersisted)

	p.(*mysqlPersister).Close()
}

func TestReadHistoricalConfig(t *testing.T) {
	assert := a.New(t)
	require := r.New(t)

	setup(require, db)
	p, err := New("root", "secret", "localhost", int(port), "quotaservice")
	assert.NoError(err)

	c1233 := &qsc.ServiceConfig{
		Version: 1233,
	}

	c1234 := &qsc.ServiceConfig{
		Version: 1234,
	}

	c1235 := &qsc.ServiceConfig{
		Version: 1235,
	}

	assert.NoError(p.PersistAndNotify("", c1233))

	cPersisted, err := p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Nil(cPersisted)

	select {
	case <-time.After(2 * pollingInterval):
		assert.Fail("No notification received for new config")
	case <-p.ConfigChangedWatcher():
	}

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(c1233, cPersisted)

	cHistorical, err := p.ReadHistoricalConfigs()
	assert.NoError(err)
	assert.Equal([]*qsc.ServiceConfig{c1233}, cHistorical)

	assert.NoError(p.PersistAndNotify("", c1234))
	assert.NoError(p.PersistAndNotify("", c1235))

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(c1233, cPersisted)

	select {
	case <-time.After(2 * pollingInterval):
		assert.Fail("No notification received for new config")
	case <-p.ConfigChangedWatcher():
	}

	cPersisted, err = p.ReadPersistedConfig()
	assert.NoError(err)
	assert.Equal(c1235, cPersisted)

	cHistorical, err = p.ReadHistoricalConfigs()
	assert.NoError(err)
	assert.Equal([]*qsc.ServiceConfig{c1233, c1234, c1235}, cHistorical)

	p.(*mysqlPersister).Close()
}
