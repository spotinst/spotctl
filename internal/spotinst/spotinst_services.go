package spotinst

import (
	"fmt"

	"github.com/spotinst/spotinst-sdk-go/service/ocean"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

type apiServices struct {
	session *session.Session
}

func (x *apiServices) Ocean(provider CloudProviderName, orchestrator OrchestratorName) (OceanInterface, error) {
	switch provider {
	case CloudProviderAWS:
		return x.oceanAWS(orchestrator)
	default:
		return nil, fmt.Errorf("spotinst: unsupported cloud provider: %s", provider)
	}
}

func (x *apiServices) oceanAWS(orchestrator OrchestratorName) (OceanInterface, error) {
	switch orchestrator {
	case OrchestratorKubernetes:
		return &oceanKubernetesAWS{
			svc: ocean.New(x.session).CloudProviderAWS(),
		}, nil
	default:
		return nil, fmt.Errorf("spotinst: unsupported orchestrator: %s", orchestrator)
	}
}
