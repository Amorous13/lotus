package dao

import (
	"context"
	"errors"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type SectorExpirationInfo struct {
	Id              uint64 `gorm:"primary_key"`
	MinerId         string
	SectorId        uint64
	SealRandEpoch   int64
	ExpirationEpoch int64
	CreateTime      time.Time
}

func (SectorExpirationInfo) TableName() string {
	return "t_sector_expiration_info"
}

func (this *SectorExpirationInfo) Create(ctx context.Context, sei *SectorExpirationInfo) error {
	var res SectorExpirationInfo
	err := ipfsunion.MysqlDB.Model(&SectorExpirationInfo{}).
		Where("sector_id = ?", sei.SectorId).
		Take(&res).Error
	if err == nil && res.SectorId > 0 {
		return errors.New("sector info already exist")
	}

	sei.CreateTime = time.Now().Truncate(time.Second)
	return ipfsunion.MysqlDB.Create(sei).Error
}

func (this *SectorExpirationInfo) UpdateSealRandEpochAndExpirationEpoch(ctx context.Context, sectorId uint64, sealSeedEpoch int64, expirationEpoch int64) error {
	return ipfsunion.MysqlDB.Model(&SectorExpirationInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"seal_rand_epoch": sealSeedEpoch, "expiration_epoch": expirationEpoch}).Error
}
