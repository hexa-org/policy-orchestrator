package orchestrator

import (
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"log"
	"sync/atomic"
)

type DiscoveryWorker struct {
	Providers []provider.Provider
	Gateway   ApplicationsDataGateway
}

func (n *DiscoveryWorker) Run(work interface{}) error {
	// todo - explore just in time provider creation
	for _, p := range n.Providers {
		log.Printf("Found discovery provider %s.", p.Name())

		for _, record := range work.([]IntegrationRecord) {
			log.Printf("Finding applications for integration provider %s.", p.Name())
			applications := p.DiscoverApplications(provider.IntegrationInfo{Name: record.Provider, Key: record.Key})

			log.Printf("Found %d applications for integration provider %s.", len(applications), p.Name())
			for _, app := range applications {
				_, err := n.Gateway.Create(record.ID, app.ID, app.Name, app.Description)
				if err != nil {
					log.Printf(err.Error())
				}
			}
		}
	}
	return nil
}

type DiscoveryWorkFinder struct {
	Completed    int32
	NotCompleted int32
	Gateway      IntegrationsDataGateway
}

func (finder *DiscoveryWorkFinder) MarkErroneous() {
	atomic.AddInt32(&finder.NotCompleted, 1)
}

func (finder *DiscoveryWorkFinder) MarkCompleted() {
	atomic.AddInt32(&finder.Completed, 1)
}

func (finder *DiscoveryWorkFinder) FindRequested() (results []interface{}) {
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
