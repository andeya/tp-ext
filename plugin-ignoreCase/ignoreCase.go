// Package ignoreCase dynamically ignoring the case of path
package ignoreCase

import (
	"strings"

	tp "github.com/henrylee2cn/teleport"
)

// NewIgnoreCase Returns a ignoreCase plugin.
func NewIgnoreCase() *ignoreCase {
	return &ignoreCase{}
}

type ignoreCase struct{}

var (
	_ tp.PostReadPullHeaderPlugin = new(ignoreCase)
	_ tp.PostReadPushHeaderPlugin = new(ignoreCase)
)

func (i *ignoreCase) Name() string {
	return "ignoreCase"
}

func (i *ignoreCase) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
	// Dynamic transformation path is lowercase
	ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
	return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
	// Dynamic transformation path is lowercase
	ctx.UriObject().Path = strings.ToLower(ctx.UriObject().Path)
	return nil
}
