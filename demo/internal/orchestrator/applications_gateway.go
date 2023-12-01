package orchestrator

import (
	"database/sql"
	"log"
)

type ApplicationRecord struct {
	ID            string
	IntegrationId string
	ObjectId      string
	Name          string
	Description   string
	Service       string
}

type ApplicationsDataGateway struct {
	DB *sql.DB
}

func (gateway ApplicationsDataGateway) CreateIfAbsent(integrationId string, objectId string, name string, description string, service string) (string, error) {
	existing, _ := gateway.FindByObjectId(objectId)
	if existing != nil {
		log.Println("Found existing application record.")
		return existing.ID, nil
	}
	var id string
	err := gateway.DB.QueryRow(`insert into applications (integration_id, object_id, name, description, service) values ($1, $2, $3, $4, $5) returning id`,
		integrationId, objectId, name, description, service).Scan(&id)
	return id, err
}

func (gateway ApplicationsDataGateway) Find() ([]ApplicationRecord, error) {
	rows, err := gateway.DB.Query("select id, integration_id, object_id, name, description, service from applications order by name")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	records := make([]ApplicationRecord, 0)
	for rows.Next() {
		var record ApplicationRecord
		if erroneousScan := rows.Scan(&record.ID, &record.IntegrationId, &record.ObjectId, &record.Name, &record.Description, &record.Service); erroneousScan != nil {
			return nil, erroneousScan
		}
		records = append(records, record)
	}
	return records, nil
}

func (gateway ApplicationsDataGateway) FindByIntegrationId(integrationId string) (*ApplicationRecord, error) {
	s := "select id, integration_id, object_id, name, description, service from applications where integration_id=$1"
	return gateway.queryRow(s, integrationId)
}

func (gateway ApplicationsDataGateway) FindByObjectId(objectId string) (*ApplicationRecord, error) {
	s := "select id, integration_id, object_id, name, description, service from applications where object_id=$1"
	return gateway.query(objectId, s)
}

func (gateway ApplicationsDataGateway) FindById(id string) (*ApplicationRecord, error) {
	s := "select id, integration_id, object_id, name, description, service from applications where id=$1"
	return gateway.queryRow(s, id)
}

func (gateway ApplicationsDataGateway) DeleteById(id string) error {
	s := "delete from applications where id=$1"
	_, err := gateway.DB.Exec(s, id)
	return err
}

func (gateway ApplicationsDataGateway) queryRow(sql string, id string) (*ApplicationRecord, error) {
	row := gateway.DB.QueryRow(sql, id)
	var record ApplicationRecord
	err := row.Scan(&record.ID, &record.IntegrationId, &record.ObjectId, &record.Name, &record.Description, &record.Service)
	return &record, err
}

func (gateway ApplicationsDataGateway) query(objectId string, s string) (*ApplicationRecord, error) {
	rows, queryErr := gateway.DB.Query(s, objectId)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()

	var record ApplicationRecord
	if rows.Next() {
		err := rows.Scan(&record.ID, &record.IntegrationId, &record.ObjectId, &record.Name, &record.Description, &record.Service)
		return &record, err
	}
	return nil, nil
}
