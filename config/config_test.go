package config

import (
	"fmt"
	"testing"
)

func TestSecret(t *testing.T) {

	var bts = []byte("hahaser_@1")
	fmt.Printf("%+v\n", bts)
	fmt.Printf("long of bytes : %d\n", len(bts))
}
