package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
				logrus.Infof("have dealed block height: %d, symbol: %s", willDealBlock, client.GetDenom())
			}
			willDealBlock++

			retry = 0
		}
	}
}

func (t *Task) processBlockEvents(client *hubClient.Client, currentBlock int64, metaData *dao_station.FeeStationMetaData) error {

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
			logrus.Warnf("transfer denom not equal,expect %s got %s", metaData.Symbol, coin.GetDenom())
			return nil
		}

		_, err = dao_station.GetFeeStationTransInfoByTx(t.db, txHash)
		if err == nil {
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		transInfo := &dao_station.FeeStationTransInfo{
			Uuid:            "",
			StafihubAddress: "",
			Symbol:          metaData.Symbol,
			Txhash:          txHash,
			PoolAddress:     recipient,
			InAmount:        coin.Amount.String(),
		}

		// check memo
		memoInTx, err := client.GetTxMemo(txValue)
		if err != nil {
			logrus.Warnf("memo format err: %s", err.Error())
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}
		memos := strings.Split(memoInTx, ":")
		if len(memos) != 2 {
			logrus.Warnf("memo format err, memo: %s", memoInTx)
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}
		uuid := memos[0]
		stafihubAddress := memos[1]

		_, err = types.GetFromBech32(stafihubAddress, "stafi")
		if err != nil {
			logrus.Warnf("stafi address err, memo: %s", memoInTx)
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}
		if len(uuid) == 0 {
			logrus.Warnf("uuid err, memo: %s", memoInTx)
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}

		// set stafi address and uuid
		transInfo.StafihubAddress = stafihubAddress
		transInfo.Uuid = uuid

		swapInfo, err := dao_station.GetFeeStationSwapInfoByUuid(t.db, uuid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		//uuid not exist
		if err != nil && err == gorm.ErrRecordNotFound {
			logrus.Warnf("uuid not exist in swap info, tranInfo: %+v", transInfo)
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}

		// stake != notSynced
		if swapInfo.State != utils.SwapStateNotSynced {
			logrus.Warnf("swap state not match, swap state: %d, transInfo: %+v", swapInfo.State, transInfo)
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}

		// case below will update swapinfo's state
		//amount not match
		if !strings.EqualFold(swapInfo.InAmount, transInfo.InAmount) {
			logrus.Warnf("amount not match, tranInfo: %+v, swapInfo: %+v", transInfo, swapInfo)
			swapInfo.State = utils.SwapStateInAmountNotMatch
			err := dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)
			if err != nil {
				return err
			}
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}
		//stafihub address not match
		if !strings.EqualFold(swapInfo.StafihubAddress, transInfo.StafihubAddress) {
			logrus.Warnf("stafi address not match, tranInfo: %+v, swapInfo: %+v", transInfo, swapInfo)
			swapInfo.State = utils.SwapStateStafihubAddressNotMatch
			err := dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)
			if err != nil {
				return err
			}
			return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
		}

		//update state
		swapInfo.State = utils.SwapStateAlreadySynced
		logrus.Debug("find transfer event", "block number", blockNumber)
		err = dao_station.UpOrInFeeStationSwapInfo(t.db, swapInfo)
		if err != nil {
			return err
		}
		return dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)

	default:
		return nil
	}

}
