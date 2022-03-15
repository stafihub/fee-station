package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
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
	swapMaxLimit decimal.Decimal
	stop         chan struct{}
	db           *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb, client *hubClient.Client) *Task {
	s := &Task{
		taskTicker:   cfg.TaskTicker,
		payerAccount: cfg.PayerAccount,
		client:       client,
		stop:         make(chan struct{}),
		db:           dao,
	}
	return s
}

func (task *Task) Start() error {
	limitInfo, err := dao_station.GetLimitInfo(task.db)
	if err != nil {
		return err
	}
	maxLimitDeci, err := decimal.NewFromString(limitInfo.SwapMaxLimit)
	if err != nil {
		return err
	}
	task.swapMaxLimit = maxLimitDeci
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
			err := task.CheckPayInfo(task.db)
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
