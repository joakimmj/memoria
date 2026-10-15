// Command mia is the memoria binary.
package main

import (
	"os"

	"github.com/joakimmj/memoria/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
