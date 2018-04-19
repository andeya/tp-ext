// Copyright 2018 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package heartbeat

import (
	"strconv"
	"time"

	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport"
)

// NewPong returns a heartbeat receiver plugin.
func NewPong() Pong {
	return new(heartPong)
}

type (
	// Pong receive heartbeat.
	Pong interface {
		// Name returns name.
		Name() string
		// PostNewPeer runs ping woker.
		PostNewPeer(peer tp.EarlyPeer) error
		// PostWritePull updates heartbeat information.
		PostWritePull(ctx tp.WriteCtx) *tp.Rerror
		// PostWritePush updates heartbeat information.
		PostWritePush(ctx tp.WriteCtx) *tp.Rerror
		// PostReadPullHeader updates heartbeat information.
		PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror
		// PostReadPushHeader updates heartbeat information.
		PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror
	}
	heartPong struct{}
)

var (
	_ tp.PostNewPeerPlugin        = Pong(nil)
	_ tp.PostWritePullPlugin      = Pong(nil)
	_ tp.PostWritePushPlugin      = Pong(nil)
	_ tp.PostReadPullHeaderPlugin = Pong(nil)
	_ tp.PostReadPushHeaderPlugin = Pong(nil)
)

func (h *heartPong) Name() string {
	return "heart-pong"
}

func (h *heartPong) PostNewPeer(peer tp.EarlyPeer) error {
	peer.RoutePullFunc((*pongPull).heartbeat)
	peer.RoutePushFunc((*pongPush).heartbeat)
	rangeSession := peer.RangeSession
	const initial = time.Second*minRateSecond - 1
	interval := initial
	go func() {
		for {
			time.Sleep(interval)
			rangeSession(func(sess tp.Session) bool {
				info, ok := getHeartbeatInfo(sess.Swap())
				if !ok {
					return true
				}
				cp := info.elemCopy()
				if !sess.Health() || cp.last.Add(cp.rate*2).Before(coarsetime.CeilingTimeNow()) {
					sess.Close()
				}
				if cp.rate < interval || interval == initial {
					interval = cp.rate
				}
				return true
			})
		}
	}()
	return nil
}

func (h *heartPong) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartPong) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartPong) PostWritePull(ctx tp.WriteCtx) *tp.Rerror {
	return h.PostWritePush(ctx)
}

func (h *heartPong) PostWritePush(ctx tp.WriteCtx) *tp.Rerror {
	sess := ctx.Session()
	if !sess.Health() {
		return nil
	}
	updateHeartbeatInfo(sess.Swap(), -1)
	return nil
}

func (h *heartPong) update(ctx tp.ReadCtx) {
	if ctx.Path() == HeartbeatUri {
		return
	}
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	updateHeartbeatInfo(sess.Swap(), -1)
}

type pongPull struct {
	tp.PullCtx
}

func (ctx *pongPull) heartbeat(_ *struct{}) (*struct{}, *tp.Rerror) {
	sess := ctx.Session()
	rateStr := ctx.Query().Get(heartbeatQueryKey)
	rateSecond := parseHeartbeatRateSecond(rateStr)
	if rateSecond == -1 {
		return nil, tp.NewRerror(tp.CodeBadPacket, "Invalid Heartbeat Rate", rateStr)
	}
	if rateSecond == 0 {
		tp.Tracef("heart-pong: %s", sess.Id())
	} else {
		tp.Tracef("heart-pong: %s, set rate: %ds", sess.Id(), rateSecond)
	}
	updateHeartbeatInfo(sess.Swap(), time.Second*time.Duration(rateSecond))
	return nil, nil
}

type pongPush struct {
	tp.PushCtx
}

func (ctx *pongPush) heartbeat(_ *struct{}) *tp.Rerror {
	sess := ctx.Session()
	rateStr := ctx.Query().Get(heartbeatQueryKey)
	rateSecond := parseHeartbeatRateSecond(rateStr)
	if rateSecond == -1 {
		tp.Debugf("invalid heart rate: %s, %s(/s)", sess.Id(), rateStr)
		return nil
	}
	if rateSecond == 0 {
		tp.Tracef("heart-pong: %s", sess.Id())
	} else {
		tp.Tracef("heart-pong: %s, set rate: %ds", sess.Id(), rateSecond)
	}
	updateHeartbeatInfo(sess.Swap(), time.Second*time.Duration(rateSecond))
	return nil
}

func parseHeartbeatRateSecond(s string) int {
	if len(s) == 0 {
		return -1
	}
	r, err := strconv.Atoi(s)
	if err != nil || r <= 0 {
		return 0
	}
	return r
}
