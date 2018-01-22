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
func NewPing(rate time.Duration) Ping {
	p := new(heartPing)
	p.SetRate(rate)
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
		// SetRate sets heartbeat rate.
		SetRate(rate time.Duration)
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
	_ tp.PostNewPeerPlugin   = Ping(nil)
	_ tp.PostDialPlugin      = Ping(nil)
	_ tp.PostAcceptPlugin    = Ping(nil)
	_ tp.PostWritePullPlugin = Ping(nil)
	_ tp.PostWritePushPlugin = Ping(nil)
)

// SetRate sets heartbeat rate.
func (h *heartPing) SetRate(rate time.Duration) {
	h.mu.Lock()
	h.pingRate = rate
	h.uri = heartbeatUri + "?" + heartbeatQueryKey + "=" + strconv.FormatInt(int64(rate), 10)
	h.mu.Unlock()
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
				if !sess.Health() || h.tryPull(sess) != nil {
					sess.Close()
				}
				return true
			})
		}
	}()
	return nil
}

func (h *heartPing) PostDial(sess tp.EarlySession) *tp.Rerror {
	initHeartbeatInfo(sess.Public(), h.getRate())
	return nil
}

func (h *heartPing) PostAccept(sess tp.EarlySession) *tp.Rerror {
	initHeartbeatInfo(sess.Public(), h.getRate())
	return nil
}

func (h *heartPing) PostWritePull(ctx tp.WriteCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartPing) PostWritePush(ctx tp.WriteCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartPing) tryPull(sess tp.Session) *tp.Rerror {
	info, ok := getHeartbeatInfo(sess)
	cp := info.elemCopy()
	if !ok || cp.last.Add(cp.rate).After(coarsetime.CeilingTimeNow()) {
		return nil
	}
	rerr := sess.Pull(h.getUri(), nil, nil).Rerror()
	if rerr == nil {
		tp.Tracef("%s heartbeat: ping", sess.Id())
	} else {
		tp.Errorf("%s heartbeat: ping fail: %s", sess.Id(), rerr.String())
	}
	return rerr
}

func (h *heartPing) update(ctx tp.PreCtx) {
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	updateHeartbeatInfo(sess, h.getRate())
}
