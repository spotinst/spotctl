module github.com/spotinst/spotctl

go 1.12

require (
	github.com/AlecAivazis/survey/v2 v2.0.2
	github.com/Netflix/go-expect v0.0.0-20180814212900-124a37274874 // indirect
	github.com/aws/aws-sdk-go v1.27.0
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.51.0
	github.com/go-logr/logr v0.3.0
	github.com/golang/groupcache v0.0.0-20191227052852-215e87163ea7 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-version v1.2.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/riywo/loginshell v0.0.0-20190610082906-2ed199a032f6
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spotinst/spotinst-sdk-go v1.66.0
	github.com/spotinst/wave-operator v0.0.0-20201102154306-c6fccf1c60ef
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	helm.sh/helm/v3 v3.3.4
	k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v0.18.8
	k8s.io/gengo v0.0.0-20200413195148-3a45101e95ac // indirect
	k8s.io/klog/v2 v2.2.0 // indirect
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/structured-merge-diff/v4 v4.0.1 // indirect
)

replace github.com/spotinst/wave-operator => github.com/spotinst/wave-operator v0.0.0-20201110225715-09bebc57514b
