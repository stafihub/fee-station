package dao_station

import "fee-station/pkg/db"

// trans info, syncer will sync transinfos from other chain
type FeeStationTransInfo struct {
	db.BaseModel
	Uuid            string `gorm:"type:varchar(80) not null;default:'';column:uuid"`             //hex string
	StafihubAddress string `gorm:"type:varchar(80) not null;default:'';column:stafihub_address"` //bech32 string
	Symbol          string `gorm:"type:varchar(10) not null;default:'symbol';column:symbol"`
	Txhash          string `gorm:"type:varchar(80) not null;default:'';column:tx_hash;uniqueIndex"`
	PoolAddress     string `gorm:"type:varchar(80) not null;default:'';column:pool_address"`
	InAmount        string `gorm:"type:varchar(30) not null;default:'0';column:in_amount"`
}

func (f FeeStationTransInfo) TableName() string {
	return "fee_station_trans_infos"
}

func UpOrInFeeStationTransInfo(db *db.WrapDb, c *FeeStationTransInfo) error {
	return db.Save(c).Error
}

func GetFeeStationTransInfoByTx(db *db.WrapDb, tx string) (info *FeeStationTransInfo, err error) {
	info = &FeeStationTransInfo{}
	err = db.Take(info, "tx_hash = ?", tx).Error
	return
}

func GetFeeStationTransInfoTotalCount(db *db.WrapDb, symbol string) (count int64, err error) {
	err = db.Model(&FeeStationTransInfo{}).Where("symbol = ?", symbol).Count(&count).Error
	return
}
