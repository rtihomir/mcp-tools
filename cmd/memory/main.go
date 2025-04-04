package main

import (
	"fmt"

	"github.com/rtihomir/mcp-tools/internal/memory/config"
	"github.com/rtihomir/mcp-tools/internal/memory/store"
)

func main() {
	c := config.NewConfig()
	s := store.NewStore(c)
	fmt.Println(s)
}
