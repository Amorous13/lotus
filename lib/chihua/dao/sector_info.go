package dao

import (
	"context"
	"errors"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

var log = logging.Logger("sectors")

type C2DispatcherStatusType int

const (
	C2DispatcherStatusInit        C2DispatcherStatusType = iota + 1 // 1
	C2DispatcherStatusProcessing                                    // 2
	C2DispatcherStatusSuccess                                       // 3
	C2DispatcherStatusFailure                                       // 4
	C2DispatcherStatusReady                                         // 5
	C2DispatcherStatusOutsourcing                                   // 6 正在外包中
)

const (
	SectorMoveStatusDefault = ""
	SectorMoveStatusNotMove = "not"
	SectorMoveStatusMoving  = "doing"
	SectorMoveStatusMoved   = "finished"
	SectorMoveStatusFailed  = "fail"
)

const (
	SectorRemoveStatusNot      = "not"
	SectorRemoveStatusDoing    = "doing"
	SectorRemoveStatusFinished = "finished"
	SectorRemoveStatusFail     = "fail"
)

var removeStatusMap = map[string]struct{}{
	SectorRemoveStatusNot:      struct{}{},
	SectorRemoveStatusDoing:    struct{}{},
	SectorRemoveStatusFinished: struct{}{},
	SectorRemoveStatusFail:     struct{}{},
}

func CheckRemoveFileState(state string) bool {
	_, ok := removeStatusMap[state]
	return ok
}

const (
	SectorC2inStatusNot      = "not"
	SectorC2inStatusDoing    = "doing"
	SectorC2inStatusFinished = "finished"
	SectorC2inStatusFail     = "fail"
)

var c2inStatusMap = map[string]struct{}{
	SectorC2inStatusNot:      struct{}{},
	SectorC2inStatusDoing:    struct{}{},
	SectorC2inStatusFinished: struct{}{},
	SectorC2inStatusFail:     struct{}{},
}

func CheckC2inFileState(state string) bool {
	_, ok := c2inStatusMap[state]
	return ok
}

type SectorInfo struct {
	SectorId     uint64 `gorm:"primary_key"`
	MinerId      string
	State        string
	MinerAddress string
	StorageId    string
	CommD        string
	CommR        string
	TicketValue  string
	TicketEpoch  int64
	SeedValue    string
	SeedEpoch    int64
	Proof        string
	//RetryTimes          uint64
	//Deals               string
	//SectorType          int64
	PreCommit1Out       string
	PreCommit1ErrorInfo string
	PreCommit2ErrorInfo string
	Commit1ErrorInfo    string
	Commit2ErrorInfo    string
	//SeedErrorInfo       string
	PreCommitMsgCid string
	//PreCommit2FailTimes int
	CommP              string
	CommitMsgCid       string
	Commit1Out         string
	C2DispatcherStatus C2DispatcherStatusType
	//Commit2FailTimes   int
	//FaultReportMsgCid   string
	CreateTime time.Time
}

func (SectorInfo) TableName() string {
	return "t_sector_info"
}

func (this *SectorInfo) Get(ctx context.Context, id uint64) (*SectorInfo, error) {
	var result SectorInfo
	err := ipfsunion.MysqlDB.Model(&SectorInfo{}).Where("sector_id = ?", id).Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// 批量获取Ready状态的c2任务，并置状态为outsourcing，返回的数据长度可能比limit小
func (this *SectorInfo) GetListC2TodoTask(ctx context.Context, minerId string, limit int) ([]*SectorInfo, error) {
	var (
		result    []*SectorInfo
		sectorIds []uint64
		err       error
		fields    = []string{"sector_id", "miner_id", "c2_dispatcher_status", "commit1_out", "sys_auto_update_time"}
	)

	err = ipfsunion.MysqlDB.
		Select(fields).
		Where("c2_dispatcher_status = ?", C2DispatcherStatusReady).Order("sector_id ASC").
		Limit(limit).Find(&result).Error
	if err == nil {
		for _, s := range result {
			sectorIds = append(sectorIds, s.SectorId)
		}
		if len(sectorIds) > 0 {
			db := ipfsunion.MysqlDB.Model(&SectorInfo{}).
				Set("gorm:query_option", "FOR_UPDATE").
				Where("sector_id IN (?) AND c2_dispatcher_status = ?", sectorIds, C2DispatcherStatusReady).
				Update(map[string]interface{}{
					"c2_dispatcher_status": C2DispatcherStatusOutsourcing,
					"state":                "C2DispatcherStatusOutsourcing"})
			if db.Error != nil {
				return nil, xerrors.Errorf("update sector's c2_dispatcher_status error, err: %v", db.Error)
			}

			rowsAffected := db.RowsAffected
			if rowsAffected == 0 {
				log.Errorf("ipfsunion::SectorInfo::GetListC2TodoTask failed, roll back, MinerId:[%v], rowsAffected:[%v], error:[%v]", minerId, rowsAffected, err)
				return nil, errors.New("update sector's c2_dispatcher_status failed")
			}
			// 只有部分更新成功，取出这部分返回
			if rowsAffected < int64(len(sectorIds)) {
				result = []*SectorInfo{}
				err = ipfsunion.MysqlDB.
					Select(fields).
					Where("sector_id IN (?) AND c2_dispatcher_status = ?", sectorIds, C2DispatcherStatusOutsourcing).
					Order("sector_id ASC").
					Limit(limit).Find(&result).Error
			}
			if err == nil {
				log.Infof("ipfsunion::SectorInfo::GetListC2TodoTask success, MinerId:[%v], rowsAffected:[%v]", minerId, rowsAffected)
			}
		}
		return result, err
	}

	log.Error("ipfsunion::SectorInfo::GetListC2TodoTask failed, sectors to do C2 don't exist, error")
	return nil, errors.New("GetListC2TodoTask sectors to do C2 don't exist")
}

func (this *SectorInfo) Update(ctx context.Context, si *SectorInfo) error {
	if si == nil {
		return errors.New("invalid input")
	}
	return ipfsunion.MysqlDB.Model(&SectorInfo{}).Where("sector_id = ?", si.SectorId).Updates(si).Error
}

func (this *SectorInfo) GetBySectorId(ctx context.Context, sectorId uint64, minerId string) (*SectorInfo, error) {
	var res SectorInfo
	err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Find(&res).Error
	return &res, err
}

// addpiece 调用插入扇区信息
func (this *SectorInfo) CreateSector(ctx context.Context, info *SectorInfo) (*SectorInfo, error) {
	if info == nil || info.MinerId == "" {
		return nil, errors.New("invalid input")
	}

	var res SectorInfo
	err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", info.SectorId).
		Take(&res).Error
	if err == nil && res.SectorId > 0 {
		return nil, errors.New("sector info already exist")
	}
	info.State = "SectorCreated"
	info.CreateTime = time.Now().Truncate(time.Second)
	err = ipfsunion.MysqlDB.Create(info).Error
	if err != nil {
		return nil, err
	}
	return info, nil
}

// 重启后retry P1统计校正
func (this *SectorInfo) UpdateSectorForP1Retry(ctx context.Context, sectorId uint64, minerId string) error {

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": "SectorCreated"}).Error
}

// P1成功调用，记录p1输出
func (this *SectorInfo) UpdateSectorForP1Success(ctx context.Context, sectorId uint64, minerId, p1Out, ticketValue string, ticketEpoch int64) error {

	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit1_out": p1Out, "ticket_value": ticketValue, "ticket_epoch": ticketEpoch, "state": "PreCommit1Success"}).Error
}

// 重启程序调用，校正统计信息
func (this *SectorInfo) UpdateSectorForP1RestartStatistic(ctx context.Context, sectorId uint64, minerId, comment string) error {

	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit1_error_info": comment}).Error
}

// P1失败调用，记录p1失败信息
func (this *SectorInfo) UpdateSectorForP1Failed(ctx context.Context, sectorId uint64, minerId string, p1ErrorInfo string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit1_error_info": p1ErrorInfo, "state": "PreCommit1Failed"}).Error
}

// P2成功调用，记录p2输出
func (this *SectorInfo) UpdateSectorForP2Success(ctx context.Context, sectorId uint64, minerId, commDCid, commRCid string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"comm_d": commDCid, "comm_r": commRCid, "state": "PreCommit2Success"}).Error
}

// 重启程序调用，校正统计信息
func (this *SectorInfo) UpdateSectorForP2RestartStatistic(ctx context.Context, sectorId uint64, minerId, comment string) error {

	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit2_error_info": comment}).Error
}

// P2失败调用，记录p2失败信息
func (this *SectorInfo) UpdateSectorForP2Failed(ctx context.Context, sectorId uint64, minerId, p2ErrorInfo string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit2_error_info": p2ErrorInfo, "state": "PreCommit2Failed"}).Error
}

// PreCommitting成功调用
func (this *SectorInfo) UpdateSectorForPreCommittingSuccess(ctx context.Context, sectorId uint64, minerId, msgCid string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": "PreCommittingSuccess", "pre_commit_msg_cid": msgCid}).Error
}

// PreCommitting失败调用
func (this *SectorInfo) UpdateSectorForPreCommittingFailed(ctx context.Context, sectorId uint64, minerId, preCommitErrorInfo string) error {
	if minerId == "" || preCommitErrorInfo == "" {
		return errors.New("invalid input")
	}

	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit2_error_info": preCommitErrorInfo, "state": "PreCommittingFailed"}).Error
}

// PreCommit成功调用
func (this *SectorInfo) UpdateSectorForPreCommitWaitSuccess(ctx context.Context, sectorId uint64, minerId string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": "PreCommitWaitSuccess"}).Error
}

// PreCommit失败调用
func (this *SectorInfo) UpdateSectorForPreCommitWaitFailed(ctx context.Context, sectorId uint64, minerId, preCommitWaitErrorInfo string) error {
	//if minerId == "" || preCommitWaitErrorInfo == "" {
	//	return errors.New("invalid input")
	//}
	//
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"pre_commit2_error_info": preCommitWaitErrorInfo, "state": "PreCommitWaitFailed"}).Error
}

// WaitSeed成功调用，记录WaitSeed输出
func (this *SectorInfo) UpdateWaitSeedSuccess(ctx context.Context, sectorId uint64, minerId, seedValue string, seedEpoch uint64) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"seed_value": seedValue, "seed_epoch": seedEpoch, "state": "WaitSeedSuccess"}).Error
}

// WaitSeed失败调用，记录WaitSeed失败信息
func (this *SectorInfo) UpdateWaitSeedFailed(ctx context.Context, sectorId uint64, minerId, waitSeedErrorInfo string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"seed_error_info": waitSeedErrorInfo, "state": "WaitSeedFailed"}).Error
}

// C1成功调用，记录C1输出
func (this *SectorInfo) UpdateSectorForC1Success(ctx context.Context, sectorId uint64, minerId, c1Out string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"commit1_out": c1Out, "c2_dispatcher_status": C2DispatcherStatusInit, "state": "Commit1Success"}).Error
}

// C1失败调用，记录C1失败信息
func (this *SectorInfo) UpdateSectorForC1Failed(ctx context.Context, sectorId uint64, minerId string, c1ErrorInfo string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"commit1_error_info": c1ErrorInfo, "state": "Commit1Failed"}).Error
}

// C2成功调用，记录C2输出
func (this *SectorInfo) UpdateSectorForC2Success(ctx context.Context, sectorId uint64, minerId, proof string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"proof": proof, "state": "Commit2Success", "c2_dispatcher_status": C2DispatcherStatusSuccess}).Error
}

// C2失败调用，记录C2失败信息
func (this *SectorInfo) UpdateSectorForC2Failed(ctx context.Context, sectorId uint64, minerId, c2ErrorInfo string) error {
	//if minerId == "" || c2ErrorInfo == "" {
	//	return errors.New("invalid input")
	//}
	//
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"commit2_error_info": c2ErrorInfo, "state": "Commit2Failed", "c2_dispatcher_status": C2DispatcherStatusFailure}).Error
}

// SubmitCommit成功调用
func (this *SectorInfo) UpdateSectorForSubmitCommitSuccess(ctx context.Context, sectorId uint64, minerId, msgCid string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": "SubmitCommitSuccess", "commit_msg_cid": msgCid}).Error
}

// SubmitCommit失败调用，记录SubmitCommit失败信息
func (this *SectorInfo) UpdateSectorForSubmitCommitFailed(ctx context.Context, sectorId uint64, minerId, submitCommitErrorInfo string) error {
	//if minerId == "" || submitCommitErrorInfo == "" {
	//	return errors.New("invalid input")
	//}
	//
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"commit2_error_info": submitCommitErrorInfo, "state": "SubmitCommitFailed"}).Error
}

// CommitWait成功调用
func (this *SectorInfo) UpdateSectorForCommitWaitSuccess(ctx context.Context, sectorId uint64, minerId string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": "CommitWaitSuccess"}).Error
}

// Finalize成功调用
func (this *SectorInfo) UpdateSectorForFinalizeSuccess(ctx context.Context, sectorId uint64, minerId string) error {
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"state": "FinalizeSuccess"}).Error
}

// CommitWait失败调用，记录CommitWait失败信息
func (this *SectorInfo) UpdateSectorForCommitWaitFailed(ctx context.Context, sectorId uint64, minerId, commitWaitErrorInfo string) error {
	//if minerId == "" || commitWaitErrorInfo == "" {
	//	return errors.New("invalid input")
	//}
	//
	//var res SectorInfo
	//err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
	//	Where("sector_id = ?", sectorId).
	//	Take(&res).Error
	//if err != nil {
	//	return err
	//}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"commit2_error_info": commitWaitErrorInfo, "state": "CommitWaitFailed"}).Error
}

func (this *SectorInfo) UpdateSectorStorageId(ctx context.Context, sectorId uint64, minerId string, storageId string) error {
	if minerId == "" || storageId == "" {
		return errors.New("invalid input")
	}

	var res SectorInfo
	err := ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return err
	}

	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"storage_id": storageId}).Error
}

func (this *SectorInfo) GetOneC2TodoTask(ctx context.Context, minerId string) (*SectorInfo, error) {
	if minerId == "" {
		return nil, errors.New("invalid input")
	}

	var res SectorInfo
	err := ipfsunion.MysqlDB.
		Select([]string{"sector_id", "miner_id", "c2_dispatcher_status", "commit1_out", "sys_auto_update_time"}).
		Where("c2_dispatcher_status = ?", C2DispatcherStatusReady).Order("sector_id ASC").
		Take(&res).Error

	if err == nil {
		rowsAffected := ipfsunion.MysqlDB.Model(&SectorInfo{}).Set("gorm:query_option", "FOR_UPDATE").Where("sector_id = ? AND c2_dispatcher_status = ?", res.SectorId, C2DispatcherStatusReady).Update(map[string]interface{}{"c2_dispatcher_status": C2DispatcherStatusProcessing, "state": "C2DispatcherStatusProcessing"}).RowsAffected
		if rowsAffected == 0 {
			log.Errorf("ipfsunion::SectorInfo::GetOneC2TodoTask failed, roll back, MinerId:[%v], SectorId:[%v], rowsAffected:[%v], error:[%v]", res.MinerId, res.SectorId, rowsAffected, err)
			return nil, errors.New("update sector's c2_dispatcher_status failed")
		}
		log.Infof("ipfsunion::SectorInfo::GetOneC2TodoTask success, MinerId:[%v], SectorId:[%v], C2DispatcherStatus:[%v], rowsAffected:[%v]", res.MinerId, res.SectorId, res.C2DispatcherStatus, rowsAffected)
		return &res, err
	}

	log.Error("ipfsunion::SectorInfo::GetOneC2TodoTask failed, sectors to do C2 don't exist, error")
	return nil, errors.New("sectors to do C2 don't exist")
}

func (this *SectorInfo) UpdateSectorC2DispatcherStatus(ctx context.Context, sectorId uint64, minerId string, status C2DispatcherStatusType) error {
	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"c2_dispatcher_status": status}).Error
}

func (this *SectorInfo) UpdateSectorC2DispatcherStatusForProcessing(ctx context.Context, sectorId uint64) error {
	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id = ?", sectorId).
		Update(map[string]interface{}{"c2_dispatcher_status": C2DispatcherStatusProcessing, "state": "C2DispatcherStatusProcessing", "commit2_error_info": time.Now().Truncate(time.Second).String()}).Error
}

func (this *SectorInfo) BatchUpdateSectorC2DispatcherStatusForRetryOutsourcing(ctx context.Context, sectorIds []uint64, minerId string) error {
	return ipfsunion.MysqlDB.Model(&SectorInfo{}).
		Where("sector_id in (?)", sectorIds).
		Update(map[string]interface{}{"c2_dispatcher_status": C2DispatcherStatusOutsourcing, "state": "C2DispatcherStatusOutsourcing", "commit2_error_info": "retry outsourcing:" + time.Now().Truncate(time.Second).String()}).Error
}

func (this *SectorInfo) GetSectorInfoByMinerIdAndSectorId(ctx context.Context, sectorId uint64, minerId string) (*SectorInfo, error) {
	if minerId == "" {
		return nil, errors.New("invalid input")
	}

	var res SectorInfo
	err := ipfsunion.MysqlDB.Select([]string{"sector_id", "miner_id", "state", "proof", "commit1_out", "c2_dispatcher_status"}).
		Where("sector_id = ?", sectorId).
		Take(&res).Error
	if err != nil {
		return &SectorInfo{}, err
	}

	return &res, nil
}

func (this *SectorInfo) GetStockedCommit2Sectors(ctx context.Context, state string, upperLimit *time.Time) ([]*SectorInfo, error) {
	var (
		result []*SectorInfo
	)
	db := ipfsunion.MysqlDB.Select([]string{"sector_id", "miner_id", "c2_dispatcher_status", "commit1_out", "sys_auto_update_time"})

	if state != "C2DispatcherStatusOutsourcing" && state != "C2DispatcherStatusProcessing" {
		return nil, errors.New("invalid c2 stocked state")
	}
	db = db.Where("state = ?", state)

	if upperLimit == nil {
		tl := time.Now().Add(-5 * time.Hour)
		upperLimit = &tl
	}
	db = db.Where("sys_auto_update_time < ?", upperLimit.Format("2006-01-02 15:04:05"))

	err := db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}
