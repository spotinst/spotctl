package tide

import (
	"context"
	"fmt"
	"time"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/spotinst/wave-operator/catalog"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func init() {
	_ = cm.AddToScheme(scheme)
}

func (m *manager) checkCertManagerPreinstallation() (bool, error) {
	ctx := context.TODO()
	config, err := m.kubeClientGetter.ToRESTConfig()
	if err != nil {
		return false, err
	}
	extClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return false, err
	}
	_, err = extClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "certificates.cert-manager.io", metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	_, err = extClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "issuers.cert-manager.io", metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return m.testCertManager()
}

func (m *manager) testCertManager() (bool, error) {
	ctx := context.TODO()
	rc, err := m.getControllerRuntimeClient()
	if err != nil {
		return false, fmt.Errorf("kubernetes config error, %w", err)
	}
	issuer := &cm.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-selfsigned",
			Namespace: catalog.SystemNamespace,
		},
		Spec: cm.IssuerSpec{
			IssuerConfig: cm.IssuerConfig{
				SelfSigned: &cm.SelfSignedIssuer{},
			},
		},
	}
	uiss := &unstructured.Unstructured{}
	if err := scheme.Convert(issuer, uiss, nil); err != nil {
		return false, err
	}
	err = rc.Create(ctx, uiss)
	if err != nil {
		return false, fmt.Errorf("error creating test case, %w", err)
	}
	defer rc.Delete(ctx, uiss)

	cert := &cm.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-selfsigned",
			Namespace: catalog.SystemNamespace,
		},
		Spec: cm.CertificateSpec{
			DNSNames:   []string{"wave.spot.io"},
			SecretName: "selfsigned-cert-tls",
			IssuerRef:  cmmeta.ObjectReference{Name: issuer.Name},
		},
	}
	uc := &unstructured.Unstructured{}
	if err := scheme.Convert(cert, uc, nil); err != nil {
		return false, err
	}
	err = rc.Create(ctx, uc)
	if err != nil {
		return false, fmt.Errorf("error creating test case, %w", err)
	}
	defer rc.Delete(ctx, uc)

	key := types.NamespacedName{
		Name:      cert.Name,
		Namespace: cert.Namespace,
	}
	err = wait.Poll(1*time.Second, 60*time.Second, func() (bool, error) {
		obj := &cm.Certificate{}
		err := rc.Get(ctx, key, obj)
		if err != nil {
			return false, err
		}
		for _, c := range obj.Status.Conditions {
			if c.Type == "Ready" && c.Status == cmmeta.ConditionTrue {
				m.log.Info("cert manager pre-check successful", "cert", obj.Name, "message", c.Message)
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return false, fmt.Errorf("Error in checking cert-manager functionality, %w", err)
	}
	return true, nil
}
