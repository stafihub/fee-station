package task

import (
	"fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stafihub/rtoken-relay-core/common/core"
)

var minReserveValue = big.NewInt(1e6)

func (task *Task) CheckPayInfoHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
			logrus.Errorf("CheckPayInfo reach retry limit")
			return
		}
		select {
		case <-task.stop:
			logrus.Info("task CheckPayInfoHandler receive stop chan, will stop")
			return
		case <-ticker.C:
			logrus.Debugf("task CheckPayInfo start -----------")
			err := task.CheckPayInfo(task.db)
			if err != nil {
				logrus.Errorf("task.CheckPayInfo err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue
			}
			logrus.Debugf("task CheckPayInfo end -----------")
			retry = 0
		}
	}
}

func (task Task) CheckPayInfo(db *db.WrapDb) error {
	swapInfoList, err := dao_station.GetFeeStationSwapInfoListByState(db, utils.SwapStateAlreadySynced)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}

	// ensure balance is enough
	balanceRes, err := task.stafihubClient.QueryBalance(task.stafihubClient.GetFromAddress(), task.stafihubClient.GetDenom(), 0)
	if err != nil {
		return err
	}
	if balanceRes.GetBalance().Amount.BigInt().Cmp(minReserveValue) < 0 {
		return fmt.Errorf("insufficient balance")
	}
	// the max amount we can transfer this time
	maxTransferAmount := new(big.Int).Sub(balanceRes.GetBalance().Amount.BigInt(), minReserveValue)

	// merge transfer msg
	willTransferAmount := big.NewInt(0)
	msgs := make([]types.Msg, 0)
	transferMaxIndex := -1

	done := core.UseSdkConfigContext("stafi")
	for i, swapInfo := range swapInfoList {
		stafihubAddress, err := types.AccAddressFromBech32(swapInfo.StafihubAddress)
		if err != nil {
			done()
			return err
		}
		outAmountDeci, err := decimal.NewFromString(swapInfo.OutAmount)
		if err != nil {
			done()
			return err
		}
		if outAmountDeci.Cmp(task.swapMaxLimit) > 0 {
			done()
			return fmt.Errorf("outAmount > swapLimit, out: %s", outAmountDeci.StringFixed(0))
		}

		tempAmount := new(big.Int).Add(willTransferAmount, outAmountDeci.BigInt())
		if tempAmount.Cmp(maxTransferAmount) > 0 {
			break
		}

		willTransferAmount = tempAmount
		transferMaxIndex = i
		msg := xBankTypes.NewMsgSend(task.stafihubClient.GetFromAddress(), stafihubAddress, types.NewCoins(types.NewCoin(task.stafihubClient.GetDenom(), types.NewIntFromBigInt(outAmountDeci.BigInt()))))
		msgs = append(msgs, msg)
	}
	done()

	if len(msgs) == 0 {
		return fmt.Errorf("no msgs insufficient balance")
	}
	logrus.Infof("will pay recievers: %v \n", msgs)

	retry := 0
	var txHash string
	for {
		if retry >= BlockRetryLimit {
			return fmt.Errorf("BroadcastBatchMsg reach retry limit: %s", err)
		}
		txHash, err = task.stafihubClient.BroadcastBatchMsg(msgs)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "incorrect account sequence") {
				logrus.Warn("BroadcastBatchMsg err will retry: %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue
			} else {
				return err
			}
		}
		break
	}

	retry = 0
	var txRes *types.TxResponse
	for {
		if retry >= BlockRetryLimit {
			return fmt.Errorf("QueryTxByHash reach retry limit: %s", err)
		}
		txRes, err = task.stafihubClient.QueryTxByHash(txHash)
		if err != nil || txRes.Empty() || txRes.Height == 0 {
			logrus.Warn("QueryTxByHash tx failed will retry query", err, txRes)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}
	if txRes.Code != 0 {
		return fmt.Errorf("tx err: %s", txRes.String())
	}

	logrus.Infof("pay ok, tx hash: %s", txHash)

	// update swap state in db
	tx := db.NewTransaction()
	for i, swapInfo := range swapInfoList {
		if i > transferMaxIndex {
			break
		}
		swapInfo.State = utils.SwapStatePayOk
		swapInfo.PayInfo = txHash
		err := dao_station.UpOrInFeeStationSwapInfo(tx, swapInfo)
		if err != nil {
			tx.RollbackTransaction()
			return err
		}
	}
	err = tx.CommitTransaction()
	if err != nil {
		return fmt.Errorf("tx.CommitTransaction err: %s", err)
	}

	// ensure state updated
	for i, swapInfo := range swapInfoList {
		if i > transferMaxIndex {
			break
		}
		new, err := dao_station.GetFeeStationSwapInfoByUuid(db, swapInfo.Uuid)
		if err != nil {
			return err
		}
		if new.State != utils.SwapStatePayOk {
			return fmt.Errorf("pay state in db not update")
		}
	}
	return nil
}
