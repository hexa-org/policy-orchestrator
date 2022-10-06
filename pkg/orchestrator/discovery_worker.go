package orchestrator

import (
	"log"
)

type DiscoveryWorker struct {
	Providers map[string]Provider
	Gateway   ApplicationsDataGateway
}

func (n *DiscoveryWorker) Run(work interface{}) error {
	for _, p := range n.Providers {
		log.Printf("Found discovery provider %s.", p.Name())

		for _, record := range work.([]IntegrationRecord) {
			log.Printf("Finding applications for integration provider %s.", p.Name())
			applications, _ := p.DiscoverApplications(IntegrationInfo{Name: record.Provider, Key: record.Key})

			log.Printf("Found %d applications for integration provider %s.", len(applications), p.Name())
			for _, app := range applications {
				_, err := n.Gateway.CreateIfAbsent(record.ID, app.ObjectID, app.Name, app.Description, app.Service) // idempotent work
				if err != nil {
					log.Printf("Failed to create application: %s", err)
				}
			}
		}
	}
	return nil
}

type DiscoveryWorkFinder struct {
	Results chan bool
	Gateway IntegrationsDataGateway
}

func NewDiscoveryWorkFinder(gateway IntegrationsDataGateway) DiscoveryWorkFinder {
	return DiscoveryWorkFinder{
		Results: make(chan bool),
		Gateway: gateway,
	}
}

func (finder *DiscoveryWorkFinder) MarkErroneous() {
	finder.Results <- false
}

func (finder *DiscoveryWorkFinder) MarkCompleted() {
	finder.Results <- true
}

func (finder *DiscoveryWorkFinder) Stop() {
	close(finder.Results)
}

func (finder *DiscoveryWorkFinder) FindRequested() []interface{} {
	found, err := finder.Gateway.Find()
	if err != nil {
		return nil
	}
	var results []interface{}
	if len(found) > 0 {
		results = append(results, found)
		return results
	}
	return results
}
