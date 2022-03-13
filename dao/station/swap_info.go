package dao_station

import "fee-station/pkg/db"

// swap info
type FeeStationSwapInfo struct {
	db.BaseModel
	StafihubAddress string `gorm:"type:varchar(80);not null;default:'';column:stafihub_address"` //bech32 string
	State           uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:state"`
	Symbol          string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	Txhash          string `gorm:"type:varchar(80);not null;default:'';column:tx_hash;uniqueIndex:uni_idx_tx"`
	PoolAddress     string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
	InAmount        string `gorm:"type:varchar(30);not null;default:'0';column:in_amount"`
	MinOutAmount    string `gorm:"type:varchar(30);not null;default:'0';column:min_out_amount"`
	OutAmount       string `gorm:"type:varchar(30);not null;default:'0';column:out_amount"`
	SwapRate        string `gorm:"type:varchar(30);not null;default:'0';column:swap_rate"` // decimal 18
	InTokenPrice    string `gorm:"type:varchar(30);not null;default:'0';column:in_token_price"`
	OutTokenPrice   string `gorm:"type:varchar(30);not null;default:'0';column:out_token_price"`
	PayInfo         string `gorm:"type:varchar(80);not null;default:'';column:pay_info"` //pay tx hash
}

func (f FeeStationSwapInfo) TableName() string {
	return "fee_staion_swap_infos"
}

func UpOrInFeeStationSwapInfo(db *db.WrapDb, c *FeeStationSwapInfo) error {
	return db.Save(c).Error
}

func GetFeeStationSwapInfoBySymbolTx(db *db.WrapDb, symbol, tx string) (info *FeeStationSwapInfo, err error) {
	info = &FeeStationSwapInfo{}
	err = db.Take(info, "symbol = ? and tx_hash = ?", symbol, tx).Error
	return
}

func GetFeeStationSwapInfoByTx(db *db.WrapDb, tx string) (info *FeeStationSwapInfo, err error) {
	info = &FeeStationSwapInfo{}
	err = db.Take(info, "tx_hash = ?", tx).Error
	return
}

func GetFeeStationSwapInfoListBySymbolState(db *db.WrapDb, symbol string, state uint8) (infos []*FeeStationSwapInfo, err error) {
	err = db.Find(&infos, "symbol = ? and state = ?", symbol, state).Error
	return
}

func GetFeeStationSwapInfoListByState(db *db.WrapDb, state uint8) (infos []*FeeStationSwapInfo, err error) {
	err = db.Limit(200).Find(&infos, "state = ?", state).Error
	return
}
