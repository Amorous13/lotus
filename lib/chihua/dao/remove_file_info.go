package dao

import (
	"context"
	"errors"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"

	"golang.org/x/xerrors"
)

type RemoveFileInfo struct {
	Id            uint64 `gorm:"primary_key"`
	SectorId      uint64
	MinerId       string
	WorkerAddress string
	Cache         string
	State         string
	CreateTime    time.Time
}

func (RemoveFileInfo) TableName() string {
	return "t_remove_file_info"
}

func (thiz *RemoveFileInfo) Upsert(ctx context.Context, cfInfo *RemoveFileInfo) (*RemoveFileInfo, bool, error) {
	if cfInfo == nil || cfInfo.MinerId == "" ||
		cfInfo.State == "" || cfInfo.WorkerAddress == "" ||
		!CheckRemoveFileState(cfInfo.State) {
		return nil, false, errors.New("invalid input")
	}

	cfInfo.CreateTime = time.Now().Truncate(time.Second)

	if res, err := checkIfRemoveFileInfoExists(cfInfo.SectorId, cfInfo.WorkerAddress); err == nil {
		return res, true, thiz.Update(ctx, cfInfo)
	}

	errInsert := ipfsunion.MysqlDB.Create(cfInfo).Error
	return cfInfo, false, errInsert
}

func (this *RemoveFileInfo) Update(ctx context.Context, rfi *RemoveFileInfo) error {
	if rfi == nil {
		return errors.New("invalid input")
	}

	return ipfsunion.MysqlDB.Model(&RemoveFileInfo{}).
		Where("sector_id = ? AND worker_address = ?", rfi.SectorId, rfi.WorkerAddress).
		Updates(rfi).Error
}

func (thiz *RemoveFileInfo) GetNotListLimit(ctx context.Context, workerAddress string, limit int) ([]*RemoveFileInfo, error) {
	return thiz.getListByStateLimit(workerAddress, SectorRemoveStatusNot, limit)
}

func (thiz *RemoveFileInfo) GetDoingListLimit(ctx context.Context, workerAddress string, limit int) ([]*RemoveFileInfo, error) {
	return thiz.getListByStateLimit(workerAddress, SectorRemoveStatusDoing, limit)
}

func (thiz *RemoveFileInfo) UpdateStateNot(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateRemoveFileInfoState(sectorId, workerAddress, SectorRemoveStatusNot)
}

func (thiz *RemoveFileInfo) UpdateStateDoing(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateRemoveFileInfoState(sectorId, workerAddress, SectorRemoveStatusDoing)
}

func (thiz *RemoveFileInfo) UpdateStateFinished(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateRemoveFileInfoState(sectorId, workerAddress, SectorRemoveStatusFinished)
}

func (thiz *RemoveFileInfo) UpdateStateFail(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateRemoveFileInfoState(sectorId, workerAddress, SectorRemoveStatusFail)
}

// ---------------------------------------------------------------------------------------------------------------------

func checkIfRemoveFileInfoExists(sectorId uint64, workerAddress string) (*RemoveFileInfo, error) {
	var res RemoveFileInfo
	err := ipfsunion.MysqlDB.Model(&RemoveFileInfo{}).
		Where("sector_id = ? AND worker_address = ?", sectorId, workerAddress).
		Take(&res).Error
	return &res, err
}

func updateRemoveFileInfoState(sectorId uint64, workerAddress string, state string) error {
	if !CheckRemoveFileState(state) {
		return xerrors.Errorf("unknown state %v", state)
	}
	if _, err := checkIfRemoveFileInfoExists(sectorId, workerAddress); err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&RemoveFileInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": state}).Error
}

func (thiz *RemoveFileInfo) getListByStateLimit(workerAddress string, state string, limit int) ([]*RemoveFileInfo, error) {
	var result []*RemoveFileInfo
	if err := ipfsunion.MysqlDB.Model(&RemoveFileInfo{}).
		Where("state = ? AND worker_address = ?", state, workerAddress).
		Order("create_time ASC").
		Limit(limit).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*RemoveFileInfo, 0)
	}
	return result, nil
}
