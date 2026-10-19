package cpaas

import (
	"sync"

	"github.com/patrickishaf/property-mgt/common"
)

type worker struct {
	client      *client
	cronManager *common.Scheduler
	mu          *sync.Mutex
}

func NewWorker(client *client, cronManager *common.Scheduler) *worker {
	var mu *sync.Mutex
	w := worker{
		client:      client,
		cronManager: cronManager,
		mu:          mu,
	}
	w.cronManager.Cron().AddFunc("@every 5s", w.runJob)
	return &w
}

func (this *worker) runJob() {
	/*
		fetch pending email jobs
		attempt to send email
		update job status
	*/
}
