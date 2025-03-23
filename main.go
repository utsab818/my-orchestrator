/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/utsab818/my-orchestrator/cmd"

func main() {
	cmd.Execute()
}

// go run main.go worker
// go run main.go worker -p 5557
// go run main.go manager -w 'localhost:5556,localhost:5557'
// go run main.go run --filename task1.json
// go run main.go status
// go run main.go node
// go run main.go stop <id from status>
