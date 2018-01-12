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

func (i *ignoreCase) Name() string {
	return "ignoreCase"
}

func (i *ignoreCase) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
	// Dynamic transformation path is lowercase
	ctx.Url().Path = strings.ToLower(ctx.Url().Path)
	return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
	// Dynamic transformation path is lowercase
	ctx.Url().Path = strings.ToLower(ctx.Url().Path)
	return nil
}
