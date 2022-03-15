package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"math/big"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

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
			logrus.Info("task UpdatePrice has stopped")
			break out
		case <-ticker.C:

			logrus.Debug("task UpdatePrice start -----------")
			err := task.UpdatePrice(task.db)
			if err != nil {
				logrus.Errorf("task.UpdatePrice err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Debug("task UpdatePrice end -----------")
			retry = 0
		}
	}
}

func (t *Task) UpdatePrice(db *db.WrapDb) error {

	metaDatas, err := dao_station.GetMetaDataList(t.db)
	if err != nil {
		return err
	}
	coinMarketUrl := fmt.Sprintf("%s?symbol=%s", t.coinMarketApi, utils.CoinmarketSymbolFis)
	coinGeckoUrl := fmt.Sprintf("%s?vs_currencies=usd&ids=%s", t.coinGeckoApi, utils.CoinGeckoSymbolFis)
	for _, metaData := range metaDatas {
		coinMarketUrl += "," + metaData.CoinmarketSymbol
		coinGeckoUrl += "," + metaData.CoinGeckoSymbol
	}

	retry := 0
	var resPriceMap = make(map[string]float64)
	var symbolPriceMap = make(map[string]float64)
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("cosmosRpc.NewClient reach retry limit")
		}

		resPriceMap, err = utils.GetPriceFromCoinMarket(coinMarketUrl)
		if err != nil {
			logrus.Warnf("GetPriceFromCoinMarket err: %s, will try coinGecko", err)
			resPriceMap, err = utils.GetPriceFromCoinGecko(coinGeckoUrl)
			if err != nil {
				logrus.Warnf("GetPriceFromCoinGecko err: %s, will retry", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue
			}
			for _, metaData := range metaDatas {
				if _, exist := resPriceMap[metaData.CoinGeckoSymbol]; exist {
					symbolPriceMap[metaData.Symbol] = resPriceMap[metaData.CoinGeckoSymbol]
				}
			}
			if _, exist := resPriceMap[utils.CoinGeckoSymbolFis]; exist {
				symbolPriceMap[utils.SymbolFis] = resPriceMap[utils.CoinGeckoSymbolFis]
			}

			break
		}

		for _, metaData := range metaDatas {
			if _, exist := resPriceMap[metaData.CoinmarketSymbol]; exist {
				symbolPriceMap[metaData.Symbol] = resPriceMap[metaData.CoinmarketSymbol]
			}
		}
		if _, exist := resPriceMap[utils.CoinmarketSymbolFis]; exist {
			symbolPriceMap[utils.SymbolFis] = resPriceMap[utils.CoinmarketSymbolFis]
		}

		break
	}

	for key, value := range symbolPriceMap {
		token, _ := dao_station.GetFeeStationTokenPriceBySymbol(db, key)
		token.Symbol = key
		token.Price = decimal.NewFromFloat(value).Mul(decimal.NewFromBigInt(big.NewInt(1), 6)).StringFixed(0)
		err := dao_station.UpOrInFeeStationTokenPrice(db, token)
		if err != nil {
			return err
		}
	}
	return nil
}
