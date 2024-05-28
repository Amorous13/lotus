package dao

import (
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"github.com/jinzhu/gorm"
	"golang.org/x/xerrors"
)

const (
	message_is_sent_not  = 0 // is_sent字段取值：未发送
	message_is_sent_sent = 1 // is_sent字段取值：已发送
)

type Message struct {
	Id         uint64 `gorm:"primary_key"`
	MsgCid     string
	Method     string
	From       string
	To         string
	Nonce      uint64
	Value      string
	MaxFee     int64
	ExtraInfo  string
	Params     []byte
	IsSent     int
	FailTimes  int
	CreateTime time.Time
}

func (Message) TableName() string {
	return "t_message"
}

func InsertMessage(m *Message) (uint64, error) {
	var id []uint64
	if err := ipfsunion.MysqlDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		return tx.
			Raw("select LAST_INSERT_ID() as id").
			Pluck("id", &id).
			Error
	}); err != nil {
		return 0, err
	}

	if len(id) == 0 {
		return 0, xerrors.Errorf("InsertMessage Pluck id len 0")
	}
	return id[0], nil
}

func CreateMessage(m *Message) error {
	return ipfsunion.MysqlDB.Create(m).Error
}

func GetMessagesToPush(maxCount uint64) (msgs []Message, err error) {
	err = ipfsunion.MysqlDB.Model(&Message{}).Order("sys_auto_update_time asc").Find(&msgs, "is_sent=? and fail_times<?", 0, 10).Limit(maxCount).Error

	return
}

func GetMessageWithCIDToWait(id uint64) (msg *Message, err error) {
	var msgR Message

	err = ipfsunion.MysqlDB.Model(&Message{}).Find(&msgR, "id=?", id).Error
	msg = &msgR

	return
}

func UpdateMessage(id uint64, m *Message) error {
	return ipfsunion.MysqlDB.
		Model(&Message{Id: id}).
		Omit("id", "create_time", "params", "method").
		Updates(m).
		Error
}

func UpdateMessageState(id uint64, isSent int, msgCID string, failTimes int) error {
	return ipfsunion.MysqlDB.
		Model(&Message{Id: id}).
		Updates(map[string]interface{}{"is_sent": isSent, "msg_cid": msgCID, "fail_times": failTimes}).
		Error
}
