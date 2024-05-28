package dao

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type WinningBlock struct {
	Height       uint64 `gorm:"primary_key"`
	ParentHeight uint64
	ParentKey    uint64
	Cid          string
	Miner        string
	ParentTS     string
	CreateTime   time.Time
}

func (WinningBlock) TableName() string {
	return "t_winningblock_info"
}

func NewWinningBlockDao() *WinningBlock {
	return new(WinningBlock)
}

func (wb *WinningBlock) Add(height uint64, parentHeight uint64, parentKey uint64, CID string, miner string, parentTS string) error {
	if ipfsunion.MysqlDB == nil {
		return nil
	}

	if CID == "" || miner == "" || parentTS == "" {
		return errors.New("invalid input")
	}

	wbInfo := &WinningBlock{
		Height:       height,
		ParentHeight: parentHeight,
		ParentKey:    parentKey,
		Cid:          CID,
		Miner:        miner,
		ParentTS:     parentTS,
	}

	wbInfo.CreateTime = time.Now().Truncate(time.Microsecond)

	return ipfsunion.MysqlDB.Create(wbInfo).Error
}

func (wb *WinningBlock) GetCIDByHeight(height uint64) (string, error) {
	var res WinningBlock

	if ipfsunion.MysqlDB == nil {
		return "", nil
	}

	err := ipfsunion.MysqlDB.Model(&WinningBlock{}).
		Where("height = ?", height).
		Find(&res).Error

	return res.Cid, err
}

func (wb *WinningBlock) GetCIDByParentHeight(parentHeight uint64) (string, error) {
	var res WinningBlock

	if ipfsunion.MysqlDB == nil {
		return "", nil
	}

	err := ipfsunion.MysqlDB.Model(&WinningBlock{}).
		Where("parent_height = ?", parentHeight).
		Find(&res).Error
	return res.Cid, err
}

func (wb *WinningBlock) GetCIDByParentsTS(parentTS string) (string, error) {
	var res WinningBlock

	if ipfsunion.MysqlDB == nil {
		return "", nil
	}

	err := ipfsunion.MysqlDB.Model(&WinningBlock{}).
		Where("parent_ts = ?", parentTS).
		Find(&res).Error
	return res.Cid, err
}

func (wb *WinningBlock) HasByHeight(height uint64) (bool, error) {
	var res WinningBlock

	if ipfsunion.MysqlDB == nil {
		return false, nil
	}

	err := ipfsunion.MysqlDB.Model(&WinningBlock{}).
		Where("height = ?", height).
		Find(&res).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	if err == nil && res.Height > 0 {
		return true, nil
	}

	return false, err
}

func (wb *WinningBlock) HasCIDByParentHeight(parentHeight uint64) (bool, error) {
	var res WinningBlock

	if ipfsunion.MysqlDB == nil {
		return false, nil
	}

	err := ipfsunion.MysqlDB.Model(&WinningBlock{}).
		Where("parent_height = ?", parentHeight).
		Find(&res).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	if err == nil && res.Height > 0 {
		return true, nil
	}

	return false, err
}

func (wb *WinningBlock) HasCIDByParentsTS(parentTS string) (bool, error) {
	var res WinningBlock

	if ipfsunion.MysqlDB == nil {
		return false, nil
	}

	err := ipfsunion.MysqlDB.Model(&WinningBlock{}).
		Where("parent_ts = ?", parentTS).
		Find(&res).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	if err == nil && res.Height > 0 {
		return true, nil
	}

	return false, err
}
