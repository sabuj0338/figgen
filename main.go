package main

import (
	"github.com/joho/godotenv"
	"github.com/sabujislam/figgen/cmd"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	cmd.Execute()
}
