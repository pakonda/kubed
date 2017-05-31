package main

import (
	"flag"
	"log"
	"os"

	v "github.com/appscode/go/version"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	defer logs.FlushLogs()

	rootCmd := &cobra.Command{
		Use:   "kubed",
		Short: `Run Kubernetes Daemon for AppsCode`,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	logs.InitLogs()

	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(v.NewCmdVersion())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}