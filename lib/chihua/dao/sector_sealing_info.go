package dao

import (
	"context"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"time"
)

type SealingInfoResultType int

const (
	SealingInfoResultInit       SealingInfoResultType = iota + 1 // 1
	SealingInfoResultProcessing                                  // 2
	SealingInfoResultSuccess                                     // 3
	SealingInfoResultFailure                                     // 4
)

type SectorSealingInfo struct {
	Id            uint64
	MinerId       string
	SectorId      uint64
	State         string
	WorkerAddress string                // worker地址，格式：SH-172-17-28-55-WORK，加索引
	Result        SealingInfoResultType // 任务的执行结果（1、未开始；2、进行中；3、成功；4、失败）
	StartTime     time.Time
	EndTime       time.Time
}

func (SectorSealingInfo) TableName() string {
	return "t_sector_sealing_info"
}

func (this *SectorSealingInfo) Create(ctx context.Context, ssi *SectorSealingInfo) error {
	//var res SectorSealingInfo
	//err := ipfsunion.MysqlDB.Model(&SectorSealingInfo{}).
	//	Where("sector_id = ? AND state = ?", ssi.SectorId, ssi.State).
	//	Take(&res).Error
	//if err == nil && res.Id > 0 {
	//	return nil
	//}
	ssi.StartTime = time.Now().Truncate(time.Second)
	ssi.EndTime = ssi.StartTime
	ssi.Result = SealingInfoResultProcessing
	return ipfsunion.MysqlDB.Create(ssi).Error
}

func (this *SectorSealingInfo) FinishTask(ctx context.Context, result SealingInfoResultType, state string, sectorId uint64, minerId string) error {
	return ipfsunion.MysqlDB.Model(&SectorSealingInfo{}).
		Where("sector_id = ? AND state = ?", sectorId, state).
		Updates(map[string]interface{}{"result": result, "end_time": time.Now().Truncate(time.Second)}).Error
}

func (this *SectorSealingInfo) RestartTaskForStatistic(ctx context.Context, state string, sectorId uint64, result SealingInfoResultType, minerId string) error {
	return ipfsunion.MysqlDB.Model(&SectorSealingInfo{}).
		Where("sector_id = ? AND state = ?", sectorId, state).
		Updates(map[string]interface{}{"result": result, "end_time": time.Now().Truncate(time.Second)}).Error
}
