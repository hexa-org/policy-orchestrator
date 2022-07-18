package orchestrator

import (
	"database/sql"
)

type IntegrationRecord struct {
	ID       string
	Name     string
	Provider string
	Key      []byte
}

type IntegrationsDataGateway struct {
	DB *sql.DB
}

func (gateway IntegrationsDataGateway) Create(name string, provider string, key []byte) (string, error) {
	var id string
	err := gateway.DB.QueryRow("insert into integrations (name, provider, key) values ($1, $2, $3) returning id",
		name, provider, key).Scan(&id)
	return id, err
}

func (gateway IntegrationsDataGateway) Find() ([]IntegrationRecord, error) {
	rows, err := gateway.DB.Query("select id, name, provider, key from integrations")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	records := make([]IntegrationRecord, 0)
	for rows.Next() {
		var record IntegrationRecord
		if erroneousScan := rows.Scan(&record.ID, &record.Name, &record.Provider, &record.Key); erroneousScan != nil {
			return nil, erroneousScan
		}
		records = append(records, record)
	}
	return records, nil
}

func (gateway IntegrationsDataGateway) Delete(id string) error {
	_, err := gateway.DB.Exec("delete from integrations where id=$1", id)
	return err
}

func (gateway IntegrationsDataGateway) FindById(id string) (IntegrationRecord, error) {
	row := gateway.DB.QueryRow("select id, name, provider, key from integrations where id=$1", id)
	var record IntegrationRecord
	err := row.Scan(&record.ID, &record.Name, &record.Provider, &record.Key)
	return record, err
}
