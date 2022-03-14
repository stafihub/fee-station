package dao_station

import "fee-station/pkg/db"

// metadata of other chain
type FeeStationMetaData struct {
	db.BaseModel
	Symbol            string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol;uniqueIndex"`
	SyncedBlockHeight uint64 `gorm:"type:bigint(20);unsigned;not null;default:0;column:synced_block_height"` //latest block height have dealed
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
