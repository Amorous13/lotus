package dao

import (
	"context"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type MessageBatchInfo struct {
	Id         uint64 `gorm:"primary_key"`
	MsgCid     string
	State      MessageState
	CreateTime time.Time
}

var MessageBatchInfoTable = MessageBatchInfo{}

func (mb MessageBatchInfo) TableName() string {
	return "t_message_batch_info"
}

func (mb MessageBatchInfo) Create(ctx context.Context) (*MessageBatchInfo, error) {
	var obj = MessageBatchInfo{CreateTime: time.Now().Truncate(time.Second)}
	if err := ipfsunion.MysqlDB.Create(&obj).Error; err != nil {
		return nil, err
	}

	return &obj, nil
}
