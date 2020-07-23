package cert

import (
	"encoding/base64"
	"fmt"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/pkg/pki"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

func SendCert(c *rundata.Cluster) error {
	return c.RunOnMasters(func(node *rundata.Node) error {
		return sendNodeCert(node)
	})

}

func sendNodeCert(node *rundata.Node) error {

	certTree := node.CertificateTree

	for ca, certs := range certTree {
		if err := sendCert(node, ca); err != nil {
			return err
		}

		for _, cert := range certs {
			if err := sendCert(node, cert); err != nil {
				return err
			}

			if cert.IsKubeConfig {
				if err := sendKubeConfig(node, cert); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func sendCert(node *rundata.Node, c *rundata.Cert) error {
	// TODO mddir path for cert
	// send cert
	encodeCert := pki.EncodeCertPEM(c.Cert)
	encodeCertBase64 := base64.StdEncoding.EncodeToString(encodeCert)
	if err := node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/pki/%s.crt", encodeCertBase64, c.BaseName)); err != nil {
		return err
	}

	// send key
	encodedKey, err := pki.EncodePrivateKeyPEM(c.Key)
	if err != nil {
		return err
	}
	encodedKeyBase64 := base64.StdEncoding.EncodeToString(encodedKey)
	return node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/pki/%s.key", encodedKeyBase64, c.BaseName))
}

func sendKubeConfig(node *rundata.Node, c *rundata.Cert) error {
	encodedKubeConfig, err := EncodeKubeConfig(c.KubeConfig)
	if err != nil {
		return err
	}
	encodedKubeConfigBase64 := base64.StdEncoding.EncodeToString(encodedKubeConfig)

	if c.Name == "admin" {
		if err := node.Run(fmt.Sprintf("mkdir $HOME/.kube && echo %s | base64 -d > $HOME/.kube/config", encodedKubeConfigBase64)); err != nil {
			return err
		}
	}

	return node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/%s", encodedKubeConfigBase64, c.BaseName))
}

// EncodeKubeConfig serializes the config to yaml.
// Encapsulates serialization without assuming the destination is a file.
func EncodeKubeConfig(config clientcmdapi.Config) ([]byte, error) {
	return runtime.Encode(clientcmdlatest.Codec, &config)
}
