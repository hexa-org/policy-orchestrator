package orchestrator

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/orchestrator/provider"
	"log"
)

type DiscoveryWorker struct {
	Providers map[string]provider.Provider
	Gateway   ApplicationsDataGateway
}

func (n *DiscoveryWorker) Run(work interface{}) error {
	// todo - explore just in time provider creation
	for _, p := range n.Providers {
		log.Printf("Found discovery provider %s.", p.Name())

		for _, record := range work.([]IntegrationRecord) {
			log.Printf("Finding applications for integration provider %s.", p.Name())
			applications, _ := p.DiscoverApplications(provider.IntegrationInfo{Name: record.Provider, Key: record.Key})

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

func Report(results chan bool) {
	go func() {
		for success := range results {
			if success {
				fmt.Println("work successfully completed")
			} else {
				fmt.Println("work not completed")
			}
		}
	}()
}
