module github.com/spotinst/spotctl

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.3.2
	github.com/Netflix/go-expect v0.0.0-20180814212900-124a37274874 // indirect
	github.com/aws/aws-sdk-go v1.40.54
	github.com/docker/docker v20.10.3+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.63.2
	github.com/go-logr/logr v0.4.0
	github.com/google/go-containerregistry v0.5.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/mholt/archiver/v3 v3.5.0
	github.com/riywo/loginshell v0.0.0-20190610082906-2ed199a032f6
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spotinst/spotinst-sdk-go v1.102.0
	github.com/spotinst/wave-operator v0.0.0-20210524091717-f8934344b1f2
	github.com/theckman/yacspin v0.8.0
	k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	// https://github.com/helm/helm/issues/9354
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
