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
	if existing.ObjectId != "" {
		log.Println("Found existing application record.")
		return existing.ID, nil
	} else {
		var id string
		err := gateway.DB.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
			integrationId, objectId, name, description).Scan(&id)
		return id, err
	}
}

func (gateway ApplicationsDataGateway) Find() ([]ApplicationRecord, error) {
	rows, err := gateway.DB.Query("select id, integration_id, object_id, name, description from applications")
	if err != nil {
		log.Println(err)
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
			log.Println(err)
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (gateway ApplicationsDataGateway) FindByObjectId(objectId string) (records ApplicationRecord, err error) {
	s := "select id, integration_id, object_id, name, description from applications where object_id=$1"
	return gateway.queryRow(s, objectId, err)
}

func (gateway ApplicationsDataGateway) FindById(id string) (records ApplicationRecord, err error) {
	s := "select id, integration_id, object_id, name, description from applications where id=$1"
	return gateway.queryRow(s, id, err)
}

func (gateway ApplicationsDataGateway) queryRow(sql string, id string, err error) (ApplicationRecord, error) {
	row := gateway.DB.QueryRow(sql, id)
	var record ApplicationRecord
	err = row.Scan(&record.ID, &record.IntegreationId, &record.ObjectId, &record.Name, &record.Description)
	return record, err
}
