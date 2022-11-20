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

var (
	pageLimit = 10
)

func (task *Task) SyncTransferTxHandler(client *hubClient.Client) {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()

	retry := 0
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
			return
		}

		select {
		case <-task.stop:
			logrus.Info("SyncTransferTxHandler has stopped")
			return
		case <-ticker.C:
			logrus.Debugf("task SyncTransferTxHandler start -----------")
			err := task.SyncTransferTx(client)
			if err != nil {
				logrus.Errorf("task.SyncTransferTx err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue
			}
			logrus.Debugf("task SyncTransferTxHandler end -----------")
			retry = 0
		}
	}
}

func (t *Task) SyncTransferTx(client *hubClient.Client) error {
	metaData, err := dao_station.GetMetaData(t.db, client.GetDenom())
	if err != nil {
		return err
	}

	poolAddress := metaData.PoolAddress

	filter := []string{fmt.Sprintf("transfer.recipient='%s'", poolAddress), "message.module='bank'"}

	for {

		totalCount, err := dao_station.GetFeeStationTransInfoTotalCount(t.db, client.GetDenom())
		if err != nil {
			return err
		}

		logrus.Debugf("%s totalCount in db %d", client.GetDenom(), totalCount)

		// should reduce the tx number on old chain if upgrade by new genesis
		if strings.EqualFold(client.GetDenom(), "uhuahua") {
			numberOnOldChain := int64(20)
			if totalCount > numberOnOldChain {
				totalCount -= numberOnOldChain
			}
		}

		logrus.Debugf("%s will use totalCount  %d", client.GetDenom(), totalCount)

		txResPre, err := client.GetTxs(filter, int(1), pageLimit, "asc")
		if err != nil {
			return err
		}
		logrus.Debugf("%s txs on chain, totalCount: %d, totalPage: %d, limit: %d", client.GetDenom(), txResPre.TotalCount, txResPre.PageTotal, txResPre.Limit)

		usePage := totalCount/int64(pageLimit) + 1

		//sip if localdb have
		if uint64(usePage) > txResPre.PageTotal {
			return nil
		}

		txRes, err := client.GetTxs(filter, int(usePage), pageLimit, "asc")
		if err != nil {
			return err
		}
		logrus.Debugf("%s get txs: %d", client.GetDenom(), len(txRes.Txs))

		for _, tx := range txRes.Txs {
			_, err := dao_station.GetFeeStationTransInfoByTx(t.db, tx.TxHash)
			//skip if exist
			if err == nil {
				continue
			}

			for _, log := range tx.Logs {
				for _, event := range log.Events {
					err := t.processStringEvents(client, event, tx.Height, tx.TxHash, tx.Tx.Value, metaData)
					if err != nil {
						return err
					}
				}
			}

		}

		//just break when get all
		if txRes.PageTotal == txRes.PageNumber {
			break
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
			return fmt.Errorf("got multisend event")
		}
		// return if not to this pool
		recipient := event.Attributes[0].Value
		if recipient != metaData.PoolAddress {
			return fmt.Errorf("recipient not match")
		}

		coin, err := types.ParseCoinNormalized(event.Attributes[2].Value)
		if err != nil {
			return fmt.Errorf("amount format err, %s", err)
		}

		transInfo := &dao_station.FeeStationTransInfo{
			Uuid:            "",
			StafihubAddress: "",
			Symbol:          metaData.Symbol,
			Txhash:          txHash,
			PoolAddress:     recipient,
			InAmount:        coin.Amount.String(),
		}

		if coin.GetDenom() != metaData.Symbol {
			logrus.Warnf("transfer denom not equal, expect %s got %s, transinfo: %+v", metaData.Symbol, coin.GetDenom(), transInfo)
			dao_station.UpOrInFeeStationTransInfo(t.db, transInfo)
			return nil
		}

		// skip if already exists
		_, err = dao_station.GetFeeStationTransInfoByTx(t.db, txHash)
		if err == nil {
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
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
