package station_handlers

import (
	"fee-station/dao/station"
	"fee-station/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type PoolInfo struct {
	Symbol      string `json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	PoolAddress string `json:"poolAddress"` //base58,bech32 or hex
	SwapRate    string `json:"swapRate"`    //decimals 6
}

type RspPoolInfo struct {
	PoolInfoList []PoolInfo `json:"poolInfoList"`
	SwapMaxLimit string     `json:"swapMaxLimit"` //decimals 6
	SwapMinLimit string     `json:"swapMinLimit"` //decimals 6
	PayerAddress string     `json:"payerAddress"`
}

// @Summary get pool info
// @Description get pool info
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspPoolInfo}
// @Router /v1/station/poolInfo [get]
func (h *Handler) HandleGetPoolInfo(c *gin.Context) {
	list, err := dao_station.GetMetaDataList(h.db)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		return
	}
	limitInfo, err := dao_station.GetLimitInfo(h.db)
	if err != nil {
		utils.Err(c, codeLimitInfoNotExistErr, err.Error())
		return
	}

	swapRateDeci, err := decimal.NewFromString(limitInfo.SwapRate)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,str:%s err %s", limitInfo.SwapRate, err)
		swapRateDeci = utils.DefaultSwapRateDeci
	}
	swapMaxLimitDeci, err := decimal.NewFromString(limitInfo.SwapMaxLimit)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMaxLimitStr:%s err %s", limitInfo.SwapMaxLimit, err)
		swapMaxLimitDeci = utils.DefaultSwapMaxLimitDeci
	}
	swapMinLimitDeci, err := decimal.NewFromString(limitInfo.SwapMinLimit)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMinLimitStr:%s err %s", limitInfo.SwapMinLimit, err)
		swapMinLimitDeci = utils.DefaultSwapMinLimitDeci
	}

	rsp := RspPoolInfo{
		PoolInfoList: make([]PoolInfo, 0),
		SwapMaxLimit: swapMaxLimitDeci.StringFixed(0),
		SwapMinLimit: swapMinLimitDeci.StringFixed(0),
		PayerAddress: limitInfo.PayerAddress,
	}

	//get fis price
	fisPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, utils.SymbolFis)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}

	for _, l := range list {
		//get symbol price
		symbolPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, l.Symbol)
		if err != nil {
			utils.Err(c, codeTokenPriceErr, err.Error())
			return
		}
		symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Price)
		if err != nil {
			utils.Err(c, codeTokenPriceErr, err.Error())
			return
		}
		//cal real swap rate
		realSwapRateDeci := swapRateDeci.Mul(symbolPriceDeci).Div(fisPriceDeci)

		rsp.PoolInfoList = append(rsp.PoolInfoList, PoolInfo{
			Symbol:      l.Symbol,
			PoolAddress: l.PoolAddress,
			SwapRate:    realSwapRateDeci.StringFixed(0),
			Decimals:    l.Decimals,
		})
	}

	utils.Ok(c, "success", rsp)
}
