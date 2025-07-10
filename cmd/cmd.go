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
	return RootCmd.Execute()
}

// RootCmd is the base cobra command for hs. Plugins/importers may override its fields to change the help message, etc.
var RootCmd = &cobra.Command{
	Use:   "hs",
	Short: "a tool for batch, data-driven HTTP requests",
	Long:  `HScript is a tool for making batch, data-driven HTTP requests. See full docs at https://github.com/daboyuka/hs`,
}

func init() {
	RootCmd.AddGroup(httpcmd.Group)
	RootCmd.AddCommand(httpcmd.Commands...)

	RootCmd.AddCommand(exprcmd.Cmd)

	RootCmd.AddCommand(&cobra.Command{
		Use:     "init",
		Aliases: []string{"init"},
		Short:   "initialize configuration file",
		Long:    "initialize configuration file",
		Args:    cobra.NoArgs,
		RunE:    cmdInit,
	})

	oldHelp := RootCmd.HelpFunc()
	RootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		oldHelp(c, args)
		out := c.OutOrStdout()
		_, _ = fmt.Fprintln(out, "")
		_, _ = fmt.Fprintln(out, "Supported browser loaders:", strings.Join(cookie.AllSupportedBrowsers(), " "))
	})
}
