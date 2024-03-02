package main

import (
	"flag"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/raokrutarth/golang-playspace/common"
)

func main() {
	log := common.GetLogger()
	taskName := flag.String("task-name", "", "specify a specific task instead of running as a background worker")
	flag.Parse()
	if *taskName != "" {
		log.Info("running single task", "taskName", taskName)
		return
	}

	s := gocron.NewScheduler(time.UTC)
	_, err := s.SingletonMode().Every(time.Second * 5).Do(func() {
		log.Info("running recurring")
	})
	if err != nil {
		log.Error("failed to setup task", "error", err)
		return
	}

	log.Info("running as background worker mode")
	s.StartBlocking()
}
