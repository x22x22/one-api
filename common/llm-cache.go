package common

import (
	"os"
)

var NoCacheDomain = ""

func InitLLMCache() {
	NoCacheDomain = os.Getenv("NO_CACHE_DOMAIN")
}
