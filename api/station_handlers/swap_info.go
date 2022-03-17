package station_handlers

import (
	"encoding/json"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var priceExpiredSeconds = 60 * 60 * 24 * 3 // 3 days

type ReqPostSwapInfo struct {
	StafihubAddress string `json:"stafihubAddress"` //hex
	Symbol          string `json:"symbol"`
	PoolAddress     string `json:"poolAddress"`
	InAmount        string `json:"inAmount"`     //decimal
	MinOutAmount    string `json:"minOutAmount"` //decimal
}

type RspPostSwapInfo struct {
	Uuid string `json:"uuid"`
}

// @Summary post swap info
// @Description post swap info
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqPostSwapInfo true "user swap info"
// @Success 200 {object} utils.Rsp{data=RspPostSwapInfo}
// @Router /v1/station/swapInfo [post]
func (h *Handler) HandlePostSwapInfo(c *gin.Context) {
	req := ReqPostSwapInfo{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, codeParamParseErr, err.Error())
		logrus.Errorf("bind err %v", err)
		return
	}
	reqBytes, _ := json.Marshal(req)
	logrus.Infof("req parm:\n %s", string(reqBytes))

	if _, err = types.GetFromBech32(req.StafihubAddress, "stafi"); err != nil {
		utils.Err(c, codeStafiAddressErr, "stafiAddress format err")
		return
	}

	//check pool address
	metaData, err := dao_station.GetMetaData(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, codeInternalErr, "get pool address failed")
		logrus.Errorf("dao_station.GetPoolAddressBySymbol err %v", err)
		return
	}
	if !strings.EqualFold(metaData.PoolAddress, req.PoolAddress) {
		utils.Err(c, codePoolAddressErr, "pool address not right")
		logrus.Errorf("pool address not right:req %s,db:%s", req.PoolAddress, metaData.PoolAddress)
		return
	}

	//get fis price
	fisPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, utils.SymbolFis)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	//check old price
	duration := int(time.Now().Unix()) - fisPrice.UpdatedAt
	if duration > priceExpiredSeconds {
		utils.Err(c, codeTokenPriceErr, "price too old")
		return
	}

	fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	//get symbol price
	symbolPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}

	limitInfo, err := dao_station.GetLimitInfo(h.db)
	if err != nil {
		utils.Err(c, codeLimitInfoNotExistErr, err.Error())
		return
	}
	//swap rate
	swapRateStr := limitInfo.SwapRate
	swapMaxLimitStr := limitInfo.SwapMaxLimit
	swapMinLimitStr := limitInfo.SwapMinLimit

	swapRateDeci, err := decimal.NewFromString(swapRateStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapRateStr: %s err %s", swapRateStr, err)
		swapRateDeci = utils.DefaultSwapRateDeci
	}
	swapMaxLimitDeci, err := decimal.NewFromString(swapMaxLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMaxLimitStr: %s err %s", swapMaxLimitStr, err)
		swapMaxLimitDeci = utils.DefaultSwapMaxLimitDeci
	}
	swapMinLimitDeci, err := decimal.NewFromString(swapMinLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMinLimitStr: %s err %s", swapMinLimitStr, err)
		swapMinLimitDeci = utils.DefaultSwapMinLimitDeci
	}

	//cal real swap rate
	realSwapRateDeci := swapRateDeci.Mul(symbolPriceDeci).Div(fisPriceDeci)
	//in amount
	inAmountDeci, err := decimal.NewFromString(req.InAmount)
	if err != nil {
		utils.Err(c, codeInAmountFormatErr, err.Error())
		return
	}
	//out amount
	symbolDecimals := metaData.Decimals
	outAmount := realSwapRateDeci.Mul(inAmountDeci).Mul(decimal.NewFromInt(int64(symbolDecimals)))
	if outAmount.Cmp(swapMaxLimitDeci) > 0 {
		outAmount = swapMaxLimitDeci
	}
	if outAmount.Cmp(swapMinLimitDeci) < 0 {
		utils.Err(c, codeMinLimitErr, "out amount less than min limit")
		return
	}

	//check min out amount
	minOutAmountDeci, err := decimal.NewFromString(req.MinOutAmount)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,minOutAmount: %s err %s", req.MinOutAmount, err)
		utils.Err(c, codeMinOutAmountFormatErr, err.Error())
		return
	}
	if outAmount.Cmp(minOutAmountDeci) < 0 {
		utils.Err(c, codePriceSlideErr, "real out amount < min out amount")
		logrus.Errorf("real out amount: %s < min out amount: %s", outAmount.StringFixed(0), req.MinOutAmount)
		return
	}

	uuid := utils.Uuid()
	swapInfo := dao_station.FeeStationSwapInfo{}

	swapInfo.Uuid = uuid
	swapInfo.StafihubAddress = req.StafihubAddress
	swapInfo.Symbol = req.Symbol
	swapInfo.PoolAddress = req.PoolAddress
	swapInfo.InAmount = req.InAmount
	swapInfo.MinOutAmount = req.MinOutAmount
	swapInfo.InTokenPrice = symbolPrice.Price
	swapInfo.OutTokenPrice = fisPrice.Price
	swapInfo.SwapRate = realSwapRateDeci.StringFixed(0)
	swapInfo.OutAmount = outAmount.StringFixed(0)
	swapInfo.State = utils.SwapStateNotSynced

	//update db
	err = dao_station.UpOrInFeeStationSwapInfo(h.db, &swapInfo)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("UpOrInSwapInfo err %v", err)
		return
	}

	utils.Ok(c, "success", RspPostSwapInfo{
		Uuid: uuid,
	})
}

type RspGetSwapInfo struct {
	SwapStatus uint8 `json:"swapStatus"`
}

// @Summary get swap info
// @Description get swap info
// @Tags v1
// @Param uuid query string true "uuid hex string"
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspGetSwapInfo}
// @Router /v1/station/swapInfo [get]
func (h *Handler) HandleGetSwapInfo(c *gin.Context) {
	uuid := c.Query("uuid")
	//check param
	if len(uuid) == 0 {
		utils.Err(c, codeParamParseErr, "uuid empty")
		return
	}

	swapInfo, err := dao_station.GetFeeStationSwapInfoByUuid(h.db, uuid)
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.Err(c, codeInternalErr, err.Error())
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		utils.Err(c, codeSwapInfoNotExistErr, err.Error())
		return
	}

	rsp := RspGetSwapInfo{
		SwapStatus: swapInfo.State,
	}
	utils.Ok(c, "success", rsp)
}
