package config

import (
	"fmt"
	"io"
	"net/http"
)

const OceanSparkDeployYamlUrl = "https://spotinst-public.s3.amazonaws.com/integrations/kubernetes/ocean-spark/templates/ocean-spark-deploy.yaml"

func GetDeploymentYaml() (string, error) {
	resp, err := http.Get(OceanSparkDeployYamlUrl)
	if err != nil {
		return "", fmt.Errorf("error fetching the ocean-spark-deploy.yaml from %s: %w", OceanSparkDeployYamlUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("could not read ocean-spark-deploy.yaml from %s: unexpected status code %d: %s", OceanSparkDeployYamlUrl, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body for ocean-spark-deploy.yaml from %s: %w", OceanSparkDeployYamlUrl, err)
	}

	return string(data), nil
}
