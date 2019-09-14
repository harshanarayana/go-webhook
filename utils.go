package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func TLSConfig(config Config) *tls.Config {
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

func CreateAdmissionResponse(status string, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Status:  status,
			Message: err.Error(),
		},
	}
}

// xref: credits: https://flaviocopes.com/go-random/
func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	b, err := randomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
