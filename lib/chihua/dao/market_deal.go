package dao

import (
	"context"
	"errors"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"time"
)

type MarketDealInfo struct {
	Id         uint64 `gorm:"primary_key"`
	PayloadCid string
	PieceCid   string
	DealId     uint64
	SectorId   uint64
	Offset     uint64
	Length     uint64
	CreateTime time.Time
}

func (MarketDealInfo) TableName() string {
	return "t_market_deal_info"
}

func (this *MarketDealInfo) GetByPayloadCid(ctx context.Context, payloadCid string) ([]MarketDealInfo, error) {
	var result []MarketDealInfo
	err := ipfsunion.MysqlDB.Model(&MarketDealInfo{}).Where("payload_cid = ?", payloadCid).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (this *MarketDealInfo) GetByPieceCid(ctx context.Context, pieceCid string) ([]MarketDealInfo, error) {
	var result []MarketDealInfo
	err := ipfsunion.MysqlDB.Model(&MarketDealInfo{}).Where("piece_cid = ?", pieceCid).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (this *MarketDealInfo) Create(ctx context.Context, dealInfo *MarketDealInfo) error {
	if dealInfo == nil || dealInfo.PieceCid == "" || dealInfo.PayloadCid == "" {
		return errors.New("invalid input")
	}
	dealInfo.CreateTime = time.Now()
	if err := ipfsunion.MysqlDB.Create(dealInfo).Error; err != nil {
		return err
	}
	return nil
}
