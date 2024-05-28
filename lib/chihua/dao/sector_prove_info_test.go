package dao

import (
	"context"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"testing"
	"time"
)

func TestSectorInfo_AddProveSector_Success(t *testing.T) {
	value := SectorProveInfo{
		MinerId:    "t01111",
		SectorId:   1,
		StorageId:  "test",
		CreateTime: time.Now().Truncate(time.Second),
	}
	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}

	err = value.AddProveSector(context.Background(), &value)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_GetProveSectorList_Success(t *testing.T) {
	value := SectorProveInfo{}
	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}

	lastProveSectorId := 0
	res, err := value.GetProveSectorList(context.Background(), value.MinerId, uint64(lastProveSectorId))
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Log("len(res) == 0")
	}
	t.Log("success")
}
