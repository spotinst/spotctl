package ocean

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	oceanaws "github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/theckman/yacspin"

	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/dep"
	spotctlerrors "github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/kubernetes"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/ocean/ofas"
	"github.com/spotinst/spotctl/internal/ocean/ofas/eks"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/thirdparty/commands/eksctl"
	"github.com/spotinst/spotctl/internal/uuid"
)

type (
	CmdSparkCreateCluster struct {
		cmd  *cobra.Command
		opts CmdSparkCreateClusterOptions
	}

	CmdSparkCreateClusterOptions struct {
		*CmdSparkCreateOptions
		OceanClusterID    string
		ClusterName       string
		Region            string
		Tags              []string
		KubernetesVersion string
		KubeConfigPath    string
	}
)

const (
	defaultK8sVersion   = "1.21"
	spotSystemNamespace = "spot-system"
)

var (
	errClusterNotFound = errors.New("cluster not found")
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

func (x *CmdSparkCreateCluster) shouldCreateNewCluster() bool {
	// If we have an Ocean cluster ID we will do an import, otherwise we create a new cluster
	return x.opts.OceanClusterID == ""
}

func (x *CmdSparkCreateCluster) run(ctx context.Context) error {
	if x.shouldCreateNewCluster() {
		if x.opts.ClusterName == "" {
			// Generate unique name
			x.opts.ClusterName = fmt.Sprintf("ocean-spark-cluster-%s", uuid.NewV4().Short())
		}

		// Note that controllerClusterID == cluster name in Ocean for Apache Spark
		ctrlClusterIDExists, err := x.doesControllerClusterIDExist(ctx, x.opts.ClusterName)
		if err != nil {
			return fmt.Errorf("could not check if controllerClusterID exists, %w", err)
		}

		if ctrlClusterIDExists {
			return fmt.Errorf("ocean cluster with controllerClusterID %q already exists", x.opts.ClusterName)
		}

		log.Infof("Will create Ocean for Apache Spark cluster %s", x.opts.ClusterName)
		if err := x.createEKSCluster(ctx); err != nil {
			return fmt.Errorf("could not create EKS cluster, %w", err)
		}

		createdCluster, err := x.getOceanClusterByControllerClusterID(ctx, x.opts.ClusterName)
		if err != nil {
			return fmt.Errorf("could not get ocean cluster by controller cluster ID, %w", err)
		}

		log.Infof("Successfully created Ocean cluster %s (%s)", createdCluster.ID, createdCluster.Name)
		x.opts.OceanClusterID = createdCluster.ID
	} else {
		log.Infof("Will deploy Ocean for Apache Spark on Ocean cluster %s", x.opts.OceanClusterID)

		controllerClusterID, err := x.getControllerClusterIDForOceanClusterID(ctx, x.opts.OceanClusterID)
		if err != nil {
			return fmt.Errorf("could not get controllerClusterID for cluster %s, %w", x.opts.OceanClusterID, err)
		}

		existingOceanSparkCluster, err := x.getOceanSparkClusterByControllerClusterID(ctx, controllerClusterID)
		if err != nil {
			if err != errClusterNotFound {
				return fmt.Errorf("could not check if Ocean Spark cluster already exists, %w", err)
			}
		} else {
			if existingOceanSparkCluster != nil {
				return fmt.Errorf("ocean spark cluster %s (%s) already exists on ocean cluster %s", existingOceanSparkCluster.ID, existingOceanSparkCluster.Name, x.opts.OceanClusterID)
			} else {
				return fmt.Errorf("ocean spark cluster already exists")
			}
		}

		// Note that controllerClusterID == cluster name in Ocean for Apache Spark
		x.opts.ClusterName = controllerClusterID
	}

	kubeConfig, err := kubernetes.GetConfig(x.opts.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("could not get kubeconfig, %w", err)
	}

	client, err := kubernetes.GetClient(kubeConfig)
	if err != nil {
		return fmt.Errorf("could not get kubernetes client, %w", err)
	}

	if err := ofas.ValidateClusterContext(ctx, client, x.opts.ClusterName); err != nil {
		return fmt.Errorf("cluster context validation failure, make sure your kubeconfig has the target cluster in context, %w", err)
	}

	log.Infof("Verified cluster %s", x.opts.ClusterName)

	log.Infof("Creating namespace %s", spotSystemNamespace)
	if err := kubernetes.EnsureNamespace(ctx, client, spotSystemNamespace); err != nil {
		return fmt.Errorf("could not create namespace, %w", err)
	}

	log.Infof("Creating deployer RBAC")
	if err := ofas.CreateDeployerRBAC(ctx, client, spotSystemNamespace); err != nil {
		return fmt.Errorf("could not create deployer rbac, %w", err)
	}

	log.Infof("Creating Ocean Spark cluster")

	oceanSparkClient, err := x.getOceanSparkClient()
	if err != nil {
		return fmt.Errorf("could not create Ocean Spark client, %w", err)
	}

	createdCluster, err := oceanSparkClient.CreateCluster(ctx, x.opts.OceanClusterID)
	if err != nil {
		return fmt.Errorf("could not create Ocean Spark cluster, %w", err)
	}

	log.Infof("Successfully created Ocean Spark cluster %s (%s)", createdCluster.ID, createdCluster.Name)

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
	fs.StringVar(&x.OceanClusterID, flags.FlagOFASOceanClusterID, x.OceanClusterID, "ID of Ocean cluster that should be imported into Ocean for Apache Spark. Note that your machine must be configured to access the cluster.")
	fs.StringVar(&x.ClusterName, flags.FlagOFASClusterName, x.ClusterName, "name of cluster that will be created (will be generated if empty)")
	fs.StringVar(&x.Region, flags.FlagOFASClusterRegion, os.Getenv("AWS_REGION"), "region in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Tags, "tags", x.Tags, "list of K/V pairs used to tag all cloud resources that will be created (eg: \"Owner=john@example.com,Team=DevOps\")")
	fs.StringVar(&x.KubernetesVersion, "kubernetes-version", defaultK8sVersion, "kubernetes version of cluster that will be created")
	fs.StringVar(&x.KubeConfigPath, flags.FlagOFASKubeConfigPath, kubernetes.GetDefaultKubeConfigPath(), "path to local kubeconfig")
}

func (x *CmdSparkCreateClusterOptions) Validate() error {
	if x.OceanClusterID != "" && x.ClusterName != "" {
		return spotctlerrors.RequiredXor(flags.FlagOFASOceanClusterID, flags.FlagOFASClusterName)
	}
	if x.OceanClusterID == "" && x.Region == "" {
		return spotctlerrors.Required(flags.FlagOFASClusterRegion)
	}
	if x.KubeConfigPath == "" {
		return spotctlerrors.Required(flags.FlagOFASKubeConfigPath)
	}
	return x.CmdSparkCreateOptions.Validate()
}

func (x *CmdSparkCreateCluster) installDeps(ctx context.Context) error {
	if !x.shouldCreateNewCluster() {
		// Nothing to install
		return nil
	}

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

	if err := dm.Install(ctx, dep.DependencyEksctlSpot, installOpts...); err != nil {
		return fmt.Errorf("could not install required dependency, %w", err)
	}

	present, err := dm.DependencyPresent(dep.DependencyEksctlSpot, installOpts...)
	if err != nil {
		return fmt.Errorf("could not validate required dependency, %w", err)
	}

	if !present {
		return fmt.Errorf("required dependency %s missing", dep.DependencyEksctlSpot.Name())
	}

	return nil
}

func (x *CmdSparkCreateCluster) doesControllerClusterIDExist(ctx context.Context, controllerClusterID string) (bool, error) {
	_, err := x.getOceanClusterByControllerClusterID(ctx, controllerClusterID)
	if err == nil {
		return true, nil
	}
	if err == errClusterNotFound {
		return false, nil
	}
	return false, err
}

func (x *CmdSparkCreateCluster) createEKSCluster(ctx context.Context) error {
	cloudProviderOpts := []cloud.ProviderOption{
		cloud.WithProfile(x.opts.Profile),
		cloud.WithRegion(x.opts.Region),
	}

	cloudProvider, err := x.opts.Clientset.NewCloud(cloud.ProviderName(x.opts.CloudProvider), cloudProviderOpts...)
	if err != nil {
		return fmt.Errorf("could not get cloud provider, %w", err)
	}

	cmdEksctl, err := x.opts.Clientset.NewCommand(eksctl.CommandName)
	if err != nil {
		return fmt.Errorf("could not create eksctl command, %w", err)
	}

	stacks, err := eks.GetStacksForCluster(cloudProvider, x.opts.Profile, x.opts.Region, x.opts.ClusterName)
	if err != nil {
		return fmt.Errorf("could not get stacks for cluster, %w", err)
	}

	clusterAlreadyExists := false
	if _, err := eks.GetEKSCluster(cloudProvider, x.opts.Profile, x.opts.Region, x.opts.ClusterName); err != nil {
		if !errors.As(err, &eks.ErrClusterNotFound{}) {
			return fmt.Errorf("could not check for existing EKS cluster, %w", err)
		}
	} else {
		clusterAlreadyExists = true
	}

	// TODO Allow creation of resources if previous stacks failed

	clusterStacks := eks.FilterStacks(stacks, eks.IsClusterStack)
	// Only create cluster if we don't have any cluster stacks, and it doesn't exist already
	shouldCreateCluster := len(clusterStacks) == 0 && !clusterAlreadyExists
	if !shouldCreateCluster {
		if len(clusterStacks) > 0 {
			log.Infof("Found cluster stacks, will not create cluster:\n%s", strings.Join(eks.StacksToStrings(clusterStacks), "\n"))
		} else if clusterAlreadyExists {
			log.Infof("EKS cluster %s already exists, will not create cluster", x.opts.ClusterName)
		} else {
			log.Infof("Will not create EKS cluster")
		}
	}

	if shouldCreateCluster {
		spinner := startSpinnerWithMessage(fmt.Sprintf("Creating EKS cluster %s", x.opts.ClusterName))
		createClusterArgs := x.buildEksctlCreateClusterArgs()
		if err := cmdEksctl.Run(ctx, createClusterArgs...); err != nil {
			stopSpinnerWithMessage(spinner, "Could not create EKS cluster", true)
			log.Infof("To see more log output, run spotctl with the --verbose flag")
			return fmt.Errorf("could not create EKS cluster, %w", err)
		}
		stopSpinnerWithMessage(spinner, "EKS cluster created", false)
	}

	nodegroupStacks := eks.FilterStacks(stacks, eks.IsNodegroupStack)
	createdClusterStacks := eks.FilterStacks(clusterStacks, eks.IsStackCreated)
	// Only create nodegroup if we don't have any nodegroup stacks, and if we just created the cluster or if it was created previously (via eksctl (cloudformation stacks))
	// Note that we cannot add a nodegroup using eksctl unless the cluster was created by (and therefore managed by) eksctl.
	// To check if a cluster is managed by eksctl, eksctl lists cloudformation stacks and checks for the "alpha.eksctl.io/cluster-name" tag
	// Therefore, if we have no cluster stacks at all, we know it was not created by eksctl
	shouldCreateNodegroup := len(nodegroupStacks) == 0 && (shouldCreateCluster || len(createdClusterStacks) > 0)
	if !shouldCreateNodegroup {
		if len(nodegroupStacks) > 0 {
			log.Infof("Found nodegroup stacks, will not create nodegroup:\n%s", strings.Join(eks.StacksToStrings(nodegroupStacks), "\n"))
		} else {
			log.Infof("Will not create nodegroup")
		}
	}

	if shouldCreateNodegroup {
		spinner := startSpinnerWithMessage("Creating Ocean node group")
		createNodeGroupArgs := x.buildEksctlCreateNodeGroupArgs()
		if err := cmdEksctl.Run(ctx, createNodeGroupArgs...); err != nil {
			stopSpinnerWithMessage(spinner, "Could not create node group", true)
			log.Infof("To see more log output, run spotctl with the --verbose flag")
			return fmt.Errorf("could not create node group, %w", err)
		}
		stopSpinnerWithMessage(spinner, "Spot Ocean node group created", false)
	}

	return nil
}

func (x *CmdSparkCreateCluster) getSpotClient() (spot.Client, error) {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	return x.opts.Clientset.NewSpotClient(spotClientOpts...)
}

func (x *CmdSparkCreateCluster) getOceanSparkClient() (spot.OceanSparkInterface, error) {
	spotClient, err := x.getSpotClient()
	if err != nil {
		return nil, fmt.Errorf("could not get spot client, %w", err)
	}

	return spotClient.Services().OceanSpark()
}

func (x *CmdSparkCreateCluster) getOceanClient() (spot.OceanInterface, error) {
	spotClient, err := x.getSpotClient()
	if err != nil {
		return nil, fmt.Errorf("could not get spot client, %w", err)
	}

	return spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
}

func (x *CmdSparkCreateCluster) getOceanClusterByID(ctx context.Context, id string) (*spot.OceanCluster, error) {
	oceanClient, err := x.getOceanClient()
	if err != nil {
		return nil, fmt.Errorf("could not get ocean client, %w", err)
	}

	return oceanClient.GetCluster(ctx, id)
}

func (x *CmdSparkCreateCluster) getControllerClusterIDForOceanClusterID(ctx context.Context, clusterID string) (string, error) {
	oceanCluster, err := x.getOceanClusterByID(ctx, clusterID)
	if err != nil {
		return "", fmt.Errorf("could not get ocean cluster, %w", err)
	}

	awsOceanCluster, ok := oceanCluster.Obj.(*oceanaws.Cluster)
	if !ok || awsOceanCluster == nil {
		return "", fmt.Errorf("could not get aws ocean cluster for cluster %q", clusterID)
	}

	if awsOceanCluster.ControllerClusterID == nil {
		return "", fmt.Errorf("controllerClusterID for cluster %q is nil", clusterID)
	}

	return *awsOceanCluster.ControllerClusterID, nil
}

func (x *CmdSparkCreateCluster) getOceanSparkClusterByControllerClusterID(ctx context.Context, controllerClusterID string) (*spot.OceanSparkCluster, error) {
	oceanSparkClient, err := x.getOceanSparkClient()
	if err != nil {
		return nil, fmt.Errorf("could not get ocean spark client, %w", err)
	}

	clusters, err := oceanSparkClient.ListClusters(ctx, controllerClusterID, "")
	if err != nil {
		return nil, fmt.Errorf("could not list ocean spark clusters, %w", err)
	}

	switch len(clusters) {
	case 0:
		return nil, errClusterNotFound
	case 1:
		return clusters[0], nil
	default:
		return nil, fmt.Errorf("found %d ocean spark clusters with controllerClusterID %q, expected at most 1", len(clusters), controllerClusterID)
	}
}

// getOceanClusterByControllerClusterID finds Ocean cluster with the given controllerClusterID.
// Returns errClusterNotFound if it is not found
// Returns error if multiple clusters found with the given controllerClusterID
func (x *CmdSparkCreateCluster) getOceanClusterByControllerClusterID(ctx context.Context, controllerClusterID string) (*spot.OceanCluster, error) {
	oceanClient, err := x.getOceanClient()
	if err != nil {
		return nil, fmt.Errorf("could not get ocean client, %w", err)
	}

	clusters, err := oceanClient.ListClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list ocean clusters, %w", err)
	}

	foundClusters := make([]*spot.OceanCluster, 0)
	for i := range clusters {
		clusterID := clusters[i].ID

		awsOceanCluster, ok := clusters[i].Obj.(*oceanaws.Cluster)
		if !ok || awsOceanCluster == nil {
			log.Warnf("Could not cast Ocean cluster object for cluster: %s", clusterID)
			continue
		}

		if awsOceanCluster.ControllerClusterID == nil {
			log.Warnf("Got nil controllerClusterID for Ocean cluster: %s", clusterID)
			continue
		}

		if *awsOceanCluster.ControllerClusterID == controllerClusterID {
			foundClusters = append(foundClusters, clusters[i])
		}
	}

	switch len(foundClusters) {
	case 0:
		return nil, errClusterNotFound
	case 1:
		return foundClusters[0], nil
	default:
		return nil, fmt.Errorf("found %d ocean clusters with controllerClusterID %q, expected at most 1", len(foundClusters), controllerClusterID)
	}
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
		args = append(args, "--verbose", "1")
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
			"--name", fmt.Sprintf("ocean-spark-bootstrap-%s", uuid.NewV4().Short()))
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

	args = append(args, "--managed=false") // Not EKS managed
	args = append(args, "--spot-ocean")

	if x.opts.Verbose {
		args = append(args, "--verbose", "4")
	} else {
		args = append(args, "--verbose", "1")
	}

	return args
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
