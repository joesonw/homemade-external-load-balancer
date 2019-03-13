package proxy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type sslCert struct {
	Cert []byte
	Key  []byte
}

func (s *Server) fetchSSLSecret(name, namespace string) (*sslCert, error) {
	secret, err := s.clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if secret.Type != corev1.SecretTypeTLS {
		return nil, fmt.Errorf("unsupportede secret type for ssl cert")
	}

	return &sslCert{
		Cert: secret.Data[corev1.TLSCertKey],
		Key:  secret.Data[corev1.TLSPrivateKeyKey],
	}, nil
}
