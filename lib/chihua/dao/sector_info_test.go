package dao

import (
	"context"
	"github.com/filecoin-project/lotus/lib/ipfsunion"
	"reflect"
	"testing"
	"time"
)

var (
	UserName = "ipfsunion"
	Password = "ipfsunion@123456"
	Address  = "172.18.22.2"
	DBName   = "starpool_mining_test"
)

func TestSectorInfo_TableName(t *testing.T) {
	err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.17.201.33", "starpool_mining")
	if err != nil {
		t.Fatal(err)
	}
	err = ipfsunion.MysqlDB.Table("t_sector_dispatcher").CreateTable(&SectorInfo{}).Error
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_UpdateSectorForP1Success_Success(t *testing.T) {
	value := SectorInfo{
		SectorId:      1,
		MinerId:       "t01101",
		PreCommit1Out: "p1Out test",
		TicketValue:   "ticketValue test",
		TicketEpoch:   1000,
		CreateTime:    time.Now().Truncate(time.Second),
	}

	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}

	err = value.UpdateSectorForP1Success(context.Background(), value.SectorId, value.MinerId, value.PreCommit1Out, value.TicketValue, value.TicketEpoch)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_UpdateSectorForP1Success_WrongSectorId(t *testing.T) {
	value := SectorInfo{
		SectorId:      1000,
		MinerId:       "t01101",
		PreCommit1Out: "p1Out test",
		TicketValue:   "ticketValue test",
		TicketEpoch:   1000,
		CreateTime:    time.Now().Truncate(time.Second),
	}

	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}

	err = value.UpdateSectorForP1Success(context.Background(), value.SectorId, value.MinerId, value.PreCommit1Out, value.TicketValue, value.TicketEpoch)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_UpdateSectorForP1Success_NullPreCommit1Out(t *testing.T) {
	value := SectorInfo{
		SectorId:    1,
		MinerId:     "t01101",
		TicketValue: "ticketValue test",
		TicketEpoch: 1000,
		CreateTime:  time.Now().Truncate(time.Second),
	}

	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}

	err = value.UpdateSectorForP1Success(context.Background(), value.SectorId, value.MinerId, value.PreCommit1Out, value.TicketValue, value.TicketEpoch)
	if err != nil {
		t.Log("null PreCommit1Out")
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_GetBySectorId(t *testing.T) {
	//err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//var v SectorInfo
	//r, err := v.GetBySectorId(context.Background(), 1105, 1101)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("%+v", *r)
}

func TestSectorInfo_UpdateSectorForC1Success_Success(t *testing.T) {
	value := SectorInfo{
		SectorId:   6,
		MinerId:    "1101",
		Commit1Out: "c1Out test",
	}

	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}

	err = value.UpdateSectorForC1Success(context.Background(), value.SectorId, value.MinerId, value.Commit1Out)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestSectorInfo_GetOneC2TodoTask_Success(t *testing.T) {
	//err := ipfsunion.SetDB("ipfsunion", "ipfsunion@123456", "172.18.22.2", "starpool_mining")
	//if err != nil {
	value := SectorInfo{
		MinerId: "1101",
	}

	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}

	res, err := value.GetOneC2TodoTask(context.Background(), value.MinerId)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
	t.Log("success")
}

func TestSectorInfo_GetOneC2TodoTask(t *testing.T) {
	type fields struct {
		Id                  uint64
		MinerId             string
		SectorId            uint64
		State               string
		MinerAddress        string
		StorageId           string
		CommD               string
		CommR               string
		TicketValue         string
		TicketEpoch         int64
		SeedValue           string
		SeedEpoch           int64
		Proof               string
		RetryTimes          uint64
		Deals               string
		SectorType          int64
		PreCommit1Out       string
		PreCommit1ErrorInfo string
		PreCommit2ErrorInfo string
		Commit1ErrorInfo    string
		Commit2ErrorInfo    string
		SeedErrorInfo       string
		PreCommitMsgCid     string
		PreCommit2FailTimes int
		CommP               string
		CommitMsgCid        string
		Commit1Out          string
		C2DispatcherStatus  C2DispatcherStatusType
		Commit2FailTimes    int
		FaultReportMsgCid   string
		CreateTime          time.Time
	}
	type args struct {
		ctx     context.Context
		minerId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SectorInfo
		wantErr bool
	}{
		{name: "1", fields: fields{}, args: args{context.Background(), "t01011"}, wantErr: false},
	}
	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &SectorInfo{
				SectorId:     tt.fields.SectorId,
				MinerId:      tt.fields.MinerId,
				State:        tt.fields.State,
				MinerAddress: tt.fields.MinerAddress,
				StorageId:    tt.fields.StorageId,
				CommD:        tt.fields.CommD,
				CommR:        tt.fields.CommR,
				TicketValue:  tt.fields.TicketValue,
				TicketEpoch:  tt.fields.TicketEpoch,
				SeedValue:    tt.fields.SeedValue,
				SeedEpoch:    tt.fields.SeedEpoch,
				Proof:        tt.fields.Proof,
				//RetryTimes:          tt.fields.RetryTimes,
				//Deals:               tt.fields.Deals,
				//SectorType:          tt.fields.SectorType,
				PreCommit1Out:       tt.fields.PreCommit1Out,
				PreCommit1ErrorInfo: tt.fields.PreCommit1ErrorInfo,
				PreCommit2ErrorInfo: tt.fields.PreCommit2ErrorInfo,
				Commit1ErrorInfo:    tt.fields.Commit1ErrorInfo,
				Commit2ErrorInfo:    tt.fields.Commit2ErrorInfo,
				//SeedErrorInfo:       tt.fields.SeedErrorInfo,
				PreCommitMsgCid: tt.fields.PreCommitMsgCid,
				//PreCommit2FailTimes: tt.fields.PreCommit2FailTimes,
				CommP:              tt.fields.CommP,
				CommitMsgCid:       tt.fields.CommitMsgCid,
				Commit1Out:         tt.fields.Commit1Out,
				C2DispatcherStatus: tt.fields.C2DispatcherStatus,
				//Commit2FailTimes:    tt.fields.Commit2FailTimes,
				//FaultReportMsgCid:   tt.fields.FaultReportMsgCid,
				CreateTime: tt.fields.CreateTime,
			}
			got, err := this.GetOneC2TodoTask(tt.args.ctx, tt.args.minerId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOneC2TodoTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOneC2TodoTask() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSectorInfo_GetOneC2TodoTask_2(t *testing.T) {
	type fields struct {
		Id                  uint64
		MinerId             string
		SectorId            uint64
		State               string
		MinerAddress        string
		StorageId           string
		CommD               string
		CommR               string
		TicketValue         string
		TicketEpoch         int64
		SeedValue           string
		SeedEpoch           int64
		Proof               string
		RetryTimes          uint64
		Deals               string
		SectorType          int64
		PreCommit1Out       string
		PreCommit1ErrorInfo string
		PreCommit2ErrorInfo string
		Commit1ErrorInfo    string
		Commit2ErrorInfo    string
		SeedErrorInfo       string
		PreCommitMsgCid     string
		PreCommit2FailTimes int
		CommP               string
		CommitMsgCid        string
		Commit1Out          string
		C2DispatcherStatus  C2DispatcherStatusType
		Commit2FailTimes    int
		FaultReportMsgCid   string
		CreateTime          time.Time
	}
	type args struct {
		ctx     context.Context
		minerId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SectorInfo
		wantErr bool
	}{
		{name: "1", fields: fields{}, args: args{context.Background(), "t01011"}, wantErr: false},
	}
	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &SectorInfo{
				SectorId:     tt.fields.SectorId,
				MinerId:      tt.fields.MinerId,
				State:        tt.fields.State,
				MinerAddress: tt.fields.MinerAddress,
				StorageId:    tt.fields.StorageId,
				CommD:        tt.fields.CommD,
				CommR:        tt.fields.CommR,
				TicketValue:  tt.fields.TicketValue,
				TicketEpoch:  tt.fields.TicketEpoch,
				SeedValue:    tt.fields.SeedValue,
				SeedEpoch:    tt.fields.SeedEpoch,
				Proof:        tt.fields.Proof,
				//RetryTimes:          tt.fields.RetryTimes,
				//Deals:               tt.fields.Deals,
				//SectorType:          tt.fields.SectorType,
				PreCommit1Out:       tt.fields.PreCommit1Out,
				PreCommit1ErrorInfo: tt.fields.PreCommit1ErrorInfo,
				PreCommit2ErrorInfo: tt.fields.PreCommit2ErrorInfo,
				Commit1ErrorInfo:    tt.fields.Commit1ErrorInfo,
				Commit2ErrorInfo:    tt.fields.Commit2ErrorInfo,
				//SeedErrorInfo:       tt.fields.SeedErrorInfo,
				PreCommitMsgCid: tt.fields.PreCommitMsgCid,
				//PreCommit2FailTimes: tt.fields.PreCommit2FailTimes,
				CommP:              tt.fields.CommP,
				CommitMsgCid:       tt.fields.CommitMsgCid,
				Commit1Out:         tt.fields.Commit1Out,
				C2DispatcherStatus: tt.fields.C2DispatcherStatus,
				//Commit2FailTimes:    tt.fields.Commit2FailTimes,
				//FaultReportMsgCid:   tt.fields.FaultReportMsgCid,
				CreateTime: tt.fields.CreateTime,
			}
			_, err := this.GetOneC2TodoTask(tt.args.ctx, tt.args.minerId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOneC2TodoTask_2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(err)
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetOneC2TodoTask_2() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestSectorInfo_CreateSector(t *testing.T) {
	type fields struct {
		SectorId            uint64
		MinerId             string
		State               string
		MinerAddress        string
		StorageId           string
		CommD               string
		CommR               string
		TicketValue         string
		TicketEpoch         int64
		SeedValue           string
		SeedEpoch           int64
		Proof               string
		RetryTimes          uint64
		Deals               string
		SectorType          int64
		PreCommit1Out       string
		PreCommit1ErrorInfo string
		PreCommit2ErrorInfo string
		Commit1ErrorInfo    string
		Commit2ErrorInfo    string
		SeedErrorInfo       string
		PreCommitMsgCid     string
		PreCommit2FailTimes int
		CommP               string
		CommitMsgCid        string
		Commit1Out          string
		C2DispatcherStatus  C2DispatcherStatusType
		Commit2FailTimes    int
		FaultReportMsgCid   string
		CreateTime          time.Time
	}
	type args struct {
		ctx  context.Context
		info *SectorInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SectorInfo
		wantErr bool
	}{
		{name: "1", fields: fields{}, args: args{context.Background(), &SectorInfo{MinerId: "t01011", C2DispatcherStatus: C2DispatcherStatusInit}}, wantErr: false},
		{name: "2", fields: fields{}, args: args{context.Background(), &SectorInfo{MinerId: "t01011", C2DispatcherStatus: C2DispatcherStatusInit}}, wantErr: false},
		{name: "3", fields: fields{}, args: args{context.Background(), &SectorInfo{MinerId: "t01011", C2DispatcherStatus: C2DispatcherStatusInit}}, wantErr: false},
		{name: "4", fields: fields{}, args: args{context.Background(), &SectorInfo{MinerId: "t01011", C2DispatcherStatus: C2DispatcherStatusInit}}, wantErr: false},
		{name: "5", fields: fields{}, args: args{context.Background(), &SectorInfo{MinerId: "t01011", C2DispatcherStatus: C2DispatcherStatusInit}}, wantErr: false},
	}
	err := ipfsunion.SetDB(UserName, Password, Address, DBName)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &SectorInfo{
				SectorId:     tt.fields.SectorId,
				MinerId:      tt.fields.MinerId,
				State:        tt.fields.State,
				MinerAddress: tt.fields.MinerAddress,
				StorageId:    tt.fields.StorageId,
				CommD:        tt.fields.CommD,
				CommR:        tt.fields.CommR,
				TicketValue:  tt.fields.TicketValue,
				TicketEpoch:  tt.fields.TicketEpoch,
				SeedValue:    tt.fields.SeedValue,
				SeedEpoch:    tt.fields.SeedEpoch,
				Proof:        tt.fields.Proof,
				//RetryTimes:          tt.fields.RetryTimes,
				//Deals:               tt.fields.Deals,
				//SectorType:          tt.fields.SectorType,
				PreCommit1Out:       tt.fields.PreCommit1Out,
				PreCommit1ErrorInfo: tt.fields.PreCommit1ErrorInfo,
				PreCommit2ErrorInfo: tt.fields.PreCommit2ErrorInfo,
				Commit1ErrorInfo:    tt.fields.Commit1ErrorInfo,
				Commit2ErrorInfo:    tt.fields.Commit2ErrorInfo,
				//SeedErrorInfo:       tt.fields.SeedErrorInfo,
				PreCommitMsgCid: tt.fields.PreCommitMsgCid,
				//PreCommit2FailTimes: tt.fields.PreCommit2FailTimes,
				CommP:              tt.fields.CommP,
				CommitMsgCid:       tt.fields.CommitMsgCid,
				Commit1Out:         tt.fields.Commit1Out,
				C2DispatcherStatus: tt.fields.C2DispatcherStatus,
				//Commit2FailTimes:    tt.fields.Commit2FailTimes,
				//FaultReportMsgCid:   tt.fields.FaultReportMsgCid,
				CreateTime: tt.fields.CreateTime,
			}
			got, err := this.CreateSector(tt.args.ctx, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.SectorId == 0 {
				t.Errorf("CreateSector()  error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Log(got)
		})
	}
}
