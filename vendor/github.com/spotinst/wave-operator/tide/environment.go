package tide

import (
	"context"
	"fmt"

	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/internal/version"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *manager) CreateWaveEnvironment(
	ctx context.Context,
	name, namespace string,
	certManagerDeployed, k8sClusterProvisioned, oceanClusterProvisioned bool,
) error {
	env := &v1alpha1.WaveEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.WaveEnvironmentSpec{
			OperatorVersion:         version.BuildVersion,
			CertManagerDeployed:     certManagerDeployed,
			K8sClusterProvisioned:   k8sClusterProvisioned,
			OceanClusterProvisioned: oceanClusterProvisioned,
		},
	}
	rc, err := m.getControllerRuntimeClient()
	if err != nil {
		return fmt.Errorf("kubernetes config error, %w", err)
	}

	err = rc.Create(ctx, env)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			m.log.Info("wave environment already exists", "name", env.Name)
		} else {
			return fmt.Errorf("cannot create environment %s, %w", env.Name, err)
		}
	}
	return nil
}
