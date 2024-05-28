package dao

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"github.com/ipfs/go-cid"
	"github.com/jinzhu/gorm"
	mh "github.com/multiformats/go-multihash"

	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
)

func makeCID(s string) cid.Cid {
	h1, err := mh.Sum([]byte(s), mh.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}
	}
	return cid.NewCidV1(0x55, h1)
}

func initDB() error {
	var err error
	ipfsunion.MysqlDB, err = gorm.Open("mysql", "ipfsunion:ipfsunion@123456@(172.18.22.2:3306)/starpool_mining_test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return err
	}
	return nil
}

func TestParallelInsertAndGet(t *testing.T) {
	var wg sync.WaitGroup

	initDB()

	params := &miner.SectorPreCommitInfo{
		Expiration:    1865,
		SectorNumber:  100,
		SealedCID:     makeCID("test"),
		SealRandEpoch: 73,
		DealIDs:       []abi.DealID{10, 25},
	}

	enc := new(bytes.Buffer)
	if err := params.MarshalCBOR(enc); err != nil {
		t.Fatalf("MarshalCBOR error: %v", err)
	}

	fmt.Printf("params: %v\n", enc.Bytes())

	var param1 miner.SectorPreCommitInfo
	if err := param1.UnmarshalCBOR(bytes.NewReader(enc.Bytes())); err != nil {
		t.Fatalf("UnmarshalCBOR Err: %v", err)
	}

	fmt.Printf("params1: %v\n", param1)

	go func() {
		for {
			msgs, err := GetMessagesToPush(50)
			if err != nil {
				t.Fatalf("GetMessagesToPush Err: %v", err)
			}

			for _, msg := range msgs {
				var param miner.SectorPreCommitInfo

				if msg.Method == "" {
					continue
				}

				isEq := bytes.Equal(enc.Bytes(), msg.Params)
				fmt.Printf("params: %v， %v\n", isEq, msg.Params)

				if err := param.UnmarshalCBOR(bytes.NewReader(msg.Params)); err != nil {
					t.Fatalf("UnmarshalCBOR Err: %v", err)
				}

				isSent := msg.IsSent
				msgCid := msg.MsgCid
				failTimes := msg.FailTimes

				failTimes++
				if failTimes >= 10 {
					msgCid = "cidtimeout"
					isSent = -1
				}

				UpdateMessageState(msg.Id, isSent, msgCid, failTimes)

				continue
			}
		}
	}()

	for i := 0; i < 2; i++ {
		go func() {
			wg.Add(1)
			msg := &Message{
				CreateTime: time.Now(),
				Method:     "ProveCommitSector",
				Params:     enc.Bytes(),
			}

			id, err := InsertMessage(msg)
			if err != nil {
				t.Fatalf("InsertMessage error: %v", err)
			}

			for {
				msgW, err := GetMessageWithCIDToWait(id)
				if err != nil {
					t.Fatalf("GetMessageWithCIDToWait error: %v", err)
				}

				//fmt.Printf("sent msg: cid=%s\n", msgW.MsgCid)

				if msgW.MsgCid == "cidtimeout" {
					fmt.Printf("sent msg: cid=%s\n", msgW.MsgCid)
					break
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()

	select {}
}
