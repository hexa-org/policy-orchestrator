package workflowsupport

import (
	"log"
	"time"
)

type Worker interface {
	Run(interface{}) error
}

type WorkFinder interface {
	FindRequested() []interface{}
	MarkCompleted()
	MarkErroneous()
	Stop()
}

type WorkScheduler struct {
	Finder  interface{}
	Workers []Worker
	Delay   int64

	done chan bool
}

func NewScheduler(finder interface{}, workers []Worker, delay int64) WorkScheduler {
	return WorkScheduler{
		Finder:  finder,
		Workers: workers,
		Delay:   delay,
		done:    make(chan bool),
	}
}

func (ws *WorkScheduler) Start() {
	log.Printf("Starting the scheduler.\n")
	ticker := time.NewTicker(time.Duration(ws.Delay) * time.Millisecond)
	for _, w := range ws.Workers {
		go func(worker Worker) {
			for {
				select {
				case <-ws.done:
					return
				case <-ticker.C:
					log.Printf("Scheduling work.\n")
					ws.checkForWork(worker)
				}
			}
		}(w)
	}
}

func (ws *WorkScheduler) checkForWork(worker Worker) {
	finder := ws.Finder.(WorkFinder)
	log.Printf("Checking for work.\n")

	for _, t := range finder.FindRequested() {
		log.Printf("Found work.\n")

		go func(task interface{}) {
			if err := worker.Run(task); err != nil {
				finder.MarkErroneous()
				return
			}
			log.Printf("Completed work.\n")
			finder.MarkCompleted()
		}(t)
	}
}

func (ws *WorkScheduler) Stop() {
	ws.done <- true
	ws.Finder.(WorkFinder).Stop()
	log.Printf("Scheduler stopped.\n")
}
