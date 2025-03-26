package common

import (
	"github.com/songquanpeng/one-api/common/env"
)

var EnableTrace = false

func InitTrace() {
	EnableTrace = env.Bool("ENABLE_TRACE", false)
}
