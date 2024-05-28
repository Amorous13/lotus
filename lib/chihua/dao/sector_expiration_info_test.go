package dao

import (
	"context"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"testing"
	"time"
)

func TestSectorExpirationInfo_Create(t *testing.T) {
	value := SectorExpirationInfo{
		MinerId:         "t01111",
		SectorId:        1,
		ExpirationEpoch: 123456,
		SealRandEpoch:   123456,
		CreateTime:      time.Now().Truncate(time.Second),
	}
	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}
	defer ipfsunion.CloseDB()

	err = value.Create(context.Background(), &value)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorExpirationInfo_UpdateSealRandEpochAndExpirationEpoch(t *testing.T) {
	value := SectorExpirationInfo{
		SectorId: 1,
		MinerId:  "t01111",
	}

	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}
	defer ipfsunion.CloseDB()

	err = value.UpdateSealRandEpochAndExpirationEpoch(context.Background(), value.SectorId, 123, 456)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
