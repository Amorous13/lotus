package dao

import (
	"context"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"testing"
	"time"
)

func TestSectorInfo_AddOrUpdateProveFetchInfo_Success(t *testing.T) {
	var value = SectorProveFetchInfo{
		MinerId:     "t01113",
		LastProveId: 100,
		MinerType:   "test",
		IpAddress:   "test ip address",
		CreateTime:  time.Now().Truncate(time.Second),
	}
	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}

	err = value.AddOrUpdateProveFetchInfo(context.Background(), value.LastProveId, value.MinerId, value.MinerType, value.IpAddress)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_GetProveFetchInfo_Success(t *testing.T) {
	var value = SectorProveFetchInfo{
		MinerId:     "t01113",
		LastProveId: 1,
		MinerType:   "test",
		IpAddress:   "test ip address",
		CreateTime:  time.Now().Truncate(time.Second),
	}
	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}

	res, err := value.GetProveFetchInfo(context.Background(), value.MinerId, value.MinerType)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
	t.Log("success")
}
