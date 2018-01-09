// Dynamically ignoring the case of path
package ignore_case

import (
	"strings"

	tp "github.com/henrylee2cn/teleport"
)

// NewIgnoreCase Returns a ignore_case plugin.
func NewIgnoreCase() *ignoreCase {
	return &ignoreCase{}
}

type ignoreCase struct{}

func (i *ignoreCase) Name() string {
	return "ignore_case"
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
