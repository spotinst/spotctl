package spot

import (
	"fmt"

	"github.com/spotinst/spotinst-sdk-go/service/ocean"
	waveService "github.com/spotinst/spotinst-sdk-go/service/wave"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

type apiServices struct {
	session *session.Session
}

func (x *apiServices) Ocean(provider CloudProviderName, orchestrator OrchestratorName) (OceanInterface, error) {
	switch provider {
	case CloudProviderAWS:
		return x.oceanAWS(orchestrator)
	case CloudProviderGCP:
		return x.oceanGCP(orchestrator)
	default:
		return nil, fmt.Errorf("spot: unsupported cloud provider: %s", provider)
	}
}

func (x *apiServices) oceanAWS(orchestrator OrchestratorName) (OceanInterface, error) {
	svc := ocean.New(x.session).CloudProviderAWS()

	switch orchestrator {
	case OrchestratorKubernetes:
		return &oceanKubernetesAWS{svc}, nil
	case OrchestratorECS:
		return &oceanECS{svc}, nil
	default:
		return nil, fmt.Errorf("spot: unsupported orchestrator: %s", orchestrator)
	}
}

func (x *apiServices) oceanGCP(orchestrator OrchestratorName) (OceanInterface, error) {
	svc := ocean.New(x.session).CloudProviderGCP()

	switch orchestrator {
	case OrchestratorKubernetes:
		return &oceanKubernetesGCP{svc}, nil
	default:
		return nil, fmt.Errorf("spot: unsupported orchestrator: %s", orchestrator)
	}
}

func (x *apiServices) Wave() (WaveInterface, error) {
	svc := waveService.New(x.session)
	return &wave{svc}, nil
}
