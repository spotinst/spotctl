package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/log"
)

func Log(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		log.Debugf("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
