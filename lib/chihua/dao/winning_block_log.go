package dao

import (
	"context"
	"errors"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type WinningBlockLog struct {
	Id            uint64 `gorm:"primary_key"`
	Height        uint64
	Miner         string
	Cid           string
	WorkerAddress string
	Error         string
	CreateTime    time.Time
}

func (WinningBlockLog) TableName() string {
	return "t_winningblock_log"
}

func NewWinningBlockLog() *WinningBlockLog {
	return new(WinningBlockLog)
}

func (*WinningBlockLog) Create(ctx context.Context, wbl *WinningBlockLog) (*WinningBlockLog, error) {
	wbl.CreateTime = time.Now().Truncate(time.Microsecond)

	if ipfsunion.MysqlDB == nil {
		return nil, errors.New("can't create WinningBlockLog: ipfsunion.MysqlDB nil")
	}

	if err := ipfsunion.MysqlDB.Create(wbl).Error; err != nil {
		return nil, err
	}
	return wbl, nil
}
