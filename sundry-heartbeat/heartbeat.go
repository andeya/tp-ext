// Heartbeat is a generic timing heartbeat package.
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

	tp "github.com/henrylee2cn/teleport"
)

// WithPong makes the peer a heartbeat sender.
func WithPing(peer *tp.Peer, rate time.Duration) {
	rangeSession := peer.RangeSession
	go func() {
		for {
			time.Sleep(rate)
			rangeSession(func(sess tp.Session) bool {
				if !sess.Health() || sess.Pull("/heart/beat", nil, nil).Rerror() != nil {
					sess.Close()
				}
				tp.Tracef("%s heartbeat: ping", sess.Id())
				return true
			})
		}
	}()
}

// WithPong makes the peer a heartbeat receiver.
func WithPong(peer *tp.Peer) {
	peer.PullRouter.Reg(new(heart))
}

type heart struct {
	tp.PullCtx
}

func (h *heart) Beat(*interface{}) (interface{}, *tp.Rerror) {
	tp.Tracef("%s heartbeat: pong", h.Session().Id())
	return nil, nil
}
