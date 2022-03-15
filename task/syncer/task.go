package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
	hubClient "github.com/stafihub/cosmos-relay-sdk/client"
)

// Frequency of polling for a new block
const (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 100
	BlockConfirmNumber = int64(1)
)

type Task struct {
	taskTicker    int64
	coinMarketApi string
	coinGeckoApi  string
	swapMaxLimit  decimal.Decimal
	swapMinLimit  decimal.Decimal
	swapRate      decimal.Decimal
	stop          chan struct{}
	db            *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:    cfg.TaskTicker,
		coinMarketApi: cfg.CoinMarketApi,
		coinGeckoApi:  cfg.CoinGeckoApi,
		stop:          make(chan struct{}),
		db:            dao,
	}
	return s
}

func (task *Task) Start() error {

	limitInfo, err := dao_station.GetLimitInfo(task.db)
	if err != nil {
		return err
	}
	maxLimit, err := decimal.NewFromString(limitInfo.SwapMaxLimit)
	if err != nil {
		return err
	}
	minLimit, err := decimal.NewFromString(limitInfo.SwapMinLimit)
	if err != nil {
		return err
	}
	swapRate, err := decimal.NewFromString(limitInfo.SwapRate)
	if err != nil {
		return err
	}
	swapRate = swapRate.Div(decimal.NewFromInt(1e6))
	task.swapMaxLimit = maxLimit
	task.swapMinLimit = minLimit
	task.swapRate = swapRate

	utils.SafeGoWithRestart(task.PriceUpdateHandler)

	metaDatas, err := dao_station.GetMetaDataList(task.db)
	if err != nil {
		return err
	}
	for _, metaData := range metaDatas {
		client, err := hubClient.NewClient(nil, "", "", metaData.Endpoint)
		if err != nil {
			return err
		}
		client.SetAccountPrefix(metaData.AccountPrefix)
		utils.SafeGoWithRestart(func() { task.pollBlocks(client) })
	}
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}
