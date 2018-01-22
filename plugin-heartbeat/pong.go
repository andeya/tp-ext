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
		Name() string
		PostNewPeer(peer tp.EarlyPeer) error
		PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror
		PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror
	}
	heartPong struct{}
)

var (
	_ tp.PostNewPeerPlugin        = Pong(nil)
	_ tp.PostReadPullHeaderPlugin = Pong(nil)
	_ tp.PostReadPushHeaderPlugin = Pong(nil)
)

func (h *heartPong) Name() string {
	return "heart-pong"
}

func (h *heartPong) PostNewPeer(peer tp.EarlyPeer) error {
	peer.RoutePull(new(heart))
	rangeSession := peer.RangeSession
	interval := time.Second
	go func() {
		for {
			time.Sleep(interval)
			rangeSession(func(sess tp.Session) bool {
				info, ok := getHeartbeatInfo(sess)
				if !ok {
					return true
				}
				cp := info.elemCopy()
				if !sess.Health() || cp.last.Add(cp.rate*2).Before(coarsetime.CeilingTimeNow()) {
					sess.Close()
				}
				if cp.rate < interval {
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

func (h *heartPong) update(ctx tp.ReadCtx) {
	if ctx.Path() == heartbeatUri {
		return
	}
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	updateHeartbeatInfo(sess, -1)
}

const (
	heartbeatUri      = "/heart/beat"
	heartbeatQueryKey = heartbeatKey
)

type heart struct {
	tp.PullCtx
}

func (h *heart) Beat(*interface{}) (interface{}, *tp.Rerror) {
	sess := h.Session()
	tp.Tracef("%s heartbeat: pong", sess.Id())
	rateStr := h.Query().Get(heartbeatQueryKey)
	rate := getHeartbeatRate(rateStr)
	if rate == -1 {
		return nil, tp.NewRerror(tp.CodeBadPacket, "invalid heart rate", rateStr)
	}
	updateHeartbeatInfo(sess, rate)
	return nil, nil
}

func getHeartbeatRate(s string) time.Duration {
	if len(s) == 0 {
		return -1
	}
	r, err := strconv.ParseInt(s, 10, 64)
	if err != nil || r <= 0 {
		return 0
	}
	return time.Duration(r)
}
