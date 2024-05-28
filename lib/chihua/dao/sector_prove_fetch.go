package dao

import (
	"context"
	"errors"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"time"
)

type SectorProveFetchInfo struct {
	Id          uint64 `gorm:"primary_key"`
	MinerId     string
	LastProveId uint64
	MinerType   string
	IpAddress   string
	CreateTime  time.Time
}

func (SectorProveFetchInfo) TableName() string {
	return "t_sector_prove_fetch_info"
}

func (this *SectorProveFetchInfo) Update(ctx context.Context, info *SectorProveFetchInfo) error {
	if info == nil {
		return errors.New("invalid input")
	}
	return ipfsunion.MysqlDB.Model(&SectorProveFetchInfo{}).Where("id = ?", info.Id).Updates(info).Error
}

func (this *SectorProveFetchInfo) AddOrUpdateProveFetchInfo(ctx context.Context, lastProved uint64, minerId, minerType, ipAddress string) error {
	var res SectorProveFetchInfo

	err := ipfsunion.MysqlDB.Model(&SectorProveFetchInfo{}).
		Where("miner_type = ?", minerType).
		Take(&res).Error

	res.MinerId = minerId
	res.LastProveId = lastProved
	res.MinerType = minerType
	res.IpAddress = ipAddress
	if err == nil && res.Id > 0 {
		return ipfsunion.MysqlDB.Model(&SectorProveFetchInfo{}).
			Where("miner_type = ?", minerType).
			Update(map[string]interface{}{"miner_id": minerId, "last_prove_id": lastProved, "miner_type": minerType, "ip_address": ipAddress}).Error
	}
	res.CreateTime = time.Now().Truncate(time.Second)
	return ipfsunion.MysqlDB.Create(&res).Error
}

func (this *SectorProveFetchInfo) GetProveFetchInfo(ctx context.Context, minerId, minerType string) (SectorProveFetchInfo, error) {
	var result SectorProveFetchInfo
	err := ipfsunion.MysqlDB.Model(&SectorProveFetchInfo{}).Where("miner_type = ?", minerType).Take(&result).Error
	if err != nil {
		return SectorProveFetchInfo{}, err
	}
	return result, nil
}
