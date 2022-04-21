package dao_station

import "fee-station/pkg/db"

// swap info, syncer will sync transinfos from other chain and change state if match the uuid
type FeeStationSwapInfo struct {
	db.BaseModel
	Uuid            string `gorm:"type:varchar(80) not null;default:'';column:uuid;uniqueIndex"` //hex string
	StafihubAddress string `gorm:"type:varchar(80) not null;default:'';column:stafihub_address"` //bech32 string
	State           uint8  `gorm:"type:tinyint(1) unsigned not null;default:0;column:state"`
	Symbol          string `gorm:"type:varchar(10) not null;default:'symbol';column:symbol"`
	PoolAddress     string `gorm:"type:varchar(80) not null;default:'';column:pool_address"`
	InAmount        string `gorm:"type:varchar(30) not null;default:'0';column:in_amount"`
	OutAmount       string `gorm:"type:varchar(30) not null;default:'0';column:out_amount"`
	MinOutAmount    string `gorm:"type:varchar(30) not null;default:'0';column:min_out_amount"`
	SwapRate        string `gorm:"type:varchar(30) not null;default:'0';column:swap_rate"` //real swaprate decimals 6
	InTokenPrice    string `gorm:"type:varchar(30) not null;default:'0';column:in_token_price"`
	OutTokenPrice   string `gorm:"type:varchar(30) not null;default:'0';column:out_token_price"`
	PayInfo         string `gorm:"type:varchar(80) not null;default:'';column:pay_info"` //pay tx hash
}

func (f FeeStationSwapInfo) TableName() string {
	return "fee_station_swap_infos"
}

func UpOrInFeeStationSwapInfo(db *db.WrapDb, c *FeeStationSwapInfo) error {
	return db.Save(c).Error
}

func GetFeeStationSwapInfoByUuid(db *db.WrapDb, uuid string) (info *FeeStationSwapInfo, err error) {
	info = &FeeStationSwapInfo{}
	err = db.Take(info, "uuid = ?", uuid).Error
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
