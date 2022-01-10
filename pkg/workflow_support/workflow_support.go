package workflow_support

import (
	"log"
	"time"
)

type Worker interface {
	Run(interface{}) error
}

type WorkFinder interface {
	FindRequested() []interface{}
	MarkCompleted(interface{})
	MarkErroneous(interface{})
}

type WorkScheduler struct {
	Finder  interface{}
	Workers []Worker
	Delay   int64

	ticker time.Ticker
	done   chan bool
}

func (w *WorkScheduler) Start() {
	log.Printf("Starting the scheduler.\n")
	w.ticker = *time.NewTicker(time.Duration(w.Delay) * time.Millisecond)
	w.done = make(chan bool)
	for _, worker := range w.Workers {
		worker := worker
		go func() {
			for {
				select {
				case <-w.done:
					return
				case _ = <-w.ticker.C:
					log.Printf("Scheduling work.\n")
					w.checkForWork(worker)
				}
			}
		}()
	}
}

func (w *WorkScheduler) checkForWork(worker Worker) {
	finder := w.Finder.(WorkFinder)
	log.Printf("Checking for work.\n")

	for _, task := range finder.FindRequested() {
		log.Printf("Found work.\n")

		task := task
		go func() {
			err := worker.Run(task)
			if err != nil {
				log.Printf("oops. %v\n", err)
				finder.MarkErroneous(task)
				return
			}
			log.Printf("Completed work.\n")
			finder.MarkCompleted(task)
		}()
	}
}

func (w *WorkScheduler) Stop() {
	w.ticker.Stop()
	w.done <- true
	log.Printf("Scheduler stopped.\n")
}
