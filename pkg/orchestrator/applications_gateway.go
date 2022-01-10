package orchestrator

import "database/sql"

type ApplicationsDataGateway struct {
	DB *sql.DB
}