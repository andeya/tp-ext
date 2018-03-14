// Heartbeat is a generic timing heartbeat plugin.
//
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
//
package heartbeat

import (
	"strconv"
	"sync"
	"time"

	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport"
)

// NewPing returns a heartbeat sender plugin.
func NewPing(rateSecond int) Ping {
	p := new(heartPing)
	p.SetRate(rateSecond)
	return p
}

type (
	// Ping send heartbeat.
	Ping interface {
		Name() string
		PostNewPeer(peer tp.EarlyPeer) error
		PostDial(sess tp.EarlySession) *tp.Rerror
		PostAccept(sess tp.EarlySession) *tp.Rerror
		PostWritePull(ctx tp.WriteCtx) *tp.Rerror
		PostWritePush(ctx tp.WriteCtx) *tp.Rerror
		PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror
		PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror
		// SetRate sets heartbeat rate.
		SetRate(rateSecond int)
	}
	heartPing struct {
		peer     tp.Peer
		pingRate time.Duration
		uri      string
		mu       sync.RWMutex
		once     sync.Once
	}
)

var (
	_ tp.PostNewPeerPlugin        = Ping(nil)
	_ tp.PostDialPlugin           = Ping(nil)
	_ tp.PostAcceptPlugin         = Ping(nil)
	_ tp.PostWritePullPlugin      = Ping(nil)
	_ tp.PostWritePushPlugin      = Ping(nil)
	_ tp.PostReadPullHeaderPlugin = Ping(nil)
	_ tp.PostReadPushHeaderPlugin = Ping(nil)
)

// SetRate sets heartbeat rate.
func (h *heartPing) SetRate(rateSecond int) {
	if rateSecond < minRateSecond {
		rateSecond = minRateSecond
	}
	h.mu.Lock()
	h.pingRate = time.Second * time.Duration(rateSecond)
	h.uri = HeartbeatUri + "?" + heartbeatQueryKey + "=" + strconv.Itoa(rateSecond)
	h.mu.Unlock()
	tp.Infof("set heartbeat rate: %ds", rateSecond)
}

func (h *heartPing) getRate() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pingRate
}

func (h *heartPing) getUri() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.uri
}

func (h *heartPing) Name() string {
	return "heart-ping"
}

func (h *heartPing) PostNewPeer(peer tp.EarlyPeer) error {
	rangeSession := peer.RangeSession
	go func() {
		for {
			time.Sleep(h.getRate())
			rangeSession(func(sess tp.Session) bool {
				if !sess.Health() {
					sess.Close()
					return true
				}
				info, ok := getHeartbeatInfo(sess.Public())
				cp := info.elemCopy()
				if !ok || cp.last.Add(cp.rate).After(coarsetime.CeilingTimeNow()) {
					return true
				}
				h.goPush(sess)
				return true
			})
		}
	}()
	return nil
}

func (h *heartPing) PostDial(sess tp.EarlySession) *tp.Rerror {
	return h.PostAccept(sess)
}

func (h *heartPing) PostAccept(sess tp.EarlySession) *tp.Rerror {
	rate := h.getRate()
	initHeartbeatInfo(sess.Public(), rate)
	return nil
}

func (h *heartPing) PostWritePull(ctx tp.WriteCtx) *tp.Rerror {
	return h.PostWritePush(ctx)
}

func (h *heartPing) PostWritePush(ctx tp.WriteCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartPing) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
	return h.PostReadPushHeader(ctx)
}

func (h *heartPing) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartPing) goPush(sess tp.Session) {
	tp.Go(func() {
		if sess.Push(h.getUri(), nil) != nil {
			sess.Close()
		}
	})
}

func (h *heartPing) update(ctx tp.PreCtx) {
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	updateHeartbeatInfo(sess.Public(), h.getRate())
}
