package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	types1 "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/sirupsen/logrus"
	hubClient "github.com/stafihub/cosmos-relay-sdk/client"
	"gorm.io/gorm"
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

	latestBlock, err := client.GetCurrentBlockHeight()
	if err != nil {
		return fmt.Errorf("client.GetCurrentBlockHeight err: %s denom: %s", err.Error(), client.GetDenom())
	}
	startBlock := int64(metaData.DealedBlock + 1)

	for willDealBlock := startBlock; willDealBlock < latestBlock-2; willDealBlock++ {

		txs, err := client.GetBlockTxsWithParseErrSkip(willDealBlock)
		if err != nil {
			return fmt.Errorf("GetBlockTxsWithParseErrSkip err: %s, block %d, denom: %s",
				err.Error(), willDealBlock, client.GetDenom())
		}

		for _, tx := range txs {
			_, err := dao_station.GetFeeStationTransInfoByTx(t.db, tx.TxHash)
			//skip if exist
			if err == nil {
				continue
			}

			for _, event := range tx.Events {
				err := t.processStringEvents(client, event, tx.Height, tx.TxHash, tx.Tx.Value, metaData)
				if err != nil {
					return fmt.Errorf("processStringEvents err: %s, block %d, denom: %s",
						err.Error(), willDealBlock, client.GetDenom())
				}
			}

		}

		metaData.DealedBlock = uint64(willDealBlock)
		err = dao_station.UpOrInMetaData(t.db, metaData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) processStringEvents(client *hubClient.Client, event types1.Event, blockNumber int64, txHash string, txValue []byte, metaData *dao_station.FeeStationMetaData) error {
	logrus.Debug("processStringEvents", "event", event)

	switch {
	case event.Type == xBankTypes.EventTypeTransfer:
		// skip if multisend
		if len(event.Attributes) != 4 {
			logrus.Debug("got multisend transfer event", "txHash", txHash, "event", event)
			return nil
		}
		// skip if not to this pool
		recipient := event.Attributes[0].Value
		if recipient != metaData.PoolAddress {
			return nil
		}

		coins, err := types.ParseCoinsNormalized(event.Attributes[2].Value)
		if err != nil {
			return fmt.Errorf("amount format err, %s", err)
		}
		amount := coins.AmountOf(metaData.Symbol)
		transInfo := &dao_station.FeeStationTransInfo{
			Uuid:            "",
			StafihubAddress: "",
			Symbol:          metaData.Symbol,
			Txhash:          txHash,
			PoolAddress:     recipient,
			InAmount:        amount.String(),
		}

		if amount == math.ZeroInt() {
			logrus.Warnf("transfer denom not equal, expect %s got %s, transinfo: %+v", metaData.Symbol, coins.String(), transInfo)
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
