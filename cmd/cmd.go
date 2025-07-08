package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daboyuka/hs/cmd/exprcmd"
	"github.com/daboyuka/hs/cmd/httpcmd"
	"github.com/daboyuka/hs/hsruntime/cookie"
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "hs",
	Short: "a tool for batch, data-driven HTTP requests",
	Long:  `HScript is a tool for making batch, data-driven HTTP requests. See full docs at https://github.com/daboyuka/hs.`,
}

func init() {
	rootCmd.AddGroup(httpcmd.Group)
	rootCmd.AddCommand(httpcmd.Commands...)

	rootCmd.AddCommand(exprcmd.Cmd)

	oldHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		oldHelp(c, args)
		out := c.OutOrStdout()
		_, _ = fmt.Fprintln(out, "")
		_, _ = fmt.Fprintln(out, "Supported browser loaders:", strings.Join(cookie.AllSupportedBrowsers(), " "))
	})
}
