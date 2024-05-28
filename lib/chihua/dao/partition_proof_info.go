package dao

import (
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type PartitionProofInfo struct {
	MinerId      string
	DeadlineOpen int64
	Partitions   string // 逗号分隔，升序排列，partition序号集合
	Proof        string
	CreateTime   time.Time
}

var PartitionProofInfoTable PartitionProofInfo

func (this *PartitionProofInfo) TableName() string {
	return "t_partition_proof_info"
}

func (this *PartitionProofInfo) GetListByDeadlineOpen(open int64) ([]*PartitionProofInfo, error) {
	var result []*PartitionProofInfo
	err := ipfsunion.MysqlDB.Model(&PartitionProofInfo{}).
		Where("deadline_open = ?", open).
		Order("create_time ASC").
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*PartitionProofInfo, 0)
	}
	return result, nil
}

func (this *PartitionProofInfo) Create(open int64, partitions string, proof string) error {
	var info = &PartitionProofInfo{
		DeadlineOpen: open,
		Partitions:   partitions,
		Proof:        proof,
		CreateTime:   time.Now().Truncate(time.Second),
	}
	return ipfsunion.MysqlDB.Create(info).Error
}

func (thiz *PartitionProofInfo) Upsert(open int64, partitions string, proof string, minerId string) error {
	var info = &PartitionProofInfo{
		MinerId:      minerId,
		DeadlineOpen: open,
		Partitions:   partitions,
		Proof:        proof,
		CreateTime:   time.Now().Truncate(time.Second),
	}
	if _, err := checkIfPartitionProofInfoExists(open, partitions); err == nil {
		return thiz.Update(info)
	}
	return ipfsunion.MysqlDB.Create(info).Error
}

func (thiz *PartitionProofInfo) Update(ppi *PartitionProofInfo) error {
	return ipfsunion.MysqlDB.Model(&PartitionProofInfo{}).
		Where("deadline_open = ? AND partitions = ?", ppi.DeadlineOpen, ppi.Partitions).
		Updates(ppi).Error
}

func checkIfPartitionProofInfoExists(open int64, partitions string) (*PartitionProofInfo, error) {
	var res PartitionProofInfo
	err := ipfsunion.MysqlDB.Model(&PartitionProofInfo{}).
		Where("deadline_open = ? AND partitions = ?", open, partitions).
		Take(&res).Error
	return &res, err
}
