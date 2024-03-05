package main

import (
	"fmt"

	"github.com/hnimtadd/run/internal/manager"
)

func main() {
	server := manager.NewServer(nil)
	fmt.Println(server)
}
