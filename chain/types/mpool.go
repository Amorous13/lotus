package types

import (
	"github.com/filecoin-project/go-state-types/abi"
	"time"

	"github.com/filecoin-project/go-address"
)

type MpoolConfig struct {
	PriorityAddrs             []address.Address
	SizeLimitHigh             int
	SizeLimitLow              int
	ReplaceByFeeRatio         Percent
	PruneCooldown             time.Duration
	GasLimitOverestimation    float64
	IsOnlyWhiteList           bool              //chihua add
	ForcePriPacking           bool              //chihua add
	MaxFeeCap                 abi.TokenAmount   //chihua add
	FeeCapOverRatio           float64           //chihua add
	GasLimitOverestimationPre float64           //chihua add
	GasLimitOverestimationPro float64           //chihua add
	NegFeeAllowedAddrs        []address.Address //chihua add
	GasLimitWhiteList         bool              //chihua add
	GasLimitRatio             float64           //chihua add
}

/*chihua begin*/
const (
	GasLimitOverestimation    string = `GasLimitOverestimation`
	GasLimitOverestimationPre string = `GasLimitOverestimationPre`
	GasLimitOverestimationPro string = `GasLimitOverestimationPro`
)

/*chihua end*/

func (mc *MpoolConfig) Clone() *MpoolConfig {
	r := new(MpoolConfig)
	*r = *mc
	return r
}
