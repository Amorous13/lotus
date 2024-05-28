package chain

import (
	"bytes"
	"fmt"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/lib/chihua/dao"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
	"time"
)

var log = logging.Logger("chihua_chain")

func InsertMessageAsync(msg *types.SignedMessage) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("InsertMessageAsync recover: %v", r)
			}
		}()

		t := time.Now()
		msgCid := msg.Cid().String()

		var methodName string
		if _, mi, err := ParseActorMessage(&msg.Message); err != nil {
			log.Error("InsertMessageAsync ParseActorMessage %v failed %v", msgCid, err)
			methodName = fmt.Sprintf("Method:%v", msg.Message.Method)
		} else {
			if mi == nil {
				log.Error("InsertMessageAsync ParseActorMessage %v mi returned nil", msgCid)
			} else {
				methodName = mi.Name
			}
		}

		params := msg.Message.Params
		if params == nil {
			params = make([]byte, 0)
		}
		extraInfo, err := messageExtraInfo(methodName, params)
		if err != nil {
			log.Errorf("InsertMessageAsync messageExtraInfo %v", err)
		}
		for i := 0; i < 3; i++ {
			m := dao.Message{
				MsgCid:     msgCid,
				Method:     methodName,
				From:       msg.Message.From.String(),
				To:         msg.Message.To.String(),
				Nonce:      msg.Message.Nonce,
				Value:      msg.Message.Value.String(),
				MaxFee:     msg.Message.GasLimit,
				ExtraInfo:  extraInfo,
				Params:     params,
				IsSent:     0,
				FailTimes:  0,
				CreateTime: t,
			}

			if err := dao.CreateMessage(&m); err != nil {
				log.Error("InsertMessageAsync CreateMessage %v failed %v", msgCid, err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			return
		}
	}()
}

// ---------------------------------------------------------------------------------------------------------------------

func messageExtraInfo(method string, params []byte) (string, error) {
	switch method {
	case "PreCommitSector":
		var p = &miner.SectorPreCommitInfo{}
		if err := p.UnmarshalCBOR(bytes.NewReader(params)); err != nil {
			return "", xerrors.Errorf("messageExtraInfo UnmarshalCBOR method %v | %v", method, err)
		}
		return fmt.Sprintf("%v", p.SectorNumber), nil

	case "ProveCommitSector":
		var p = &miner.ProveCommitSectorParams{}
		if err := p.UnmarshalCBOR(bytes.NewReader(params)); err != nil {
			return "", xerrors.Errorf("messageExtraInfo UnmarshalCBOR method %v | %v", method, err)
		}
		return fmt.Sprintf("%v", p.SectorNumber), nil
	}

	return "", xerrors.Errorf("messageExtraInfo ignore method %v", method)
}
