// Package encrypt encrypting the packet body.
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
package encrypt

import (
	"crypto/aes"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// NewEncryptPlugin creates a AES encryption plugin.
// The cipherkey argument should be the AES key,
// either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewEncryptPlugin(rerrCode int32, cipherkey string) tp.Plugin {
	b := []byte(cipherkey)
	if _, err := aes.NewCipher(b); err != nil {
		tp.Fatalf("NewEncryptPlugin: %v", err)
	}
	return &encryptPlugin{
		cipherkey: b,
		rerrCode:  rerrCode,
	}
}

var (
	_ tp.PreWritePullPlugin      = (*encryptPlugin)(nil)
	_ tp.PreWritePushPlugin      = (*encryptPlugin)(nil)
	_ tp.PreWriteReplyPlugin     = (*encryptPlugin)(nil)
	_ tp.PreReadPullBodyPlugin   = (*encryptPlugin)(nil)
	_ tp.PostReadPullBodyPlugin  = (*encryptPlugin)(nil)
	_ tp.PreReadReplyBodyPlugin  = (*encryptPlugin)(nil)
	_ tp.PostReadReplyBodyPlugin = (*encryptPlugin)(nil)
	_ tp.PreReadPushBodyPlugin   = (*encryptPlugin)(nil)
	_ tp.PostReadPushBodyPlugin  = (*encryptPlugin)(nil)
)

type encryptPlugin struct {
	cipherkey []byte
	rerrCode  int32
}

func (e *encryptPlugin) Name() string {
	return "encrypt"
}

func (e *encryptPlugin) PreWritePull(ctx tp.WriteCtx) *tp.Rerror {
	bodyBytes, err := ctx.Output().MarshalBody()
	if err != nil {
		return tp.NewRerror(e.rerrCode, "marshal raw body error", err.Error())
	}
	ciphertext := goutil.AESEncrypt(e.cipherkey, bodyBytes)
	ctx.Output().SetBody(&Encrypt{goutil.BytesToString(ciphertext)})
	return nil
}

func (e *encryptPlugin) PreWritePush(ctx tp.WriteCtx) *tp.Rerror {
	return e.PreWritePull(ctx)
}

func (e *encryptPlugin) PreWriteReply(ctx tp.WriteCtx) *tp.Rerror {
	return e.PreWritePull(ctx)
}

func (e *encryptPlugin) PreReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	ctx.Swap().Store("encrypt_rawbody", ctx.Input().Body())
	ctx.Input().SetBody(new(Encrypt))
	return nil
}

func (e *encryptPlugin) PostReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	ciphertext := ctx.Input().Body().(*Encrypt).GetCiphertext()
	bodyBytes, err := goutil.AESDecrypt(e.cipherkey, goutil.StringToBytes(ciphertext))
	if err != nil {
		return tp.NewRerror(e.rerrCode, "decrypt ciphertext error", err.Error())
	}
	rawbody, ok := ctx.Swap().Load("encrypt_rawbody")
	if !ok {
		return tp.NewRerror(e.rerrCode, "encrypt_rawbody is not exist!", "")
	}
	ctx.Swap().Delete("encrypt_rawbody")
	ctx.Input().SetBody(rawbody)
	err = ctx.Input().UnmarshalBody(bodyBytes)
	if err != nil {
		return tp.NewRerror(e.rerrCode, "unmarshal raw body error", err.Error())
	}
	return nil
}

func (e *encryptPlugin) PreReadReplyBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PreReadPullBody(ctx)
}

func (e *encryptPlugin) PostReadReplyBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PostReadPullBody(ctx)
}

func (e *encryptPlugin) PreReadPushBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PreReadPullBody(ctx)
}

func (e *encryptPlugin) PostReadPushBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PostReadPullBody(ctx)
}
