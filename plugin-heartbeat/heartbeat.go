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
package heartbeat

import (
	"time"

	"github.com/henrylee2cn/goutil/coarsetime"
	tp "github.com/henrylee2cn/teleport"
)

// NewPing returns a heartbeat sender plugin.
func NewPing(rate time.Duration) *heartbeat {
	return &heartbeat{
		isPing: true,
		rate:   rate,
	}
}

// NewPong returns a heartbeat receiver plugin.
func NewPong(rate time.Duration) *heartbeat {
	return &heartbeat{
		isPing: false,
		rate:   rate,
	}
}

type heartbeat struct {
	peer   tp.Peer
	isPing bool
	rate   time.Duration
}

var (
	_ tp.PostNewPeerPlugin         = new(heartbeat)
	_ tp.PostDialPlugin            = new(heartbeat)
	_ tp.PostAcceptPlugin          = new(heartbeat)
	_ tp.PostReadPullHeaderPlugin  = new(heartbeat)
	_ tp.PostReadReplyHeaderPlugin = new(heartbeat)
	_ tp.PostReadPushHeaderPlugin  = new(heartbeat)
	_ tp.PostWritePullPlugin       = new(heartbeat)
	_ tp.PostWriteReplyPlugin      = new(heartbeat)
	_ tp.PostWritePushPlugin       = new(heartbeat)
)

func (h *heartbeat) Name() string {
	return "heartbeat"
}

func (h *heartbeat) PostNewPeer(peer tp.EarlyPeer) error {
	h.peer = peer.(tp.Peer)
	return nil
}

func (h *heartbeat) PostListen() error {
	if h.isPing {
		rangeSession := h.peer.RangeSession
		go func() {
			for {
				time.Sleep(h.rate)
				rangeSession(func(sess tp.Session) bool {
					if !sess.Health() || h.tryPull(sess) != nil {
						sess.Close()
					}
					return true
				})
			}
		}()
	} else {
		h.peer.RoutePull(new(heart))
		rangeSession := h.peer.RangeSession
		go func() {
			for {
				time.Sleep(h.rate)
				rangeSession(func(sess tp.Session) bool {
					if !sess.Health() || h.isExpired(sess) {
						sess.Close()
					}
					return true
				})
			}
		}()
	}
	return nil
}

func (h *heartbeat) PostDial(sess tp.EarlySession) *tp.Rerror {
	return h.PostAccept(sess)
}

func (h *heartbeat) PostAccept(sess tp.EarlySession) *tp.Rerror {
	t := coarsetime.CeilingTimeNow()
	sess.Public().Store(heartbeatKey, &t)
	return nil
}

func (h *heartbeat) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartbeat) PostReadReplyHeader(ctx tp.ReadCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartbeat) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartbeat) PostWritePull(ctx tp.WriteCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartbeat) PostWriteReply(ctx tp.WriteCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

func (h *heartbeat) PostWritePush(ctx tp.WriteCtx) *tp.Rerror {
	h.update(ctx)
	return nil
}

const heartbeatKey = "_HB_"

func (h *heartbeat) tryPull(sess tp.Session) *tp.Rerror {
	t, ok := sess.Public().Load(heartbeatKey)
	if !ok || t.(*time.Time).Add(h.rate).After(coarsetime.CeilingTimeNow()) {
		return nil
	}
	rerr := sess.Pull("/heart/beat", nil, nil).Rerror()
	if rerr == nil {
		tp.Tracef("%s heartbeat: ping", sess.Id())
	}
	return rerr

}

func (h *heartbeat) isExpired(sess tp.Session) bool {
	t, ok := sess.Public().Load(heartbeatKey)
	if !ok {
		return false
	}
	return t.(*time.Time).Add(h.rate * 3).Before(coarsetime.CeilingTimeNow())
}

func (h *heartbeat) update(ctx tp.PreCtx) {
	sess := ctx.Session()
	if !sess.Health() {
		return
	}
	t := coarsetime.CeilingTimeNow()
	sess.Public().Store(heartbeatKey, &t)
}

type heart struct {
	tp.PullCtx
}

func (h *heart) Beat(*interface{}) (interface{}, *tp.Rerror) {
	tp.Tracef("%s heartbeat: pong", h.Session().Id())
	return nil, nil
}
