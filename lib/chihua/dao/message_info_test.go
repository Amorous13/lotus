package dao

import (
	"context"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"log"
	"testing"
	"time"
)

func TestMessageInfo_Update(t *testing.T) {
	if err := ipfsunion.SetDB("ipfsunion",
		"ipfsunion@123456",
		"172.18.22.2",
		"starpool_mining_wenyi"); err != nil {
		log.Fatalf("ipfsunion::SetDB error, err:[%v]", err)
	}
	defer ipfsunion.CloseDB()

	mi := &MessageInfo{
		SectorId:        22,
		MessageType:     2,
		From:            "FROM",
		To:              "TO",
		Method:          7,
		Value:           "VALUE",
		MaxFee:          "MAXFEE",
		Params:          "PARAMS",
		ExpirationEpoch: 1212121,
		WorkerAddress:   "WORKERADDRESS",
		MsgCid:          "MSGCID",
		State:           MessageStateNotSend,
		CreateTime:      time.Now(),
		ErrorInfo:       "",
	}

	if _, err := MessageInfoTable.Update(context.TODO(), 22, 2, mi); err != nil {
		log.Fatalf("%v", err)
	}
}
