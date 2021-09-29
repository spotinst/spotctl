package ocean

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/kubernetes"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/ocean/ofas"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/thirdparty/commands/eksctl"
	"github.com/spotinst/spotctl/internal/uuid"
	"github.com/theckman/yacspin"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/dep"
	"github.com/spotinst/spotctl/internal/flags"
)

type (
	CmdSparkCreateCluster struct {
		cmd  *cobra.Command
		opts CmdSparkCreateClusterOptions
	}

	// TODO
	/*
		- What happens if I create a cluster with a controllerClusterId that already exists?
	*/

	CmdSparkCreateClusterOptions struct {
		*CmdSparkCreateOptions
		ClusterID         string
		ClusterName       string
		Region            string
		Tags              []string
		KubernetesVersion string
	}
)

const (
	defaultK8sVersion   = "1.21"
	spotSystemNamespace = "spot-system" // TODO Get this from deployer job config
)

func NewCmdSparkCreateCluster(opts *CmdSparkCreateOptions) *cobra.Command {
	return newCmdSparkCreateCluster(opts).cmd
}

func newCmdSparkCreateCluster(opts *CmdSparkCreateOptions) *CmdSparkCreateCluster {
	var cmd CmdSparkCreateCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Create a new Ocean for Apache Spark cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"c"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdSparkCreateCluster) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return x.installDeps(ctx)
}

func (x *CmdSparkCreateCluster) Run(ctx context.Context) error {
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

func (x *CmdSparkCreateCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkCreateCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkCreateCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdSparkCreateCluster) run(ctx context.Context) error {
	shouldCreateCluster := x.opts.ClusterID == ""

	if shouldCreateCluster {
		log.Infof("Will create Ocean for Apache Spark cluster")

		if x.opts.ClusterName == "" {
			// Generate unique name
			x.opts.ClusterName = fmt.Sprintf("ocean-spark-cluster-%s", uuid.NewV4().Short())
		}

		cloudProviderOpts := []cloud.ProviderOption{
			cloud.WithProfile(x.opts.Profile),
			cloud.WithRegion(x.opts.Region),
		}

		cloudProvider, err := x.opts.Clientset.NewCloud(cloud.ProviderName(x.opts.CloudProvider), cloudProviderOpts...)
		if err != nil {
			return fmt.Errorf("could not get cloud provider, %w", err)
		}

		stackCollection, err := x.newStackCollection(cloudProvider)
		if err != nil {
			return fmt.Errorf("could not get stack collection, %w", err)
		}

		stackExists := true
		if _, err = stackCollection.describeStacks(); err != nil {
			if err.Error() == stackCollection.errStackNotFound().Error() {
				stackExists = false
			}
		}

		cmdEksctl, err := x.opts.Clientset.NewCommand(eksctl.CommandName)
		if err != nil {
			return fmt.Errorf("could not create eksctl command, %w", err)
		}

		// TODO Allow creation of cluster if previous stack failed
		// TODO Check for in-progress stacks
		if !stackExists {
			spinner := startSpinnerWithMessage(fmt.Sprintf("Creating EKS cluster %s", x.opts.ClusterName))
			createClusterArgs := x.buildEksctlCreateClusterArgs()
			if err := cmdEksctl.Run(ctx, createClusterArgs...); err != nil {
				stopSpinnerWithMessage(spinner, "Could not create EKS cluster", true)
				return fmt.Errorf("could not create EKS cluster, %w", err)
			}
			stopSpinnerWithMessage(spinner, "EKS cluster created", false)
		}

		// TODO Don't create nodegroup if it already exists
		spinner := startSpinnerWithMessage("Creating Ocean node group")
		createNodeGroupArgs := x.buildEksctlCreateNodeGroupArgs()
		if err := cmdEksctl.Run(ctx, createNodeGroupArgs...); err != nil {
			stopSpinnerWithMessage(spinner, "Could not create node group", true)
			return fmt.Errorf("could not create node group, %w", err)
		}
		stopSpinnerWithMessage(spinner, "Spot Ocean node group created", false)

	} else {
		log.Infof("Will deploy Ocean for Apache Spark on cluster %s", x.opts.ClusterID)

		spotClientOpts := []spot.ClientOption{
			spot.WithCredentialsProfile(x.opts.Profile),
		}

		spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
		if err != nil {
			return fmt.Errorf("could not get Spot client, %w", err)
		}

		oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
		if err != nil {
			return fmt.Errorf("could not get Ocean client, %w", err)
		}

		oceanCluster, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
		if err != nil {
			return fmt.Errorf("could not get Ocean cluster, %w", err)
		}

		x.opts.ClusterName = oceanCluster.Name // TODO Does this have to be the controller cluster id?
	}

	if err := ofas.ValidateClusterContext(ctx, x.opts.ClusterName); err != nil {
		return fmt.Errorf("cluster context validation failure, %w", err)
	}

	log.Infof("Verified cluster %s", x.opts.ClusterName)

	// TODO Should we be doing this here? (does not play well with beta versions)
	log.Infof("Updating Ocean controller")
	if err := updateOceanController(ctx); err != nil {
		return fmt.Errorf("could not apply ocean update, %w", err)
	}

	log.Infof("Creating namespace %s", spotSystemNamespace)
	if err := kubernetes.EnsureNamespace(ctx, spotSystemNamespace); err != nil {
		return fmt.Errorf("could not create namespace, %w", err)
	}

	log.Infof("Creating deployer RBAC")
	if err := ofas.CreateDeployerRBAC(ctx, spotSystemNamespace); err != nil {
		return fmt.Errorf("could not create deployer rbac, %w", err)
	}

	spinner := startSpinnerWithMessage("Installing Ocean for Apache Spark")
	if err := ofas.Deploy(ctx, spotSystemNamespace); err != nil {
		stopSpinnerWithMessage(spinner, "Ocean for Apache Spark installation failure", true)
		return fmt.Errorf("could not deploy, %w", err)
	}
	stopSpinnerWithMessage(spinner, "Ocean for Apache Spark installed", false)

	log.Infof("Cluster %s successfully created", x.opts.ClusterName)

	return nil
}

func (x *CmdSparkCreateClusterOptions) Init(fs *pflag.FlagSet, opts *CmdSparkCreateOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkCreateClusterOptions) initDefaults(opts *CmdSparkCreateOptions) {
	x.CmdSparkCreateOptions = opts
}

func (x *CmdSparkCreateClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOFASClusterID, x.ClusterID, "cluster id (will be created if empty)") // TODO better explanation that this is an import
	fs.StringVar(&x.ClusterName, flags.FlagOFASClusterName, x.ClusterName, "cluster name (will be created if empty)")
	fs.StringVar(&x.Region, flags.FlagOFASClusterRegion, os.Getenv("AWS_REGION"), "region in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Tags, "tags", x.Tags, "list of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")")
	fs.StringVar(&x.KubernetesVersion, "kubernetes-version", defaultK8sVersion, "kubernetes version")
}

func (x *CmdSparkCreateClusterOptions) Validate() error {
	if x.ClusterID != "" && x.ClusterName != "" {
		return errors.RequiredXor(flags.FlagOFASClusterID, flags.FlagOFASClusterName)
	}
	if x.Region == "" {
		return errors.Required(flags.FlagOFASClusterRegion)
	}
	return x.CmdSparkCreateOptions.Validate()
}

func (x *CmdSparkCreateCluster) installDeps(ctx context.Context) error {
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

func (x *CmdSparkCreateCluster) buildEksctlCreateClusterArgs() []string {
	log.Debugf("Building up command arguments (create cluster)")

	args := []string{
		"create", "cluster",
		"--timeout", "60m",
		"--color", "false",
		"--without-nodegroup",
	}

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

	if x.opts.Verbose {
		args = append(args, "--verbose", "4")
	} else {
		args = append(args, "--verbose", "0")
	}

	return args
}

func (x *CmdSparkCreateCluster) buildEksctlCreateNodeGroupArgs() []string {
	log.Debugf("Building up command arguments (create nodegroup)")

	args := []string{
		"create", "nodegroup",
		"--timeout", "60m",
		"--color", "false",
	}

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

	//if len(x.opts.Profile) > 0 {
	//	args = append(args, "--spot-profile", x.opts.Profile)
	//}

	args = append(args, "--managed=false") // Not EKS managed
	args = append(args, "--spot-ocean")

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

func (x *CmdSparkCreateCluster) newStackCollection(cloudProvider cloud.Provider) (*stackCollection, error) {
	sess, err := cloudProvider.Session(x.opts.Region, x.opts.Profile)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud provider session, %w", err)
	}

	return &stackCollection{
		clusterName: x.opts.ClusterName,
		svc:         cloudformation.New(sess.(*session.Session)),
	}, nil
}

type Stack = cloudformation.Stack

func fmtStacksRegexForCluster(name string) string {
	const ourStackRegexFmt = "^(eksctl|EKS)-%s-((cluster|nodegroup-.+|addon-.+)|(VPC|ServiceRole|ControlPlane|DefaultNodeGroup))$"
	return fmt.Sprintf(ourStackRegexFmt, name)
}

// listStacks gets all of CloudFormation stacks.
func (c *stackCollection) listStacks(statusFilters ...string) ([]*Stack, error) {
	return c.listStacksMatching(fmtStacksRegexForCluster(c.clusterName), statusFilters...)
}

func defaultStackStatusFilter() []*string {
	return aws.StringSlice(allNonDeletedStackStatuses())
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

func (c *stackCollection) errStackNotFound() error {
	return fmt.Errorf("no eksctl-managed CloudFormation stacks found for %q", c.clusterName)
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

func updateOceanController(ctx context.Context) error {
	conf, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("could not get cluster config, %w", err)
	}

	// TODO Stop using beta when it is out of beta
	const oceanControllerURL = "https://s3.amazonaws.com/spotinst-public/integrations/kubernetes/cluster-controller-beta/spotinst-kubernetes-cluster-controller-ga.yaml"
	//const oceanControllerURL = "https://s3.amazonaws.com/spotinst-public/integrations/kubernetes/cluster-controller/spotinst-kubernetes-cluster-controller-ga.yaml"

	res, err := http.Get(oceanControllerURL)
	if err != nil {
		return fmt.Errorf("error fetching ocean manifests, %w", err)
	}

	data, err := ioutil.ReadAll(res.Body)
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Warnf("Could not close response body, err: %s", err.Error())
		}
	}()
	if err != nil {
		return fmt.Errorf("error reading ocean manifests, %w", err)
	}

	delim := regexp.MustCompile("(?m)^---$")
	objects := delim.Split(string(data), -1)

	whitespace := regexp.MustCompile("^[[:space:]]*$")

	for _, o := range objects {
		if whitespace.Match([]byte(o)) {
			log.Debugf("Whitespace match: %s", o)
			continue
		}
		err := kubernetes.DoServerSideApply(ctx, conf, o, log.GetLogrLogger())
		if err != nil {
			return fmt.Errorf("error applying object from manifests <<%s>>, %w", o, err)
		}
	}

	return nil
}

// startSpinnerWithMessage starts a new spinner logger with the given message.
// Best effort. On error, logs the message using the default logger and returns nil.
func startSpinnerWithMessage(message string) *yacspin.Spinner {
	cfg := yacspin.Config{
		Frequency:         250 * time.Millisecond,
		CharSet:           yacspin.CharSets[33],
		Suffix:            " ",
		SuffixAutoColon:   false,
		Message:           message,
		StopCharacter:     "✓",
		StopColors:        []string{"green"},
		StopFailCharacter: "x",
		StopFailColors:    []string{"red"},
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		log.Warnf("Could not create spinner, err: %s", err.Error())
		log.Infof("%s", message)
		return nil
	}

	err = spinner.Start()
	if err != nil {
		log.Warnf("Could not start spinner, err: %s", err.Error())
		log.Infof("%s", message)
		return nil
	}

	return spinner
}

// stopSpinnerWithMessage stops the given spinner, setting the message as the stop message
// or the stop failure message. Fail determines if the spinner should succeed or fail.
// Best effort. On error, log the message using the default logger.
func stopSpinnerWithMessage(spinner *yacspin.Spinner, message string, fail bool) {
	if spinner != nil {
		var stopError error
		if fail {
			spinner.StopFailMessage(message)
			stopError = spinner.StopFail()
		} else {
			spinner.StopMessage(message)
			stopError = spinner.Stop()
		}
		if stopError != nil {
			log.Warnf("Could not stop spinner, err: %s", stopError.Error())
			log.Infof(message)
		}
	} else {
		log.Infof(message)
	}
}
