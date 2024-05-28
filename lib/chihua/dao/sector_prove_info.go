package dao

import (
	"context"
	"errors"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"time"
)

type SectorProveInfo struct {
	Id           uint64 `gorm:"primary_key"`
	MinerId      string
	SectorId     uint64
	StorageId    string
	IsProved     int // 上链标志，0：未成功上链、1：成功上链
	ProveContent []byte
	CreateTime   time.Time
}

func (SectorProveInfo) TableName() string {
	return "t_sector_prove_info"
}

func (this *SectorProveInfo) AddProveSector(ctx context.Context, info *SectorProveInfo) error {
	if info == nil || info.MinerId == "" {
		return errors.New("invalid input")
	}

	info.CreateTime = time.Now().Truncate(time.Second)
	return ipfsunion.MysqlDB.Create(info).Error
}

// 获取lastProveSectorId之后的十条信息（最多10条），按照Id升序排序
func (this *SectorProveInfo) GetProveSectorList(ctx context.Context, minerId string, lastProveSectorId uint64) ([]SectorProveInfo, error) {

	var res []SectorProveInfo
	err := ipfsunion.MysqlDB.Model(&SectorProveInfo{}).
		Where("id > ?", lastProveSectorId).Order("id ASC").Limit(10).
		Find(&res).Error

	return res, err
}

func (this *SectorProveInfo) GetProveSectorListLimit(ctx context.Context, minerId string, lastProveSectorId uint64, limit int) ([]SectorProveInfo, error) {

	var res []SectorProveInfo
	err := ipfsunion.MysqlDB.Model(&SectorProveInfo{}).
		Where("id > ?", lastProveSectorId).Order("id ASC").Limit(limit).
		Find(&res).Error

	return res, err
}

// 更新is_proved字段
func (this *SectorProveInfo) UpdateSectorIsProved(ctx context.Context, sectorId uint64, minerId string) error {
	var res SectorProveInfo
	err := ipfsunion.MysqlDB.Model(&SectorProveInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&SectorProveInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"is_proved": 1}).Error
}

func (this *SectorProveInfo) GetProveSectorBySectorId(ctx context.Context, sectorId uint64, minerId string) (*SectorProveInfo, error) {
	var res SectorProveInfo
	err := ipfsunion.MysqlDB.Model(&SectorProveInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}
