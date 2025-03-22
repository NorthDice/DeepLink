package main

import (
	"fmt"
	"github.com/NorthDice/deepLink/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
}
