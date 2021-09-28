package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRBACManifests(t *testing.T) {

	var expectedServiceAccount = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: bigdata-deployer
  namespace: my-namespace
`

	var expectedRoleBinding = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: bigdata-deployer-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: bigdata-deployer
    namespace: my-namespace
`

	t.Run("whenSuccessful", func(tt *testing.T) {

		res, err := getRBACManifests("my-namespace")
		assert.NoError(tt, err)

		assert.Equal(tt, expectedServiceAccount, res.serviceAccount)
		assert.Equal(tt, expectedRoleBinding, res.clusterRoleBinding)
	})
}
