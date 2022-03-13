package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
)

// Frequency of polling for a new block
const (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 100
	BlockConfirmNumber = int64(1)
)

type Task struct {
	taskTicker    int64
	atomDenom     string
	dotTypesPath  string
	ksmTypesPath  string
	coinMarketApi string
	coinGeckoApi  string
	stop          chan struct{}
	endPoint      config.Endpoint
	db            *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:    cfg.TaskTicker,
		atomDenom:     cfg.AtomDenom,
		dotTypesPath:  cfg.DotTypesPath,
		ksmTypesPath:  cfg.KsmTypesPath,
		coinMarketApi: cfg.CoinMarketApi,
		coinGeckoApi:  cfg.CoinGeckoApi,
		stop:          make(chan struct{}),
		endPoint:      cfg.Endpoint,
		db:            dao,
	}
	return s
}

func (task *Task) Start() error {
	utils.SafeGoWithRestart(task.PriceUpdateHandler)
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}


func (task *Task) PriceUpdateHandler() {
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

			logrus.Infof("task UpdatePrice start -----------")
			err := UpdatePrice(task.db, task.coinMarketApi, task.coinGeckoApi)
			if err != nil {
				logrus.Errorf("task.UpdatePrice err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task UpdatePrice end -----------")
			retry = 0
		}
	}
}
