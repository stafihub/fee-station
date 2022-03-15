package dao_station

import "fee-station/pkg/db"

// metadata of other chain
type FeeStationLimitInfo struct {
	db.BaseModel
	SwapMaxLimit string `gorm:"type:varchar(30);not null;default:'0';column:swap_max_limit"`
	SwapMinLimit string `gorm:"type:varchar(30);not null;default:'0';column:swap_min_limit"`
	SwapRate     string `gorm:"type:varchar(30);not null;default:'0';column:swap_rate"`
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