package ocean

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotinst-sdk-go/service/ocean"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

type (
	CmdGetClusterLogs struct {
		cmd  *cobra.Command
		opts CmdGetClusterLogsOptions
	}

	CmdGetClusterLogsOptions struct {
		*CmdGetClusterOptions
	}
)

func NewCmdGetClusterLogs(opts *CmdGetClusterOptions) *cobra.Command {
	return newCmdGetClusterLogs(opts).cmd
}

func newCmdGetClusterLogs(opts *CmdGetClusterOptions) *CmdGetClusterLogs {
	var cmd CmdGetClusterLogs

	cmd.cmd = &cobra.Command{
		Use:           "logs",
		Short:         "Get Logs from ocean ecs",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"ecs"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetClusterLogs) Run(ctx context.Context) error {
	steps := []func(context.Context) error{
		x.survey,
		x.log,
		x.validate,
		x.run,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *CmdGetClusterLogs) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetClusterLogs) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetClusterLogs) validate(ctx context.Context) error {
	return x.opts.Validate()
}

// get cluster logs

func (x *CmdGetClusterLogs) run(ctx context.Context) error {
	sess := session.New()
	svc := ocean.New(sess)

	today := strconv.FormatInt(time.Now().Sub(time.Unix(0, 0)).Milliseconds(), 10)
	lastWeek := strconv.FormatInt(time.Now().Sub(time.Unix(0, 0)).Milliseconds()-604800000, 10)

	t := time.Now().Unix()
	timeT := time.Unix(t, 0)
	fmt.Printf("From 1 week ago until: %s\n", timeT)

	// Get log events.

	cluster := os.Getenv("OCEAN_CLUSTER")
	out, err := svc.CloudProviderAWS().GetLogEvents(ctx, &aws.GetLogEventsInput{
		ClusterID: spotinst.String(cluster),
		FromDate:  spotinst.String(lastWeek),
		ToDate:    spotinst.String(today),
	})
	if err != nil {
		log.Fatalf("spotinst: failed to get log events: %v", err)
	}

	// Output log events, if any.
	if len(out.Events) > 0 {
		for _, event := range out.Events {
			fmt.Printf("%s [%s] %s\n",
				spotinst.TimeValue(event.CreatedAt).Format(time.RFC3339),
				spotinst.StringValue(event.Severity),
				spotinst.StringValue(event.Message))
		}
	}

	return err
}

func (x *CmdGetClusterLogsOptions) Init(fs *pflag.FlagSet, opts *CmdGetClusterOptions) {
	x.CmdGetClusterOptions = opts
}

func (x *CmdGetClusterLogsOptions) Validate() error {
	return x.CmdGetClusterOptions.Validate()
}
