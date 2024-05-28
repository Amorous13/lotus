package dao

import (
	"context"
	"errors"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type CopyFileInfo struct {
	Id            uint64 `gorm:"primary_key"`
	SectorId      uint64
	MinerId       string
	WorkerAddress string
	Cache         string
	Unsealed      string
	Sealed        string
	State         string
	RetryTimes    int
	CreateTime    time.Time
}

func (CopyFileInfo) TableName() string {
	return "t_copy_file_info"
}

func (this *CopyFileInfo) Get(ctx context.Context, sectorId uint64, minerId string) (*CopyFileInfo, error) {
	var result CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).Where("sector_id = ?", sectorId).Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *CopyFileInfo) Update(ctx context.Context, cfi *CopyFileInfo) error {
	if cfi == nil {
		return errors.New("invalid input")
	}
	return ipfsunion.MysqlDB.Model(&CopyFileInfo{}).Where("sector_id = ?", cfi.SectorId).Updates(cfi).Error
}

func (this *CopyFileInfo) Insert(ctx context.Context, cfInfo *CopyFileInfo) (*CopyFileInfo, error) {
	if cfInfo == nil || cfInfo.MinerId == "" ||
		cfInfo.State == "" || cfInfo.WorkerAddress == "" ||
		cfInfo.Sealed == "" || cfInfo.Cache == "" {
		return nil, errors.New("invalid input")
	}

	var res CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", cfInfo.SectorId).
		Take(&res).Error
	if err == nil {
		return nil, errors.New("sector CopyFileInfo already exist")
	}

	cfInfo.State = SectorMoveStatusNotMove
	cfInfo.CreateTime = time.Now().Truncate(time.Second)
	err = ipfsunion.MysqlDB.Create(cfInfo).Error
	if err != nil {
		return nil, err
	}
	return cfInfo, nil
}

func (this *CopyFileInfo) GetNotMoveListLimit(ctx context.Context, minerId string, ip string, limit int) ([]*CopyFileInfo, error) {
	return this.getListByStateLimit(ctx, minerId, ip, SectorMoveStatusNotMove, limit)
}

func (this *CopyFileInfo) GetMovingListLimit(ctx context.Context, minerId string, ip string, limit int) ([]*CopyFileInfo, error) {
	return this.getListByStateLimit(ctx, minerId, ip, SectorMoveStatusMoving, limit)
}

func (_ *CopyFileInfo) UpdateSectorForMoveStateNotMoved(ctx context.Context, sectorId uint64, minerId string, retryTimes int) error {
	var res CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{
			"state":       SectorMoveStatusNotMove,
			"retry_times": retryTimes}).Error
}

func (_ *CopyFileInfo) UpdateSectorForMoveStateMoved(ctx context.Context, sectorId uint64, minerId string) error {
	var res CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": SectorMoveStatusMoved}).Error
}

func (_ *CopyFileInfo) UpdateSectorForMoveStateMoving(ctx context.Context, sectorId uint64, minerId string) error {
	var res CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": SectorMoveStatusMoving}).Error
}

func (_ *CopyFileInfo) UpdateSectorForMoveStateFail(ctx context.Context, sectorId uint64, minerId string) error {
	var res CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": SectorMoveStatusFailed}).Error
}

// ---------------------------------------------------------------------------------------------------------------------

func (_ *CopyFileInfo) getListByStateLimit(ctx context.Context, minerId string, ip string, state string, limit int) ([]*CopyFileInfo, error) {
	var result []*CopyFileInfo
	err := ipfsunion.MysqlDB.Model(&CopyFileInfo{}).
		Where("state = ? AND worker_address = ?",
			state, ip).
		Order("create_time ASC").
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*CopyFileInfo, 0)
	}
	return result, nil
}
