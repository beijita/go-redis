package config

import (
	"fmt"
	"testing"
)

func TestSetupConfig(t *testing.T) {

	SetupConfig("../redis.server")
	fmt.Println(Properties)
}
