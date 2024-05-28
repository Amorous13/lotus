package chain

import (
	"errors"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/cron"
	init_ "github.com/filecoin-project/specs-actors/actors/builtin/init"
	"github.com/filecoin-project/specs-actors/actors/builtin/market"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	"github.com/filecoin-project/specs-actors/actors/builtin/paych"
	"github.com/filecoin-project/specs-actors/actors/builtin/power"
	"github.com/filecoin-project/specs-actors/actors/builtin/reward"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/ipfs/go-cid"
)

var (
	null          = struct{}{}
	actorInfos    = map[cid.Cid]ActorInfo{}
	addressToCode = map[address.Address]cid.Cid{
		// TODO 这里将 SystemActorAddr 对应到了 builtin.InitActorCodeID ?

		builtin.SystemActorAddr:           builtin.SystemActorCodeID,
		builtin.StoragePowerActorAddr:     builtin.StoragePowerActorCodeID,
		builtin.StorageMarketActorAddr:    builtin.StorageMarketActorCodeID,
		builtin.VerifiedRegistryActorAddr: builtin.VerifiedRegistryActorCodeID,

		builtin.InitActorAddr:   builtin.InitActorCodeID,
		builtin.RewardActorAddr: builtin.RewardActorCodeID,
		builtin.CronActorAddr:   builtin.CronActorCodeID,
	}
)

var (
	TypeNull     = reflect.TypeOf(null)
	TypeNil      = reflect.TypeOf(nil)
	TypeActorPtr = reflect.TypeOf((*types.Actor)(nil))
	TypeVMCtx    = reflect.TypeOf(new(runtime.Runtime)).Elem()
)

var (
	ErrActorNotFound  = errors.New("cann't found actor")
	ErrMethodNotFound = errors.New("cann't found method in actor")
)

type actorInterface interface {
	Exports() []interface{}
}

func init() {
	// actorInfos[actors.AccountCodeCid] = ActorInfo{

	//actorInfos[builtin.SystemActorCodeID] = parseActor(system.Actor{}, "SystemActor",builtin.Methods)

	actorInfos[builtin.InitActorCodeID] = parseActor(init_.Actor{}, "InitActor", builtin.MethodsInit)
	actorInfos[builtin.AccountActorCodeID] = parseActor(init_.Actor{}, "AccountActor", builtin.MethodsAccount)
	actorInfos[builtin.RewardActorCodeID] = parseActor(reward.Actor{}, "RewardActor", builtin.MethodsReward)
	actorInfos[builtin.CronActorCodeID] = parseActor(cron.Actor{}, "CronActor", builtin.MethodsCron)
	actorInfos[builtin.StoragePowerActorCodeID] = parseActor(power.Actor{}, "StoragePowerActor", builtin.MethodsPower)
	actorInfos[builtin.StorageMarketActorCodeID] = parseActor(market.Actor{}, "StorageMarketActor", builtin.MethodsMarket)
	actorInfos[builtin.StorageMinerActorCodeID] = parseActor(miner.Actor{}, "StorageMinerActor", builtin.MethodsMiner)
	actorInfos[builtin.MultisigActorCodeID] = parseActor(multisig.Actor{}, "MultisigActor", builtin.MethodsMultisig)
	actorInfos[builtin.PaymentChannelActorCodeID] = parseActor(paych.Actor{}, "PaymentChannelActor", builtin.MethodsPaych)
}

func LookupByAddress(addr address.Address) (ActorInfo, bool) {
	if code, ok := addressToCode[addr]; ok {
		return Lookup(code)
	}

	return ActorInfo{}, false
}

func ParseActorMessage(message *types.Message) (*ActorInfo, *MethodInfo, error) {
	if message.Method == 0 {
		return nil, nil, ErrActorNotFound
	}
	actor, exist := LookupByAddress(message.To)
	if !exist {
		// TODO 这里是否需要？
		if actor, exist = Lookup(builtin.StorageMinerActorCodeID); !exist {
			return nil, nil, ErrActorNotFound
		}
	}

	method, exist := actor.LookupMethod(uint64(message.Method))
	if !exist {
		return nil, nil, ErrMethodNotFound
	}
	return &actor, &method, nil
}

func Lookup(code cid.Cid) (ActorInfo, bool) {
	act, ok := actorInfos[code]
	return act, ok
}

type ActorInfo struct {
	Name      string
	Methods   []MethodInfo
	methodMap map[uint64]int
}

func (a *ActorInfo) LookupMethod(num uint64) (MethodInfo, bool) {
	if idx, ok := a.methodMap[num]; ok {
		return a.Methods[idx], true
	}
	return MethodInfo{}, false
}

type MethodInfo struct {
	Name      string
	Num       uint64
	paramType reflect.Type
}

func (m *MethodInfo) NewParam() interface{} {
	if m.paramType == TypeNull {
		return nil
	}

	return reflect.New(m.paramType).Interface()
}

func parseActor(act actorInterface, actName string, methods interface{}) ActorInfo {
	methodInfos := []MethodInfo{}
	methodMap := map[uint64]int{}

	methodFuncs := act.Exports()

	mv := reflect.ValueOf(methods)
	mt := mv.Type()
	fnum := mt.NumField()

	for i := 0; i < fnum; i++ {
		mnum := mv.Field(i).Uint()
		methodMap[mnum] = len(methodInfos)

		var mtd interface{}
		if mnum < uint64(len(methodFuncs)) {
			mtd = methodFuncs[mnum]
		} else {
			log.Warnf("parseActor actName %s exported methods %v less than declared methods %v", actName, len(methodFuncs), fnum)
		}
		methodInfos = append(methodInfos, MethodInfo{
			Name:      mt.Field(i).Name,
			Num:       mnum,
			paramType: getMethodParam(mtd),
		})
	}

	return ActorInfo{
		Name:      actName, //reflect.TypeOf(act).Name(),
		Methods:   methodInfos,
		methodMap: methodMap,
	}
}

func getMethodParam(meth interface{}) reflect.Type {
	if meth == nil {
		return TypeNull
	}

	metht := reflect.TypeOf(meth)
	if metht.Kind() != reflect.Func || metht.NumIn() != 3 {
		return TypeNull
	}

	if metht.In(0) != TypeActorPtr || metht.In(1) != TypeVMCtx {
		return TypeNull
	}

	pt := metht.In(2)
	for pt.Kind() == reflect.Ptr {
		pt = pt.Elem()
	}

	if pt.Kind() != reflect.Struct {
		return TypeNull
	}

	return pt
}
