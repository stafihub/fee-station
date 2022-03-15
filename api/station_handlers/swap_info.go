package station_handlers

import (
	"encoding/hex"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var priceExpiredSeconds = 60 * 60 * 24 * 3 // 3 days

type RspSwapInfo struct {
	SwapStatus uint8 `json:"swapStatus"`
}

// @Summary get swap info
// @Description get swap info
// @Tags v1
// @Param symbol query string true "token symbol"
// @Param txHash query string true "tx hash hex string"
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspSwapInfo}
// @Router /v1/station/swapInfo [get]
func (h *Handler) HandleGetSwapInfo(c *gin.Context) {
	symbol := c.Query("symbol")
	txHash := c.Query("txHash")
	//check param
	if len(symbol) == 0 {
		utils.Err(c, codeSymbolErr, "symbol unsupport")
		return
	}
	if _, err := hex.DecodeString(txHash); err != nil {
		utils.Err(c, codeTxHashErr, "txHash format err")
		return
	}

	swapInfo, err := dao_station.GetFeeStationSwapInfoBySymbolTx(h.db, symbol, strings.ToLower(txHash))
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.Err(c, codeInternalErr, err.Error())
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		utils.Err(c, codeSwapInfoNotExistErr, err.Error())
		return
	}

	rsp := RspSwapInfo{
		SwapStatus: swapInfo.State,
	}
	utils.Ok(c, "success", rsp)
}
