package vm

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/ipfs/go-cid"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/metrics"
)

const (
	// DefaultAvailableExecutionLanes is the number of available execution lanes; it is the bound of
	// concurrent active executions.
	// This is the default value in filecoin-ffi
	DefaultAvailableExecutionLanes = 4
	// DefaultPriorityExecutionLanes is the number of reserved execution lanes for priority computations.
	// This is purely userspace, but we believe it is a reasonable default, even with more available
	// lanes.
	DefaultPriorityExecutionLanes = 2
)

var ErrExecutorDone = errors.New("executor has been released")

// the execution environment; see below for definition, methods, and initilization
var execution *executionEnv

// implementation of vm executor with simple sanity check preventing use after free.
type vmExecutor struct {
	lk    sync.RWMutex
	vmi   Interface
	token *executionToken
	done  bool
}

var _ Executor = (*vmExecutor)(nil)

func newVMExecutor(vmi Interface, token *executionToken) Executor {
	return &vmExecutor{vmi: vmi, token: token}
}

func (e *vmExecutor) ApplyMessage(ctx context.Context, cmsg types.ChainMsg) (*ApplyRet, error) {
	e.lk.RLock()
	defer e.lk.RUnlock()

	if e.done {
		return nil, ErrExecutorDone
	}

	return e.vmi.ApplyMessage(ctx, cmsg)
}

func (e *vmExecutor) ApplyImplicitMessage(ctx context.Context, msg *types.Message) (*ApplyRet, error) {
	e.lk.RLock()
	defer e.lk.RUnlock()

	if e.done {
		return nil, ErrExecutorDone
	}

	return e.vmi.ApplyImplicitMessage(ctx, msg)
}

func (e *vmExecutor) Flush(ctx context.Context) (cid.Cid, error) {
	e.lk.RLock()
	defer e.lk.RUnlock()

	if e.done {
		return cid.Undef, ErrExecutorDone
	}

	return e.vmi.Flush(ctx)
}

func (e *vmExecutor) Done() {
	e.lk.Lock()
	defer e.lk.Unlock()

	if !e.done {
		e.token.Done()
		e.token = nil
		e.done = true
	}
}

type executionToken struct {
	lane     ExecutionLane
	reserved int
}

func (token *executionToken) Done() {
	execution.putToken(token)
}

type executionEnv struct {
	mx   *sync.Mutex
	cond *sync.Cond

	// available executors
	available int
	// reserved executors
	reserved int
}

func (e *executionEnv) getToken(lane ExecutionLane) *executionToken {
	metricsUp(metrics.VMExecutionWaiting, lane)
	defer metricsDown(metrics.VMExecutionWaiting, lane)

	e.mx.Lock()
	defer e.mx.Unlock()

	switch lane {
	case ExecutionLaneDefault:
		for e.available <= e.reserved {
			e.cond.Wait()
		}

		e.available--

		metricsUp(metrics.VMExecutionRunning, lane)
		return &executionToken{lane: lane, reserved: 0}

	case ExecutionLanePriority:
		for e.available == 0 {
			e.cond.Wait()
		}

		e.available--

		reserving := 0
		if e.reserved > 0 {
			e.reserved--
			reserving = 1
		}

		metricsUp(metrics.VMExecutionRunning, lane)
		return &executionToken{lane: lane, reserved: reserving}

	default:
		// already checked at interface boundary in NewVM, so this is appropriate
		panic("bogus execution lane")
	}
}

func (e *executionEnv) putToken(token *executionToken) {
	e.mx.Lock()
	defer e.mx.Unlock()

	e.available++
	e.reserved += token.reserved

	e.cond.Broadcast()

	metricsDown(metrics.VMExecutionRunning, token.lane)
}

func metricsUp(metric *stats.Int64Measure, lane ExecutionLane) {
	metricsAdjust(metric, lane, 1)
}

func metricsDown(metric *stats.Int64Measure, lane ExecutionLane) {
	metricsAdjust(metric, lane, -1)
}

func metricsAdjust(metric *stats.Int64Measure, lane ExecutionLane, delta int) {
	laneName := "default"
	if lane > ExecutionLaneDefault {
		laneName = "priority"
	}

	ctx, _ := tag.New(
		context.Background(),
		tag.Upsert(metrics.ExecutionLane, laneName),
	)
	stats.Record(ctx, metric.M(int64(delta)))
}

func init() {
	var available, priority int
	var err error

	concurrency := os.Getenv("LOTUS_FVM_CONCURRENCY")
	if concurrency == "" {
		available = DefaultAvailableExecutionLanes
	} else {
		available, err = strconv.Atoi(concurrency)
		if err != nil {
			panic(err)
		}
	}

	reserved := os.Getenv("LOTUS_FVM_CONCURRENCY_RESERVED")
	if reserved == "" {
		priority = DefaultPriorityExecutionLanes
	} else {
		priority, err = strconv.Atoi(reserved)
		if err != nil {
			panic(err)
		}
	}

	// some sanity checks
	if available < 2 {
		panic("insufficient execution concurrency")
	}

	if available <= priority {
		panic("insufficient default execution concurrency")
	}

	mx := &sync.Mutex{}
	cond := sync.NewCond(mx)

	execution = &executionEnv{
		mx:        mx,
		cond:      cond,
		available: available,
		reserved:  priority,
	}
}