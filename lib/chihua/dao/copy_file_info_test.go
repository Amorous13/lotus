package dao

import (
	"context"
	"fmt"
	"testing"
)

func TestCopyFileInfo_Update(t *testing.T) {
	initDB()

	var obj = CopyFileInfo{
		SectorId:      6,
		WorkerAddress: "SH-172-17-202-39-WORKKKKK",
		Sealed:        "Haha Sealed",
	}

	if err := obj.Update(context.TODO(), &obj); err != nil {
		t.Fail()
	}
	t.Log("updated")

	cfi, err := obj.Get(context.TODO(), 6, "t01000")
	if err != nil {
		t.Fail()
	}

	if cfi.Sealed != "Haha Sealed" || cfi.WorkerAddress != "SH-172-17-202-39-WORKKKKK" {
		t.Fail()
	}

	fmt.Printf("cfi is %+v", *cfi)
}
