package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"

	"go.opencensus.io/trace"
)

const LookbackCap = time.Hour

var (
	ErrLookbackTooLong = fmt.Errorf("lookbacks of more than %s are disallowed", LookbackCap)
)

type GatewayAPI struct {
	api api.FullNode

	hcLk                sync.Mutex
	headCache           *types.TipSet
	headCacheLastUpdate int64
	updating            bool
}

func (a *GatewayAPI) getTipsetTimestamp(ctx context.Context, tsk types.TipSetKey) (time.Time, error) {
	if tsk.IsEmpty() {
		return time.Now(), nil
	}

	ts, err := a.api.ChainGetTipSet(ctx, tsk)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(ts.Blocks()[0].Timestamp), 0), nil
}

func (a *GatewayAPI) checkTipset(ctx context.Context, ts types.TipSetKey) error {
	when, err := a.getTipsetTimestamp(ctx, ts)
	if err != nil {
		return err
	}

	if time.Since(when) > time.Hour {
		return ErrLookbackTooLong
	}

	return nil
}

func (a *GatewayAPI) StateGetActor(ctx context.Context, actor address.Address, ts types.TipSetKey) (*types.Actor, error) {
	ctx, span := trace.StartSpan(ctx, "StateGetActor")
	defer span.End()

	if err := a.checkTipset(ctx, ts); err != nil {
		return nil, fmt.Errorf("bad tipset: %w", err)
	}

	return a.api.StateGetActor(ctx, actor, ts)
}

func (a *GatewayAPI) ChainHead(ctx context.Context) (*types.TipSet, error) {
	ctx, span := trace.StartSpan(ctx, "ChainHead")
	defer span.End()

	now := time.Now().Unix()

	a.hcLk.Lock()
	if a.updating || now-a.headCacheLastUpdate > 1 || now-a.headCache.Timestamp() > 30 {
		ts := a.headCache
		a.hcLk.Unlock()
		return ts, nil
	}
	a.updating = true
	a.hcLk.Unlock()

	span.AddAttributes(trace.BoolAttribute("uncached", true))

	h, err := a.api.ChainHead(ctx)
	if err != nil {
		log.Warnf("chain head call failed: %w", err)
		return nil, err
	}

	a.hcLk.Lock()
	a.headCacheLastUpdate = now
	a.headCache = h
	a.updating = false
	a.hcLk.Unlock()

	return h, nil
}

func (a *GatewayAPI) ChainGetTipSet(ctx context.Context, tsk types.TipSetKey) (*types.TipSet, error) {
	ctx, span := trace.StartSpan(ctx, "ChainGetTipSet")
	defer span.End()

	if err := a.checkTipset(ctx, tsk); err != nil {
		return nil, fmt.Errorf("bad tipset: %w", err)
	}

	// TODO: since we're limiting lookbacks, should just cache this (could really even cache the json response bytes)
	return a.api.ChainGetTipSet(ctx, tsk)
}

func (a *GatewayAPI) MpoolPush(ctx context.Context, sm *types.SignedMessage) (cid.Cid, error) {
	ctx, span := trace.StartSpan(ctx, "MpoolPush")
	defer span.End()

	// TODO: additional anti-spam checks

	return a.api.MpoolPush(ctx, sm)
}
