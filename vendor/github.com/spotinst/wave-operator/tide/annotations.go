package tide

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/install"
)

const AnnotationPrefix = "tide.wave.spot.io"

type Upgrade struct {
	UpgradedAt time.Time           `json:"upgradedAt"`
	Spec       install.InstallSpec `json:"spec"`
}

func addUpgradeAnnotation(spec install.InstallSpec, env *v1alpha1.WaveEnvironment) error {
	upgrades := []Upgrade{}
	u := Upgrade{
		UpgradedAt: time.Now(),
		Spec:       spec,
	}
	key := AnnotationPrefix + "/waveUpgrades"
	annotation := env.Annotations[key]
	if annotation != "" {
		err := json.Unmarshal([]byte(annotation), &upgrades)
		if err != nil {
			// ? perhaps better to overwrite it
			return fmt.Errorf("unable to add upgrade to existing annotation, %w", err)
		}
	}
	upgrades = append(upgrades, u)
	newAnnotation, err := json.Marshal(upgrades)
	if err != nil {
		return fmt.Errorf("unable to add upgrade to existing annotation, %w", err)
	}
	env.Annotations[key] = string(newAnnotation)
	return nil
}
