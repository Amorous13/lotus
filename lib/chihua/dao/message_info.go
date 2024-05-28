package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

type MessageType int

const (
	MessageTypePreCommit   MessageType = 1
	MessageTypeProveCommit             = 2
)

type MessageState string

const (
	MessageStateNotSend MessageState = "not"
	MessageStateWaiting              = "waiting"
	MessageStateFail                 = "fail"
	MessageStateFinish               = "finish"
)

var messageInfoStates = map[MessageState]struct{}{
	MessageStateNotSend: struct{}{},
	MessageStateWaiting: struct{}{},
	MessageStateFail:    struct{}{},
	MessageStateFinish:  struct{}{},
}

type MessageInfo struct {
	Id              uint64 `gorm:"primary_key"`
	SectorId        uint64
	MessageType     MessageType
	From            string
	To              string
	Method          int
	Value           string
	MaxFee          string
	Params          string
	ExpirationEpoch uint64
	MsgCid          string
	State           MessageState
	WorkerAddress   string
	ErrorInfo       string
	MsgLookUp       string
	ExitCode        int
	CreateTime      time.Time
	BatchId         int

	Mcid interface{} `gorm:"-"`
}

var MessageInfoTable = MessageInfo{}

func (mi MessageInfo) TableName() string {
	return "t_sector_message_info"
}

func InitialMessageInfoMsgCid(sectorId uint64, mt MessageType) string {
	return fmt.Sprintf("initial-%v-%v", sectorId, mt)
}

func (this MessageInfo) GetByMsgCid(ctx context.Context, msgCid string) (*MessageInfo, error) {
	if msgCid == "" {
		return nil, fmt.Errorf("msgcid empty")
	}

	var result MessageInfo
	err := ipfsunion.MysqlDB.Model(&MessageInfo{}).Where("msg_cid = ?", msgCid).Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this MessageInfo) CountByState(ctx context.Context, state MessageState) (int, error) {
	if _, ok := messageInfoStates[state]; !ok {
		return 0, fmt.Errorf("invalid state: %v", state)
	}

	var count int
	var err = ipfsunion.MysqlDB.Model(&MessageInfo{}).Where("state = ?", state).Count(&count).Error
	return count, err
}

func (this MessageInfo) CountByStateMessageType(ctx context.Context, state MessageState, mst MessageType) (int, error) {
	if _, ok := messageInfoStates[state]; !ok {
		return 0, fmt.Errorf("invalid state: %v", state)
	}

	var count int
	var err = ipfsunion.MysqlDB.Model(&MessageInfo{}).Select("count(distinct(msg_cid))").Where("state = ? AND message_type = ?", state, mst).Count(&count).Error
	return count, err
}

//func (this MessageInfo) Get(ctx context.Context, sectorId uint64, mt MessageType) (*MessageInfo, error) {
//	var result MessageInfo
//	err := ipfsunion.MysqlDB.Model(&MessageInfo{}).Where("sector_id = ? AND message_type = ?", sectorId, mt).Take(&result).Error
//	if err != nil {
//		return nil, err
//	}
//	return &result, nil
//}

func (this MessageInfo) GetById(ctx context.Context, id uint64) (*MessageInfo, error) {
	var result MessageInfo
	err := ipfsunion.MysqlDB.Model(&MessageInfo{}).Where("id = ?", id).Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (mi MessageInfo) Insert(ctx context.Context, msgInfo *MessageInfo) (*MessageInfo, error) {
	if msgInfo == nil || msgInfo.State == "" || (msgInfo.MessageType != MessageTypePreCommit && msgInfo.MessageType != MessageTypeProveCommit) {
		return nil, errors.New("invalid input")
	}

	var res MessageInfo
	err := ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id = ? AND message_type = ? AND state != ? AND exit_code = 0", msgInfo.SectorId, msgInfo.MessageType, MessageStateFail).
		Take(&res).Error
	if err == nil {
		return &res, nil
	}

	msgInfo.MsgCid = InitialMessageInfoMsgCid(msgInfo.SectorId, msgInfo.MessageType)
	msgInfo.State = MessageStateNotSend
	msgInfo.CreateTime = time.Now().Truncate(time.Second)
	if err := ipfsunion.MysqlDB.Create(msgInfo).Error; err != nil {
		return nil, err
	}
	return msgInfo, nil
}

func (thiz MessageInfo) UpdateStateNot(ctx context.Context, sectorId uint64, mt MessageType) error {
	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id = ? AND message_type = ?", sectorId, mt).
		Update(map[string]interface{}{"state": MessageStateNotSend, "msg_cid": "", "error_info": ""}).
		Error
}

func (thiz MessageInfo) Update(ctx context.Context, sectorId uint64, mt MessageType, data *MessageInfo) (int64, error) {
	db := ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id = ? AND message_type = ?", sectorId, mt).
		Save(data)
	return db.RowsAffected, db.Error
}

func (thiz MessageInfo) UpdateStateWaiting(ctx context.Context, sectorId uint64, mt MessageType, msgCid string) error {
	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id = ? AND message_type = ?", sectorId, mt).
		Update(map[string]interface{}{"state": MessageStateWaiting, "msg_cid": msgCid}).
		Error
}

func (thiz MessageInfo) UpdateStateFinished(ctx context.Context, sectorId uint64, mt MessageType, msgLookUp string, exitCode int) error {
	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id = ? AND message_type = ?", sectorId, mt).
		Update(map[string]interface{}{"state": MessageStateFinish, "msg_look_up": msgLookUp, "exit_code": exitCode}).
		Error
}

func (thiz MessageInfo) UpdateStateFail(ctx context.Context, sectorId uint64, mt MessageType, err string) error {
	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id = ? AND message_type = ?", sectorId, mt).
		Update(map[string]interface{}{"state": MessageStateFail, "error_info": err}).
		Error
}

func (thiz MessageInfo) UpdateBatchStateFailAndRestBatchId(ctx context.Context, sectorIds []uint64, mt MessageType, err string) error {
	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id IN (?) AND message_type = ?", sectorIds, mt).
		Update(map[string]interface{}{"state": MessageStateFail, "error_info": err, "batch_id": 0}).
		Error
}

// 状态为not
func (thiz MessageInfo) GetListLimitByStateNot(ctx context.Context, mt MessageType, limit int) ([]*MessageInfo, error) {
	var result []*MessageInfo
	if err := ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("state = ? AND message_type = ?", MessageStateNotSend, mt).
		Order("sys_auto_update_time ASC").
		Limit(limit).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*MessageInfo, 0)
	}
	return result, nil
}

func (thiz MessageInfo) GetListLimitByStateWaiting(ctx context.Context, mt MessageType, limit int) ([]*MessageInfo, error) {
	return thiz.getListByStateLimit(MessageStateWaiting, mt, limit)
}

func (thiz MessageInfo) GetListLimitByStateWaitingSimple(ctx context.Context, mt MessageType, limit int) ([]*MessageInfo, error) {
	var result []*MessageInfo
	if err := ipfsunion.MysqlDB.Model(&MessageInfo{}).Select("id,msg_cid,batch_id").
		Where("state = ? AND message_type = ?", MessageStateWaiting, mt).
		Order("sys_auto_update_time ASC").
		Limit(limit).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*MessageInfo, 0)
	}
	return result, nil

}

func (thiz MessageInfo) GetListLimitBeforeExpiration(ctx context.Context, mt MessageType, limit int, epoch int64) ([]*MessageInfo, error) {
	var result []*MessageInfo
	if err := ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("state = ? AND message_type = ? AND expiration_epoch <= ?", MessageStateNotSend, mt, epoch).
		Order("sys_auto_update_time ASC").
		Limit(limit).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*MessageInfo, 0)
	}
	return result, nil
}

func (thiz MessageInfo) getListByStateLimit(state MessageState, mt MessageType, limit int) ([]*MessageInfo, error) {
	var result []*MessageInfo
	if err := ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("state = ? AND message_type = ?", state, mt).
		Order("sys_auto_update_time ASC").
		Limit(limit).
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*MessageInfo, 0)
	}
	return result, nil
}

//根据批次获取列表
func (thiz MessageInfo) GetListByBatch(ctx context.Context, batchId int64) ([]*MessageInfo, error) {
	if batchId <= 0 {
		return nil, fmt.Errorf("batch invalid %v", batchId)
	}

	var result []*MessageInfo
	if err := ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("batch_id = ?", batchId).
		Order("sys_auto_update_time ASC").
		Find(&result).Error; err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]*MessageInfo, 0)
	}
	return result, nil
}

//发送消息后，更新批次
func (thiz MessageInfo) UpdateBatch(ctx context.Context, sectorIds []uint64, mt MessageType, batchId uint64) error {
	if len(sectorIds) == 0 {
		return nil
	}

	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("sector_id IN (?) AND message_type = ?", sectorIds, mt).
		Update(map[string]interface{}{"batch_id": batchId}).
		Error
}

func (thiz MessageInfo) UpdateStateFinishedByMcid(ctx context.Context, mcid string, msgLookUp string, exitCode int) error {
	if mcid == "" {
		return nil
	}

	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("msg_cid = ?", mcid).
		Update(map[string]interface{}{"state": MessageStateFinish, "msg_look_up": msgLookUp, "exit_code": exitCode}).
		Error
}

func (thiz MessageInfo) UpdateStateFailByMcid(ctx context.Context, mcid string, err string) error {
	if mcid == "" {
		return nil
	}

	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("msg_cid = ?", mcid).
		Update(map[string]interface{}{"state": MessageStateFail, "error_info": err}).
		Error
}

func (thiz MessageInfo) UpdateStateWaitingByBatch(ctx context.Context, batchId uint64, msgCid string) error {
	return ipfsunion.MysqlDB.Model(&MessageInfo{}).
		Where("batch_id = ?", batchId).
		Update(map[string]interface{}{"state": MessageStateWaiting, "msg_cid": msgCid}).
		Error
}
