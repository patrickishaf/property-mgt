package cpaas

import (
	"log/slog"
	"sync"

	"github.com/patrickishaf/property-mgt/common"
)

type worker struct {
	client      *client
	cronManager *common.Scheduler
	mu          sync.Mutex
}

func NewWorker(client *client, cronManager *common.Scheduler) *worker {
	w := &worker{
		client:      client,
		cronManager: cronManager,
	}
	if _, err := w.cronManager.Cron().AddFunc("@every 5s", w.runJob); err != nil {
		slog.Error("failed to schedule cpaas worker", "err", err.Error())
	}
	return w
}

func (this *worker) runJob() {
	/*
		fetch pending email jobs
		attempt to send email
		update job status
	*/
}
