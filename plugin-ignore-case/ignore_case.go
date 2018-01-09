package ignore_case

import (
	"fmt"
	"strings"

	tp "github.com/henrylee2cn/teleport"
)

// NewIgnoreCase Returns a ignore_case plugin.
func NewIgnoreCase() *ignoreCase {
	return &ignoreCase{}
}

type ignoreCase struct {
}

func (i *ignoreCase) Name() string {
	return "ignore_case"
}

func (i *ignoreCase) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
	fmt.Println("papapapappapa", ctx.Path())
	// Dynamic transformation path is lowercase
	ctx.Input().SetUri(strings.ToLower(ctx.Path()))
	return nil
}
