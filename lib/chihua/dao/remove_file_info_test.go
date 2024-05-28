package dao

import (
	"context"
	"testing"
)

func TestRemoveFileInfo_Update(t *testing.T) {
	initDB()

	_, bUpdate, errUpsert := (&RemoveFileInfo{}).Upsert(
		context.TODO(), &RemoveFileInfo{
			SectorId:      uint64(10),
			MinerId:       "t01000",
			WorkerAddress: "ShaoQianMac.local",
			Cache:         "/Users/mac/seal/cache/s-t01000-10",
			State:         SectorRemoveStatusNot,
		})
	if errUpsert != nil {
		t.Errorf("sealing::handleCommitting, Upsert RemoveFileInfo error, minerId:[%v], sectorId:[%v], err:[%v]", "t01000", 10, errUpsert)
	}
	t.Logf("sealing::handleCommitting, Upsert RemoveFileInfo minerId:[%v], sectorId:[%v], updated: [%v]", "t01000", 10, bUpdate)
}
