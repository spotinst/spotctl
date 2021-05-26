package tide

import "github.com/spotinst/wave-operator/api/v1alpha1"

type FakeEnvironment struct {
	Env v1alpha1.WaveEnvironment
}

func (f *FakeEnvironment) GetConfiguration() (*v1alpha1.WaveEnvironment, error) {
	return &f.Env, nil
}

func (f *FakeEnvironment) SaveConfiguration(env *v1alpha1.WaveEnvironment) error {
	env.DeepCopyInto(&f.Env)
	return nil
}
