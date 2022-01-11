package orchestrator_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"hexa/pkg/database_support"
	"hexa/pkg/hawk_support"
	"hexa/pkg/orchestrator"
	"hexa/pkg/web_support"
	"log"
	"net/http"
	"testing"
)

func setup(key string) func(t *testing.T) {
	db, _ := database_support.Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_, _ = db.Exec("delete from applications;")
	_, _ = db.Exec("delete from integrations;")

	var integrationTestId string
	_ = db.QueryRow(`insert into integrations (name, provider, key) values ($1, $2, $3) returning id`,
		"aName", "aProvider", []byte("aKey")).Scan(&integrationTestId)
	_ = db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
		integrationTestId, "anObjectId", "aName", "aDescription").Scan(&integrationTestId)

	handlers, _ := orchestrator.LoadHandlers(hawk_support.NewCredentialStore(key), "localhost:8883", db)
	server := web_support.Create("localhost:8883", handlers, web_support.Options{})

	go web_support.Start(server)
	web_support.WaitForHealthy(server)

	return func(t *testing.T) {
		defer web_support.Stop(server)
	}
}

func TestList(t *testing.T) {
	hash := sha256.Sum256([]byte("aKey"))
	key := hex.EncodeToString(hash[:])
	teardownTestCase := setup(key)
	defer teardownTestCase(t)

	resp, err := hawk_support.HawkGet(&http.Client{}, "anId", key, "http://localhost:8883/applications")
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var apps orchestrator.Applications
	_ = json.NewDecoder(resp.Body).Decode(&apps)
	assert.Equal(t, 1, len(apps.Applications))
	assert.Equal(t, "anObjectId", apps.Applications[0].ObjectId)
	assert.Equal(t, "aName", apps.Applications[0].Name)
	assert.Equal(t, "aDescription", apps.Applications[0].Description)
}
