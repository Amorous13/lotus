package dao

import (
	"context"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/lib/ipfsunion"
)

func TestSectorSealingInfo_Create(t *testing.T) {
	value := SectorSealingInfo{
		SectorId:      1103,
		MinerId:       "t01101",
		State:         "P1",
		StartTime:     time.Now().Truncate(time.Second),
		WorkerAddress: "172.18.2.2",
	}

	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}

	err = value.Create(context.Background(), &value)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
