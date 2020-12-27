package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vx416/dcard-work/cmd"
)

func main() {
	root := &cobra.Command{}
	root.AddCommand(cmd.Server)
	if err := root.Execute(); err != nil {
		log.Printf("command exec failed, err:%+v", err)
		os.Exit(1)
	}

}
