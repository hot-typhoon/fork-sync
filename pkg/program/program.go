package program

import (
	"context"
)

var Ctx, Cancel = context.WithCancel(context.Background())
