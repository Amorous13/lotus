package dao

import (
	"context"
	"errors"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"

	"golang.org/x/xerrors"
)

type C2inFileInfo struct {
	Id            uint64 `gorm:"primary_key"`
	SectorId      uint64
	MinerId       string
	WorkerAddress string
	Path          string
	State         string
	CreateTime    time.Time
}

func (C2inFileInfo) TableName() string {
	return "t_c2in_file_info"
}

func (thiz *C2inFileInfo) Upsert(ctx context.Context, cfInfo *C2inFileInfo) (*C2inFileInfo, bool, error) {
	if cfInfo == nil || cfInfo.MinerId == "" ||
		cfInfo.State == "" || cfInfo.WorkerAddress == "" ||
		!CheckC2inFileState(cfInfo.State) {
		return nil, false, errors.New("invalid input")
	}

	cfInfo.CreateTime = time.Now().Truncate(time.Second)

	if res, err := checkIfC2inFileInfoExists(cfInfo.SectorId, cfInfo.WorkerAddress); err == nil {
		return res, true, thiz.Update(ctx, cfInfo)
	}

	errInsert := ipfsunion.MysqlDB.Create(cfInfo).Error
	return cfInfo, false, errInsert
}

func (this *C2inFileInfo) Update(ctx context.Context, rfi *C2inFileInfo) error {
	if rfi == nil {
		return errors.New("invalid input")
	}

	return ipfsunion.MysqlDB.Model(&C2inFileInfo{}).
		Where("sector_id = ? AND worker_address = ?", rfi.SectorId, rfi.WorkerAddress).
		Updates(rfi).Error
}

func (thiz *C2inFileInfo) GetListLimitByStateNot(ctx context.Context, workerAddress string, limit int) ([]*C2inFileInfo, error) {
	return thiz.getListByStateLimit(workerAddress, SectorC2inStatusNot, limit)
}

func (thiz *C2inFileInfo) GetListLimitByStateDoing(ctx context.Context, workerAddress string, limit int) ([]*C2inFileInfo, error) {
	return thiz.getListByStateLimit(workerAddress, SectorC2inStatusDoing, limit)
}

func (thiz *C2inFileInfo) UpdateStateNot(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateC2inFileInfoState(sectorId, workerAddress, SectorC2inStatusNot)
}

func (thiz *C2inFileInfo) UpdateStateDoing(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateC2inFileInfoState(sectorId, workerAddress, SectorC2inStatusDoing)
}

func (thiz *C2inFileInfo) UpdateStateFinished(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateC2inFileInfoState(sectorId, workerAddress, SectorC2inStatusFinished)
}

func (thiz *C2inFileInfo) UpdateStateFail(ctx context.Context, sectorId uint64, workerAddress string) error {
	return updateC2inFileInfoState(sectorId, workerAddress, SectorC2inStatusFail)
}

// ---------------------------------------------------------------------------------------------------------------------

func checkIfC2inFileInfoExists(sectorId uint64, workerAddress string) (*C2inFileInfo, error) {
	var res C2inFileInfo
	err := ipfsunion.MysqlDB.Model(&C2inFileInfo{}).
		Where("sector_id = ? AND worker_address = ?", sectorId, workerAddress).
		Take(&res).Error
	return &res, err
}

func updateC2inFileInfoState(sectorId uint64, workerAddress string, state string) error {
	if !CheckC2inFileState(state) {
		return xerrors.Errorf("unknown state %v", state)
	}
	if _, err := checkIfC2inFileInfoExists(sectorId, workerAddress); err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&C2inFileInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": state}).Error
}

func (thiz *C2inFileInfo) getListByStateLimit(workerAddress string, state string, limit int) ([]*C2inFileInfo, error) {
	var result []*C2inFileInfo
	if err := ipfsunion.MysqlDB.Model(&C2inFileInfo{}).
		Where("state = ? AND worker_address = ?", state, workerAddress).
		Order("create_time ASC").
		Limit(limit).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*C2inFileInfo, 0)
	}
	return result, nil
}
