package dao_station

import "fee-station/pkg/db"

// metadata of other chain
type FeeStationMetaData struct {
	db.BaseModel
	Symbol           string `gorm:"type:varchar(10) not null;default:'';column:symbol;uniqueIndex"` // uatom ...
	Endpoint         string `gorm:"type:varchar(50) not null;default:'';column:endpoint"`
	AccountPrefix    string `gorm:"type:varchar(10) not null;default:'';column:account_prefix"`
	PoolAddress      string `gorm:"type:varchar(80) not null;default:'';column:pool_address"`
	CoinmarketSymbol string `gorm:"type:varchar(20) not null;default:'';column:coinmarket_symbol"`
	CoinGeckoSymbol  string `gorm:"type:varchar(20) not null;default:'';column:coingecko_symbol"`
	Decimals         uint8  `gorm:"type:tinyint(1) unsigned not null;default:0;column:decimals"`
	DealedBlock      uint64 `gorm:"type:bigint(20) unsigned not null;default:0;column:dealed_block"`
}

func (f FeeStationMetaData) TableName() string {
	return "fee_station_meta_datas"
}

func UpOrInMetaData(db *db.WrapDb, c *FeeStationMetaData) error {
	return db.Save(c).Error
}

func GetMetaData(db *db.WrapDb, symbol string) (c *FeeStationMetaData, err error) {
	c = &FeeStationMetaData{}
	err = db.Take(c, "symbol = ?", symbol).Error
	return
}

func GetMetaDataList(db *db.WrapDb) (c []*FeeStationMetaData, err error) {
	err = db.Find(&c).Error
	return
}
