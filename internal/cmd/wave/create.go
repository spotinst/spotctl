package wave

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/dep"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/thirdparty/commands/eksctl"
	"github.com/spotinst/spotctl/internal/uuid"
	"github.com/spotinst/spotctl/internal/wave"
	"github.com/theckman/yacspin"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type CmdCreate struct {
	cmd  *cobra.Command
	opts CmdCreateOptions
}

type CmdCreateOptions struct {
	*CmdOptions
	ConfigFile        string
	ClusterID         string
	ClusterName       string
	Region            string
	Tags              []string
	KubernetesVersion string
}

func (x *CmdCreateOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&x.ConfigFile, flags.FlagWaveConfigFile, "f", x.ConfigFile, "load configuration from a file (or stdin if set to '-')")
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id (will be created if empty)")
	fs.StringVar(&x.ClusterName, flags.FlagWaveClusterName, x.ClusterName, "cluster name (generated if unspecified, e.g. \"wave-9d4afe95\")")
	fs.StringVar(&x.Region, flags.FlagWaveRegion, os.Getenv("AWS_REGION"), "region in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Tags, "tags", x.Tags, "list of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")")
	fs.StringVar(&x.KubernetesVersion, "kubernetes-version", "1.18", "kubernetes version")
}

func NewCmdCreate(opts *CmdOptions) *cobra.Command {
	return newCmdCreate(opts).cmd
}

func newCmdCreate(opts *CmdOptions) *CmdCreate {
	var cmd CmdCreate

	cmd.cmd = &cobra.Command{
		Use:           "create",
		Short:         "Create a new Wave installation",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdCreate) preRun(ctx context.Context) error {
	// Call to the the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return x.installDeps(ctx)
}

func (x *CmdCreate) installDeps(ctx context.Context) error {
	// Initialize a new dependency manager.
	dm, err := x.opts.Clientset.NewDepManager()
	if err != nil {
		return err
	}

	// Install options.
	installOpts := []dep.InstallOption{
		dep.WithInstallPolicy(dep.InstallPolicy(x.opts.InstallPolicy)),
		dep.WithNoninteractive(x.opts.Noninteractive),
		dep.WithDryRun(x.opts.DryRun),
	}

	// Install!
	return dm.InstallBulk(ctx, dep.DefaultDependencyListKubernetes(), installOpts...)
}

func (x *CmdCreateOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
	x.initFlags(fs)
}

func (x *CmdCreate) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdCreateOptions) Validate() error {
	if x.ClusterID != "" && x.ClusterName != "" {
		return errors.RequiredXor(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
	}
	if x.ClusterID != "" && x.ConfigFile != "" {
		return errors.RequiredXor(flags.FlagWaveClusterID, flags.FlagWaveConfigFile)
	}
	if x.ClusterName != "" && x.ConfigFile != "" {
		return errors.RequiredXor(flags.FlagWaveClusterName, flags.FlagWaveConfigFile)
	}
	return x.CmdOptions.Validate()
}

func (x *CmdCreate) Run(ctx context.Context) error {
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

func (x *CmdCreate) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdCreate) validate(ctx context.Context) error {
	return x.opts.Validate()
}

// TODO(liran): WARNING: This is the ugliest code in the world, but it seems to work (for now).
func (x *CmdCreate) run(ctx context.Context) error {

	cfg := yacspin.Config{
		Frequency:       250 * time.Millisecond,
		CharSet:         yacspin.CharSets[33],
		Suffix:          " wave",
		SuffixAutoColon: true,
		Message:         "start",
		StopCharacter:   "✓",
		StopColors:      []string{"green"},
	}
	spinner, err := yacspin.New(cfg)
	spinner.Start()

	if x.opts.ClusterID == "" { // create a new cluster
		spinner.Message("creating")
		if x.opts.ConfigFile != "" { // extract from config
			type clusterConfig struct {
				Metadata struct {
					Name   string `json:"name"`
					Region string `json:"region"`
				} `json:"metadata"`
			}

			var r io.Reader
			var err error

			if x.opts.ConfigFile == "-" { // read from standard input
				r = os.Stdin
			} else { // read from file
				f, err := os.Open(x.opts.ConfigFile)
				if err != nil {
					return err
				}
				defer f.Close()
				r = f
			}

			c := new(clusterConfig)
			if err = yaml.NewYAMLOrJSONDecoder(r, 4096).Decode(c); err != nil {
				return err
			}

			x.opts.ClusterName = c.Metadata.Name
			x.opts.Region = c.Metadata.Region
		} else if x.opts.ClusterName == "" { // generate unique name
			x.opts.ClusterName = fmt.Sprintf("wave-%s", uuid.NewV4().Short())
		}

		// TODO(liran/validation): Validate it elsewhere.
		if x.opts.ClusterName == "" {
			return errors.Required(flags.FlagWaveClusterName)
		}
		if x.opts.Region == "" {
			return errors.Required(flags.FlagWaveRegion)
		}

		cloudProviderOpts := []cloud.ProviderOption{
			cloud.WithProfile(x.opts.Profile),
			cloud.WithRegion(x.opts.Region),
		}

		cloudProvider, err := x.opts.Clientset.NewCloud(
			cloud.ProviderName(x.opts.CloudProvider), cloudProviderOpts...)
		if err != nil {
			return err
		}
		sc, err := x.newStackCollection(cloudProvider)
		if err != nil {
			return err
		}

		createCluster := false
		if _, err = sc.describeStacks(); err != nil {
			if err.Error() == sc.errStackNotFound().Error() {
				createCluster = true
			}
		}

		cmdEksctl, err := x.opts.Clientset.NewCommand(eksctl.CommandName)
		if err != nil {
			return err
		}

		spinner.Message("creating eks cluster " + x.opts.ClusterName)
		if createCluster {
			if err = cmdEksctl.Run(ctx, x.buildEksctlCreateClusterArgs()...); err != nil {
				return err
			}
		}

		spinner.Message("creating node groups")
		if err = cmdEksctl.Run(ctx, x.buildEksctlCreateNodeGroupArgs()...); err != nil {
			return err
		}
	} else { // import an existing cluster
		spinner.Message("importing")

		// TODO(liran/validation): Validate it elsewhere.
		if x.opts.Region == "" {
			return errors.Required(flags.FlagWaveRegion)
		}

		spotClientOpts := []spot.ClientOption{
			spot.WithCredentialsProfile(x.opts.Profile),
		}

		spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
		if err != nil {
			return err
		}

		oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
		if err != nil {
			return err
		}

		c, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
		if err != nil {
			return err
		}

		x.opts.ClusterName = c.Name
	}

	spinner.Message(fmt.Sprintf("Verified cluster %q", x.opts.ClusterName))

	cloudProviderOpts := []cloud.ProviderOption{
		cloud.WithProfile(x.opts.Profile),
		cloud.WithRegion(x.opts.Region),
	}
	cloudProvider, err := x.opts.Clientset.NewCloud(
		cloud.ProviderName(x.opts.CloudProvider), cloudProviderOpts...)
	if err != nil {
		return err
	}

	spinner.Message("setting roles")
	var stacks []*Stack
	sc, err := x.newStackCollection(cloudProvider)
	if err != nil {
		return err
	}
	err = wait.Poll(10*time.Second, x.opts.Timeout, func() (bool, error) {
		stacks, err = sc.describeStacks()
		if err != nil {
			return false, nil
		}
		return len(stacks) > 0, nil
	})
	for _, stack := range stacks {
		var roleName string
		for _, output := range stack.Outputs {
			if aws.StringValue(output.OutputKey) == "InstanceRoleARN" {
				a, _ := arn.Parse(aws.StringValue(output.OutputValue))
				roleName = strings.TrimPrefix(a.Resource, "role/")
				break
			}
		}
		if roleName != "" {
			const s3FullAccessProfileARN = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
			if err = cloudProvider.IAM().AttachRolePolicy(ctx, roleName,
				s3FullAccessProfileARN); err != nil {
				return err
			}
		}
	}

	spinner.Message("installing wave")
	manager, err := wave.NewManager(x.opts.ClusterName, getSpinnerLogger(x.opts.ClusterName, spinner)) // pass in name to validate ocean controller configuration
	if err != nil {
		return err
	}

	err = manager.Create()
	if err != nil {
		spinner.StopFail()
		return err
	}
	spinner.StopMessage("wave operator is managing components")
	spinner.Stop()
	return nil
}

func (x *CmdCreate) buildEksctlCreateClusterArgs() []string {
	log.Debugf("Building up command arguments (create cluster)")

	args := []string{
		"create", "cluster",
		"--timeout", "60m",
		"--color", "false",
		"--without-nodegroup",
	}

	if len(x.opts.ConfigFile) > 0 {
		args = append(args, "--config-file", x.opts.ConfigFile)

	} else {
		if len(x.opts.ClusterName) > 0 {
			args = append(args, "--name", x.opts.ClusterName)
		}

		if len(x.opts.Region) > 0 {
			args = append(args, "--region", x.opts.Region)
		}

		if len(x.opts.Tags) > 0 {
			args = append(args, "--tags", strings.Join(x.opts.Tags, ","))
		}

		if len(x.opts.KubernetesVersion) > 0 {
			args = append(args, "--version", x.opts.KubernetesVersion)
		}
	}

	if x.opts.Verbose {
		args = append(args, "--verbose", "4")
	} else {
		args = append(args, "--verbose", "0")
	}

	return args
}

func (x *CmdCreate) buildEksctlCreateNodeGroupArgs() []string {
	log.Debugf("Building up command arguments (create nodegroup)")

	args := []string{
		"create", "nodegroup",
		"--timeout", "60m",
		"--color", "false",
	}

	if len(x.opts.ConfigFile) > 0 {
		args = append(args, "--config-file", x.opts.ConfigFile)

	} else {
		if len(x.opts.ClusterName) > 0 {
			args = append(args,
				"--cluster", x.opts.ClusterName,
				"--name", fmt.Sprintf("ocean-%s", uuid.NewV4().Short()))
		}

		if len(x.opts.Region) > 0 {
			args = append(args, "--region", x.opts.Region)
		}

		if len(x.opts.Tags) > 0 {
			args = append(args, "--tags", strings.Join(x.opts.Tags, ","))
		}

		if len(x.opts.KubernetesVersion) > 0 {
			args = append(args, "--version", x.opts.KubernetesVersion)
		}

		if len(x.opts.Profile) > 0 {
			args = append(args, "--spot-profile", x.opts.Profile)
		}

		args = append(args, "--spot-ocean")
	}

	if x.opts.Verbose {
		args = append(args, "--verbose", "4")
	} else {
		args = append(args, "--verbose", "0")
	}

	return args
}

type stackCollection struct {
	clusterName string
	svc         *cloudformation.CloudFormation
}

func (x *CmdCreate) newStackCollection(cloudProvider cloud.Provider) (*stackCollection, error) {
	sess, err := cloudProvider.Session(x.opts.Region, x.opts.Profile)
	if err != nil {
		return nil, err
	}
	return &stackCollection{
		clusterName: x.opts.ClusterName,
		svc:         cloudformation.New(sess.(*session.Session)),
	}, nil
}

type Stack = cloudformation.Stack

// listStacksMatching gets all of CloudFormation stacks with names matching nameRegex.
func (c *stackCollection) listStacksMatching(nameRegex string, statusFilters ...string) ([]*Stack, error) {
	var (
		subErr error
		stack  *Stack
	)

	re, err := regexp.Compile(nameRegex)
	if err != nil {
		return nil, fmt.Errorf("cannot list stacks: %w", err)
	}
	input := &cloudformation.ListStacksInput{
		StackStatusFilter: defaultStackStatusFilter(),
	}
	if len(statusFilters) > 0 {
		input.StackStatusFilter = aws.StringSlice(statusFilters)
	}
	var stacks []*Stack
	pager := func(p *cloudformation.ListStacksOutput, _ bool) bool {
		for _, s := range p.StackSummaries {
			if re.MatchString(*s.StackName) {
				stack, subErr = c.describeStack(&Stack{
					StackName: s.StackName,
					StackId:   s.StackId,
				})
				if subErr != nil {
					return false
				}
				stacks = append(stacks, stack)
			}
		}
		return true
	}

	if err = c.svc.ListStacksPages(input, pager); err != nil {
		return nil, err
	}
	if subErr != nil {
		return nil, subErr
	}

	return stacks, nil
}

// listStacks gets all of CloudFormation stacks.
func (c *stackCollection) listStacks(statusFilters ...string) ([]*Stack, error) {
	return c.listStacksMatching(fmtStacksRegexForCluster(c.clusterName), statusFilters...)
}

func fmtStacksRegexForCluster(name string) string {
	const ourStackRegexFmt = "^(eksctl|EKS)-%s-((cluster|nodegroup-.+|addon-.+)|(VPC|ServiceRole|ControlPlane|DefaultNodeGroup))$"
	return fmt.Sprintf(ourStackRegexFmt, name)
}

// describeStack describes a cloudformation stack.
func (c *stackCollection) describeStack(i *Stack) (*Stack, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: i.StackName,
	}
	resp, err := c.svc.DescribeStacks(input)
	if err != nil {
		return nil, fmt.Errorf("describing CloudFormation stack %q: %w", *i.StackName, err)
	}
	return resp.Stacks[0], nil
}

// describeStacks describes cloudformation stacks.
func (c *stackCollection) describeStacks() ([]*Stack, error) {
	log.Debugf("Describing stacks")

	stacks, err := c.listStacks()
	if err != nil {
		return nil, fmt.Errorf("describing CloudFormation stacks for %q: %w", c.clusterName, err)
	}
	if len(stacks) == 0 {
		return nil, c.errStackNotFound()
	}

	var out []*Stack
	for _, s := range stacks {
		if *s.StackStatus == cloudformation.StackStatusDeleteComplete {
			continue
		}
		out = append(out, s)
	}

	return out, nil
}

func (c *stackCollection) errStackNotFound() error {
	return fmt.Errorf("no eksctl-managed CloudFormation stacks found for %q", c.clusterName)
}

func allNonDeletedStackStatuses() []string {
	return []string{
		cloudformation.StackStatusCreateInProgress,
		cloudformation.StackStatusCreateFailed,
		cloudformation.StackStatusCreateComplete,
		cloudformation.StackStatusRollbackInProgress,
		cloudformation.StackStatusRollbackFailed,
		cloudformation.StackStatusRollbackComplete,
		cloudformation.StackStatusDeleteInProgress,
		cloudformation.StackStatusDeleteFailed,
		cloudformation.StackStatusUpdateInProgress,
		cloudformation.StackStatusUpdateCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusUpdateRollbackInProgress,
		cloudformation.StackStatusUpdateRollbackFailed,
		cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateRollbackComplete,
		cloudformation.StackStatusReviewInProgress,
	}
}

func defaultStackStatusFilter() []*string {
	return aws.StringSlice(allNonDeletedStackStatuses())
}
