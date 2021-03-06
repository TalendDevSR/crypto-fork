// +build evm

package db

import (
	"github.com/spf13/cobra"
)

func NewDBCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database Maintenance",
	}
	cmd.AddCommand(
		newPruneDBCommand(),
		newCompactDBCommand(),
		newDumpEVMStateCommand(),
		newDumpEVMStateMultiWriterAppStoreCommand(),
		newDumpEVMStateFromEvmDB(),
		newGetEvmHeightCommand(),
		newGetAppHeightCommand(),
		newAnalyzeCommand(),
		newExtractCurrentStateCommand(),
		newCompareCurrentStateCommand(),
	)
	return cmd
}
