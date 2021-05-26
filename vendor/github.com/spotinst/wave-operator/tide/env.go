package tide

import (
	"github.com/spotinst/wave-operator/api/v1alpha1"
)

type Environment interface {
	EnvironmentGetter
	EnvironmentSaver
}

type EnvironmentGetter interface {
	GetConfiguration() (*v1alpha1.WaveEnvironment, error)
}

type EnvironmentSaver interface {
	SaveConfiguration(env *v1alpha1.WaveEnvironment) error
}
