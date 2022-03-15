package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	hubClient "github.com/stafihub/cosmos-relay-sdk/client"
	"gorm.io/gorm"
)

func (t *Task) pollBlocksHandler(client *hubClient.Client) {
	metaData, err := dao_station.GetMetaData(t.db, client.GetDenom())
	if err != nil {
		utils.ShutdownRequestChannel <- struct{}{}
		return
	}

	var willDealBlock = metaData.SyncedBlockHeight + 1
	var retry = 0
	for {
		select {
		case <-t.stop:
			logrus.Info("task pollBlocksHandler receive stop chan, will stop")
			return
		default:
			if retry > BlockRetryLimit {
				utils.ShutdownRequestChannel <- struct{}{}
				logrus.Errorf("pollBlocks reach retry limit ")
				return
			}

			latestBlk, err := client.GetCurrentBlockHeight()
			if err != nil {
				logrus.Error("Failed to fetch latest blockNumber", "err", err)
				retry++
				time.Sleep(BlockRetryInterval)
				continue
			}
			// Sleep if the block we want comes after the most recently finalized block
			if int64(willDealBlock)+BlockConfirmNumber > latestBlk {
				time.Sleep(BlockRetryInterval)
				continue
			}
			err = t.processBlockEvents(client, int64(willDealBlock), metaData)
			if err != nil {
				logrus.Error("Failed to process events in block", "block", willDealBlock, "err", err)
				retry++
				time.Sleep(BlockRetryInterval)
				continue
			}

			// Write to blockstore

			metaData.SyncedBlockHeight = willDealBlock
			err = dao_station.UpOrInMetaData(t.db, metaData)
			if err != nil {
				logrus.Error("Failed to write to blockstore", "err", err)
			}
			if willDealBlock%1000 == 0 {
				logrus.Info("Have dealed atom block ", "height", willDealBlock)
			}
			willDealBlock++

			retry = 0
		}
	}
}

func (t *Task) processBlockEvents(client *hubClient.Client, currentBlock int64, metaData *dao_station.FeeStationMetaData) error {
	if currentBlock%100 == 0 {
		logrus.Debug("processEvents", "blockNum", currentBlock)
	}

	txs, err := client.GetBlockTxs(currentBlock)
	if err != nil {
		return fmt.Errorf("client.GetBlockTxs failed: %s", err)
	}
	for _, tx := range txs {
		for _, log := range tx.Logs {
			for _, event := range log.Events {
				err := t.processStringEvents(client, event, currentBlock, tx.TxHash, tx.Tx.Value, metaData)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *Task) processStringEvents(client *hubClient.Client, event types.StringEvent, blockNumber int64, txHash string, txValue []byte, metaData *dao_station.FeeStationMetaData) error {
	logrus.Debug("processStringEvents", "event", event)

	switch {
	case event.Type == xBankTypes.EventTypeTransfer:
		// not support multisend now
		if len(event.Attributes) != 3 {
			logrus.Debug("got multisend transfer event", "txHash", txHash, "event", event)
			return nil
		}
		// return if not to this pool
		recipient := event.Attributes[0].Value
		if recipient != metaData.PoolAddress {
			return nil
		}

		coin, err := types.ParseCoinNormalized(event.Attributes[2].Value)
		if err != nil {
			return fmt.Errorf("amount format err, %s", err)
		}
		if coin.GetDenom() != metaData.Symbol {
			logrus.Errorf("transfer denom not equal,expect %s got %s", metaData.Symbol, coin.GetDenom())
			return nil
		}

		_, err = dao_station.GetFeeStationSwapInfoBySymbolTx(t.db, metaData.Symbol, txHash)
		if err == nil {
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		fisPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(t.db, utils.SymbolFis)
		if err != nil {
			return err
		}
		atomPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(t.db, metaData.Symbol)
		if err != nil {
			return err
		}

		swapInfo := &dao_station.FeeStationSwapInfo{
			StafihubAddress: "",
			State:           utils.SwapStateAlreadySynced,
			Symbol:          metaData.Symbol,
			Txhash:          txHash,
			PoolAddress:     recipient,
			InAmount:        coin.Amount.String(),
			OutAmount:       "",
			InTokenPrice:    atomPrice.Price,
			OutTokenPrice:   fisPrice.Price,
			PayInfo:         "",
		}
		// check memo
		memoInTx, err := client.GetTxMemo(txValue)
		if err != nil {
			swapInfo.State = utils.SwapStateMemoFailed
			return dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)
		}
		_, err = types.GetFromBech32(memoInTx, "stafi")
		if err != nil {
			swapInfo.State = utils.SwapStateMemoFailed
			return dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)
		}
		swapInfo.StafihubAddress = memoInTx

		// cal outamount
		fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
		if err != nil {
			return err
		}
		atomPriceDeci, err := decimal.NewFromString(atomPrice.Price)
		if err != nil {
			return err
		}

		swapRateDeci := t.swapRate
		inAmountDeci := decimal.NewFromBigInt(coin.Amount.BigInt(), 0)

		outAmount := atomPriceDeci.Mul(swapRateDeci).Mul(inAmountDeci).Div(fisPriceDeci)

		if outAmount.LessThan(t.swapMinLimit) {
			swapInfo.OutAmount = outAmount.StringFixed(0)
			swapInfo.State = utils.SwapStateAmountLessThanMinLimit
			return dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)
		}

		if outAmount.GreaterThan(t.swapMaxLimit) {
			outAmount = t.swapMaxLimit
		}
		logrus.Debug("find swap event", "block number", blockNumber)
		swapInfo.OutAmount = outAmount.StringFixed(0)
		return dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)

	default:
		return nil
	}

}
