package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/log"
)

const (
	// Ocean.
	FlagOceanName                       string = "name"
	FlagOceanRegion                     string = "region"
	FlagOceanControllerID               string = "controller-id"
	FlagOceanClusterID                  string = "cluster-id"
	FlagOceanSpecID                     string = "spec-id"
	FlagOceanSpotPercentage             string = "spot-percentage"
	FlagOceanDrainingTimeout            string = "draining-timeout"
	FlagOceanUtilizeReserveInstances    string = "utilize-reserved-instances"
	FlagOceanFallbackOnDemand           string = "fallback-ondemand"
	FlagOceanMinSize                    string = "min-size"
	FlagOceanMaxSize                    string = "max-size"
	FlagOceanTargetSize                 string = "target-size"
	FlagOceanSubnetIDs                  string = "subnet-ids"
	FlagOceanInstancesTypesWhitelist    string = "instance-types-whitelist"
	FlagOceanInstancesTypesBlacklist    string = "instance-types-blacklist"
	FlagOceanSecurityGroupIDs           string = "security-group-ids"
	FlagOceanImageID                    string = "image-id"
	FlagOceanKeyPair                    string = "key-pair"
	FlagOceanUserData                   string = "user-data"
	FlagOceanRootVolumeSize             string = "root-volume-size"
	FlagOceanAssociatePublicIPAddress   string = "associate-public-ip-address"
	FlagOceanEnableMonitoring           string = "enable-monitoring"
	FlagOceanEnableEBSOptimization      string = "enable-ebs-optimization"
	FlagOceanIamInstanceProfileName     string = "iam-instance-profile-name"
	FlagOceanIamInstanceProfileARN      string = "iam-instance-profile-arn"
	FlagOceanLoadBalancerName           string = "load-balancer-name"
	FlagOceanLoadBalancerARN            string = "load-balancer-arn"
	FlagOceanLoadBalancerType           string = "load-balancer-type"
	FlagOceanEnableAutoScaler           string = "enable-auto-scaler"
	FlagOceanEnableAutoScalerAutoConfig string = "enable-auto-scaler-autoconfig"
	FlagOceanCooldown                   string = "cooldown"
	FlagOceanHeadroomCPUPerUnit         string = "headroom-cpu-per-unit"
	FlagOceanHeadroomMemoryPerUnit      string = "headroom-memory-per-unit"
	FlagOceanHeadroomGPUPerUnit         string = "headroom-gpu-per-unit"
	FlagOceanHeadroomNumPerUnit         string = "headroom-num-per-unit"
	FlagOceanResourceLimitMaxVCPU       string = "resource-limit-max-vcpu"
	FlagOceanResourceLimitMaxMemory     string = "resource-limit-max-memory"
	FlagOceanEvaluationPeriods          string = "evaluation-periods"
	FlagOceanMaxScaleDownPercentage     string = "max-scale-down-percentage"
	FlagOceanRolloutID                  string = "rollout-id"
	FlagOceanRolloutComment             string = "comment"
	FlagOceanRolloutBatchSizePercentage string = "batch-size-percentage"
	FlagOceanRolloutDisableAutoScaling  string = "disable-auto-scaling"
	FlagOceanRolloutSpecIDs             string = "spec-ids"
	FlagOceanRolloutInstanceIDs         string = "instance-ids"

	// Wave.
	FlagWaveRegion      string = "region"
	FlagWaveClusterID   string = "cluster-id"
	FlagWaveClusterName string = "cluster-name"
	FlagWaveConfigFile  string = "config-file"
	FlagWaveImage       string = "wave-image"
)

func Log(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		log.Debugf("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
