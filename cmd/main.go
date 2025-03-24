package main

import (
	"LogManager/cmd/command"
	"LogManager/common"
	"log"
	"math/rand"
	"os"
	"time"
)

func init() {
	if err := common.InitConfigManager(); err != nil {
		log.Printf("Failed to initialize config manager: %v", err)
		os.Exit(1)
	}
}

func main() {
	rand.NewSource(time.Now().UnixNano())
	NewLogManager := command.LogManagerCmd()
	if err := NewLogManager.Execute(); err != nil {
		os.Exit(1)
	}
}
