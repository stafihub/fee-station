package dao_station

import "fee-station/pkg/db"

// metadata of other chain
type FeeStationLimitInfo struct {
	db.BaseModel
	SwapMaxLimit string `gorm:"type:varchar(30) not null;default:'0';column:swap_max_limit"` // decimals 6
	SwapMinLimit string `gorm:"type:varchar(30) not null;default:'0';column:swap_min_limit"` // decimals 6
	SwapRate     string `gorm:"type:varchar(30) not null;default:'0';column:swap_rate"`      // decimals 6
	PayerAddress string `gorm:"type:varchar(80) not null;default:'';column:payer_address"`
}

func (f FeeStationLimitInfo) TableName() string {
	return "fee_station_limit_infos"
}

func UpOrInLimitInfo(db *db.WrapDb, c *FeeStationLimitInfo) error {
	return db.Save(c).Error
}

func GetLimitInfo(db *db.WrapDb) (c *FeeStationLimitInfo, err error) {
	c = &FeeStationLimitInfo{}
	err = db.Take(c).Error
	return
}
