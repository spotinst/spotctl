package wave

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/wave-operator/tide"
	"k8s.io/apimachinery/pkg/util/wait"

	spoterrors "github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/wave"
)

type CmdDelete struct {
	cmd  *cobra.Command
	opts CmdDeleteOptions
}

type CmdDeleteOptions struct {
	*CmdOptions
	ClusterID   string
	DeleteOcean bool
	Purge       bool
}

const (
	deletionTimeout = 5 * time.Minute
	pollInterval    = 5 * time.Second
)

func (x *CmdDeleteOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id")
	fs.BoolVar(&x.DeleteOcean, flags.FlagWaveDeleteOceanCluster, x.DeleteOcean, "delete ocean cluster")
	fs.BoolVar(&x.Purge, flags.FlagWaveDeleteClusterPurge, x.Purge, "delete all configuration (requires admin cluster access)")
}

func NewCmdDelete(opts *CmdOptions) *cobra.Command {
	return newCmdDelete(opts).cmd
}

func newCmdDelete(opts *CmdOptions) *CmdDelete {
	var cmd CmdDelete

	cmd.cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete a Wave cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDeleteOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
	x.initFlags(fs)
}

func (x *CmdDelete) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDeleteOptions) Validate() error {
	if x.ClusterID == "" {
		return spoterrors.Required(flags.FlagWaveClusterID)
	}
	return x.CmdOptions.Validate()
}

func (x *CmdDelete) Run(ctx context.Context) error {
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

func (x *CmdDelete) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDelete) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDelete) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
	}

	_, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	waveClient, err := spotClient.Services().Wave()
	if err != nil {
		return err
	}

	if x.opts.Purge {
		cluster, err := waveClient.GetCluster(ctx, x.opts.ClusterID)
		if err != nil {
			return err
		}
		if err := wave.ValidateClusterContext(cluster.Name); err != nil {
			return fmt.Errorf("cluster context validation failure, %w", err)
		}
	}

	logger := getWaveLogger()

	logger.Info("Deleting Wave cluster ...")
	if err := waveClient.DeleteCluster(ctx, x.opts.ClusterID, x.opts.DeleteOcean); err != nil {
		return err
	}

	err = wait.Poll(pollInterval, deletionTimeout, func() (bool, error) {
		_, err := waveClient.GetCluster(ctx, x.opts.ClusterID)
		if err != nil && errors.As(err, &spot.ResourceDoesNotExistError{}) {
			return true, nil
		} else {
			return false, nil
		}
	})

	if !x.opts.Purge {
		logger.Info("Wave cluster deleted")
		return nil
	} else {
		// Purge Wave Environment CRD and tide RBAC
		logger.Info("Deleting Wave configuration ...")
		manager, err := tide.NewManager(logger)
		if err != nil {
			return err
		}
		err = manager.DeleteConfiguration(true)
		if err != nil {
			return fmt.Errorf("could not delete wave configuration, %w", err)
		}
		err = manager.DeleteTideRBAC()
		if err != nil {
			return fmt.Errorf("could not delete tide rbac objects, %w", err)
		}
		logger.Info("Wave has been removed")
	}

	// TODO Delete kubernetes cluster if it was provisioned

	return nil
}
