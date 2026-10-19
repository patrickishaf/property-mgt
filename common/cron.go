package common

import (
	"context"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	c *cron.Cron
}

func CreateScheduler() *Scheduler {
	return &Scheduler{
		c: cron.New(),
	}
}

func (c *Scheduler) Cron() *cron.Cron {
	return c.c
}

func (c *Scheduler) Start() {
	c.c.Start()
}

func (c *Scheduler) Stop() context.Context {
	return c.c.Stop()
}
