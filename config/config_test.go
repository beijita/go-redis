package config

import (
	"fmt"
	"testing"
)

func TestSetupConfig(t *testing.T) {

	SetupConfig("../redis.config")
	fmt.Println(Properties)
}
