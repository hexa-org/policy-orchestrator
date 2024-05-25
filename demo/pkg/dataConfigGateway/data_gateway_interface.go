package dataConfigGateway

type IntegrationsDataGateway interface {
	Create(alias string, providerType string, key []byte) (string, error)
	Find() []IntegrationRecord
	Delete(id string) error
	FindById(id string) (IntegrationRecord, error)
}

type IntegrationRecord struct {
	ID       string
	Name     string
	Provider string
	Key      []byte
}

type ApplicationRecord struct {
	ID            string
	IntegrationId string
	ObjectId      string
	Name          string
	Description   string
	Service       string
}

type ApplicationsDataGateway interface {
	Find(refresh bool) ([]ApplicationRecord, error)
	FindByObjectId(objectId string) (*ApplicationRecord, error)
	FindById(id string) (*ApplicationRecord, error)
	DeleteById(id string) error
}
