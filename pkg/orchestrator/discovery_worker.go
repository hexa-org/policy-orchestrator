package orchestrator

import (
	"hexa/pkg/orchestrator/provider"
	"log"
)

type DiscoveryWorker struct {
	Providers []provider.Provider
	Gateway   ApplicationsDataGateway
}

func (n *DiscoveryWorker) Run(work interface{}) error {
	for _, p := range n.Providers {
		log.Printf("Found discovery provider %s.", p.Name())

		for _, record := range work.([]IntegrationRecord) {
			log.Printf("Finding applications for integration provider %s.", p.Name())
			applications := p.DiscoveryApplications(provider.IntegrationInfo{Name: record.Provider, Key: record.Key})

			log.Printf("Found %d applications for integration provider %s.", len(applications), p.Name())
		}
	}
	return nil
}

type DiscoveryWorkFinder struct {
	Completed    int
	NotCompleted int
	Gateway      IntegrationsDataGateway
}

func (finder *DiscoveryWorkFinder) MarkErroneous(task interface{}) {
	finder.NotCompleted = finder.NotCompleted + 1
}

func (finder *DiscoveryWorkFinder) MarkCompleted(task interface{}) {
	finder.Completed = finder.Completed + 1
}

func (finder DiscoveryWorkFinder) FindRequested() (results []interface{}) {
	found, err := finder.Gateway.Find()
	if err != nil {
		log.Printf(err.Error())
		return results
	}
	if len(found) > 0 {
		results = append(results, found)
		return results
	}
	return results
}
