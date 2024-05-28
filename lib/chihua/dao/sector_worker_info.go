package dao

import (
	"context"
	"errors"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"time"
)

type SectorWorkerInfo struct {
	Id            uint64
	MinerId       string
	SectorId      uint64
	WorkerAddress string // worker地址，格式：SH-172-17-28-55-WORK，加索引
	CreateTime    time.Time
}

func (SectorWorkerInfo) TableName() string {
	return "t_sector_worker_info"
}

func (this *SectorWorkerInfo) Create(ctx context.Context, smi *SectorWorkerInfo) error {
	var res SectorWorkerInfo
	err := ipfsunion.MysqlDB.Model(&SectorWorkerInfo{}).
		Where("sector_id = ?", smi.SectorId).
		Take(&res).Error
	if err == nil && res.Id > 0 {
		return nil
	}
	smi.CreateTime = time.Now().Truncate(time.Second)
	return ipfsunion.MysqlDB.Create(smi).Error
}

func (this *SectorWorkerInfo) Update(ctx context.Context, smi *SectorWorkerInfo) error {
	if smi == nil {
		return errors.New("invalid input")
	}
	return ipfsunion.MysqlDB.Model(&SectorWorkerInfo{}).Where("sector_id = ?", smi.SectorId).Updates(smi).Error
}

func (this *SectorWorkerInfo) GetBySectorId(ctx context.Context, sectorId uint64, minerId string) (*SectorWorkerInfo, error) {
	var res SectorWorkerInfo
	err := ipfsunion.MysqlDB.Model(&SectorWorkerInfo{}).
		Where("sector_id = ?", sectorId).
		Find(&res).Error
	return &res, err
}
