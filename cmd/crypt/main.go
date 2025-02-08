package main

import (
	cmd "github.com/BuddhiLW/crypt/pkg/encrypt/cmd"
)

// Binary-commands tree-branches will grow from the Root.
func main() {
	cmd.RootCmd.Exec()
}
