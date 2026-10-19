package cpaas

import (
	"fmt"
	"sync"

	"github.com/patrickishaf/property-mgt/common"
)

type worker struct {
	client      *client
	cronManager *common.CronExecutor
	mu          *sync.Mutex
}

func NewWorker(client *client, cronManager *common.CronExecutor) *worker {
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
	fmt.Println("running cpaas worker")
	/*
		fetch pending email jobs
		attempt to send email
		update job status
	*/
}
