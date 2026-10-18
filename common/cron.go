package common

import "github.com/robfig/cron/v3"

type CronExecutor struct {
	c *cron.Cron
}

func CreateCronExecutor() *CronExecutor {
	return &CronExecutor{
		c: cron.New(),
	}
}

func (c *CronExecutor) Cron() *cron.Cron {
	return c.c
}

func (c *CronExecutor) Start() {
	c.c.Start()
}
