// Command mia is the memoria binary (MEMORIA_APP.md §2).
package main

import (
	"os"

	"github.com/joakimmj/memoria/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
