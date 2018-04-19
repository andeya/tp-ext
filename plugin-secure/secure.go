// Package secure encrypting/decrypting the packet body.
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
package secure

import (
	"crypto/aes"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// SECURE_QUERY_KEY if the query parameter is existed, perform encryption operation to the body.
const SECURE_QUERY_KEY = "_secure"

type swapKey string

const encrypt_rawbody swapKey = ""

// NewSecurePlugin creates a AES encryption/decryption plugin.
// The cipherkey argument should be the AES key,
// either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewSecurePlugin(rerrCode int32, cipherkey string) tp.Plugin {
	b := []byte(cipherkey)
	if _, err := aes.NewCipher(b); err != nil {
		tp.Fatalf("NewSecurePlugin: %v", err)
	}
	version := goutil.Md5([]byte(cipherkey))
	return &securePlugin{
		encryptPlugin: &encryptPlugin{
			version:   version,
			cipherkey: b,
			rerrCode:  rerrCode,
		},
		decryptPlugin: &decryptPlugin{
			version:   version,
			cipherkey: b,
			rerrCode:  rerrCode,
		},
	}
}

// NewEncryptPlugin creates a AES encryption plugin.
// The cipherkey argument should be the AES key,
// either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewEncryptPlugin(rerrCode int32, cipherkey string) tp.Plugin {
	b := []byte(cipherkey)
	if _, err := aes.NewCipher(b); err != nil {
		tp.Fatalf("NewEncryptPlugin: %v", err)
	}
	version := goutil.Md5([]byte(cipherkey))
	return &encryptPlugin{
		version:   version,
		cipherkey: b,
		rerrCode:  rerrCode,
	}
}

// NewDecryptPlugin creates a AES decryption plugin.
// The cipherkey argument should be the AES key,
// either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewDecryptPlugin(rerrCode int32, cipherkey string) tp.Plugin {
	b := []byte(cipherkey)
	if _, err := aes.NewCipher(b); err != nil {
		tp.Fatalf("NewDecryptPlugin: %v", err)
	}
	version := goutil.Md5([]byte(cipherkey))
	return &decryptPlugin{
		version:   version,
		cipherkey: b,
		rerrCode:  rerrCode,
	}
}

var (
	_ tp.PreWritePullPlugin      = (*encryptPlugin)(nil)
	_ tp.PreWritePushPlugin      = (*encryptPlugin)(nil)
	_ tp.PreWriteReplyPlugin     = (*encryptPlugin)(nil)
	_ tp.PreReadPullBodyPlugin   = (*decryptPlugin)(nil)
	_ tp.PostReadPullBodyPlugin  = (*decryptPlugin)(nil)
	_ tp.PreReadReplyBodyPlugin  = (*decryptPlugin)(nil)
	_ tp.PostReadReplyBodyPlugin = (*decryptPlugin)(nil)
	_ tp.PreReadPushBodyPlugin   = (*decryptPlugin)(nil)
	_ tp.PostReadPushBodyPlugin  = (*decryptPlugin)(nil)
)

type (
	securePlugin struct {
		*encryptPlugin
		*decryptPlugin
	}
	encryptPlugin struct {
		version   string
		cipherkey []byte
		rerrCode  int32
	}
	decryptPlugin encryptPlugin
)

func (e *securePlugin) Name() string {
	return "secure(encrypt&decrypt)"
}

func (e *decryptPlugin) Name() string {
	return "decrypt"
}

func (e *encryptPlugin) Name() string {
	return "encrypt"
}

func (e *encryptPlugin) PreWritePull(ctx tp.WriteCtx) *tp.Rerror {
	uri := ctx.Output().UriObject()
	if _, ok := uri.Query()[SECURE_QUERY_KEY]; !ok {
		// if the query parameter SECURE_QUERY_KEY is not existed,
		// do not perform encryption operation to the body!
		return nil
	}
	// perform encryption operation to the body.
	bodyBytes, err := ctx.Output().MarshalBody()
	if err != nil {
		return tp.NewRerror(e.rerrCode, "marshal raw body error", err.Error())
	}
	ciphertext := goutil.AESEncrypt(e.cipherkey, bodyBytes)
	ctx.Output().SetBody(&Encrypt{
		Version:    e.version,
		Ciphertext: goutil.BytesToString(ciphertext),
	})
	return nil
}

func (e *encryptPlugin) PreWritePush(ctx tp.WriteCtx) *tp.Rerror {
	return e.PreWritePull(ctx)
}

func (e *encryptPlugin) PreWriteReply(ctx tp.WriteCtx) *tp.Rerror {
	return e.PreWritePull(ctx)
}

func (e *decryptPlugin) PreReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	uri := ctx.Input().UriObject()
	if _, ok := uri.Query()[SECURE_QUERY_KEY]; !ok {
		// if the query parameter SECURE_QUERY_KEY is not existed,
		// do not perform decryption operation to the body!
		return nil
	}
	// to prepare for decryption.
	ctx.Swap().Store(encrypt_rawbody, ctx.Input().Body())
	ctx.Input().SetBody(new(Encrypt))
	return nil
}

func (e *decryptPlugin) PostReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	rawbody, ok := ctx.Swap().Load(encrypt_rawbody)
	if !ok {
		return nil
	}
	ciphertext := ctx.Input().Body().(*Encrypt).GetCiphertext()
	bodyBytes, err := goutil.AESDecrypt(e.cipherkey, goutil.StringToBytes(ciphertext))
	if err != nil {
		return tp.NewRerror(e.rerrCode, "decrypt ciphertext error", err.Error())
	}
	ctx.Swap().Delete(encrypt_rawbody)
	ctx.Input().SetBody(rawbody)
	err = ctx.Input().UnmarshalBody(bodyBytes)
	if err != nil {
		return tp.NewRerror(e.rerrCode, "unmarshal raw body error", err.Error())
	}
	return nil
}

func (e *decryptPlugin) PreReadReplyBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PreReadPullBody(ctx)
}

func (e *decryptPlugin) PostReadReplyBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PostReadPullBody(ctx)
}

func (e *decryptPlugin) PreReadPushBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PreReadPullBody(ctx)
}

func (e *decryptPlugin) PostReadPushBody(ctx tp.ReadCtx) *tp.Rerror {
	return e.PostReadPullBody(ctx)
}
