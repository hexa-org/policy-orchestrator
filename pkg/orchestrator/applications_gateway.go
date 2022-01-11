package orchestrator

import (
	"database/sql"
	"log"
)

type ApplicationRecord struct {
	ID             string
	IntegreationId string
	ObjectId       string
	Name           string
	Description    string
}

type ApplicationsDataGateway struct {
	DB *sql.DB
}

func (gateway ApplicationsDataGateway) Create(integrationId string, objectId string, name string, description string) (string, error) {
	existing, _ := gateway.FindByObjectId(objectId)
	if len(existing) > 0 {
		log.Println("Found existing application record.")
		return existing[0].ID, nil
	} else {
		var id string
		err := gateway.DB.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
			integrationId, objectId, name, description).Scan(&id)
		return id, err
	}
}

func (gateway ApplicationsDataGateway) Find() ([]ApplicationRecord, error) {
	query := "select id, integration_id, object_id, name, description from applications"
	rows, err := gateway.DB.Query(query)
	return gateway.mapRecords(err, rows)
}

func (gateway ApplicationsDataGateway) FindByObjectId(objectId string) (records []ApplicationRecord, err error) {
	query := "select id, integration_id, object_id, name, description from applications where object_id=$1"
	rows, err := gateway.DB.Query(query, objectId)
	return gateway.mapRecords(err, rows)
}

///

func (gateway ApplicationsDataGateway) mapRecords(err error, rows *sql.Rows) ([]ApplicationRecord, error) {
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var records []ApplicationRecord
	for rows.Next() {
		var record ApplicationRecord
		err := rows.Scan(&record.ID, &record.IntegreationId, &record.ObjectId, &record.Name, &record.Description)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
