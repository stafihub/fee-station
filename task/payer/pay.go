package task

import (
	"fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math/big"
)

var minReserveValue = big.NewInt(1e6)

func (task Task) CheckPayInfo(db *db.WrapDb) error {
	swapInfoList, err := dao_station.GetFeeStationSwapInfoListByState(db, utils.SwapStateAlreadySynced)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}

	// ensure balance is enough
	balanceRes, err := task.client.QueryBalance(task.client.GetFromAddress(), task.client.GetDenom(), 0)
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
	for i, swapInfo := range swapInfoList {
		stafihubAddress, err := types.AccAddressFromBech32(swapInfo.StafihubAddress)
		if err != nil {
			return err
		}
		outAmountDeci, err := decimal.NewFromString(swapInfo.OutAmount)
		if err != nil {
			return err
		}
		if outAmountDeci.Cmp(task.swapMaxLimit) > 0 {
			return fmt.Errorf("outAmount > swapLimit, out: %s", outAmountDeci.StringFixed(0))
		}

		tempAmount := new(big.Int).Add(willTransferAmount, outAmountDeci.BigInt())
		if tempAmount.Cmp(maxTransferAmount) > 0 {
			break
		}

		willTransferAmount = tempAmount
		transferMaxIndex = i
		msg := xBankTypes.NewMsgSend(task.client.GetFromAddress(), stafihubAddress, types.NewCoins(types.NewCoin(task.client.GetDenom(), types.NewIntFromBigInt(outAmountDeci.BigInt()))))

		msgs = append(msgs, msg)
	}

	if len(msgs) == 0 {
		return fmt.Errorf("no msgs insufficient balance")
	}
	logrus.Infof("will pay recievers: %v \n", msgs)

	txHash, err := task.client.BroadcastBatchMsg(msgs)
	if err != nil {
		return err
	}
	logrus.Infof("pay ok, tx hash: %s", txHash)

	// update swap state in db
	tx := db.NewTransaction()
	for i, swapInfo := range swapInfoList {
		if i > transferMaxIndex {
			break
		}
		swapInfo.State = utils.SwapStatePayOk
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
		new, err := dao_station.GetFeeStationSwapInfoBySymbolTx(db, swapInfo.Symbol, swapInfo.Txhash)
		if err != nil {
			return err
		}
		if new.State != utils.SwapStatePayOk {
			return fmt.Errorf("pay state in db not update")
		}
	}
	return nil
}
