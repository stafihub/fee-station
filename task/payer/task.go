package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
)

// Frequency of polling for a new block
const (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 100
	BlockConfirmNumber = int64(6)
)

type Task struct {
	taskTicker   int64
	client       *hubClient.Client
	payerAccount string
	swapMaxLimit string
	stop         chan struct{}
	db           *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:   cfg.TaskTicker,
		payerAccount: cfg.PayerAccount,
		swapMaxLimit: cfg.SwapMaxLimit,
		stop:         make(chan struct{}),
		db:           dao,
	}
	return s
}

func (task *Task) Start() error {
	utils.SafeGoWithRestart(task.Handler)
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}

func (task *Task) Handler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
out:
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
		}
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task CheckPayInfo start -----------")
			err := task.CheckPayInfo(task.db, task.swapMaxLimit)
			if err != nil {
				logrus.Errorf("task.CheckPayInfo err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task CheckPayInfo end -----------")
			retry = 0
		}
	}
}
