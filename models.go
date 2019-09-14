package main

import "k8s.io/api/admission/v1beta1"

type Config struct {
	CertFile string
	KeyFile  string
}

type Admission func(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

type AllowedShell struct {
	Bash bool `json:"bash"`
	Sh   bool `json:"sh"`
}

type ExecOptionConfiguration struct {
	ExecOptions map[string]AllowedShell `json:"execOptions"`
}
