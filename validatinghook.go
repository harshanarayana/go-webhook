package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
)

const (
	FilePath = "/config/exec-options.json"
)

func loadConfig() ExecOptionConfiguration  {
	var option ExecOptionConfiguration
	b, err := ioutil.ReadFile(FilePath)
	if err != nil {
		return option
	}
	_ = json.Unmarshal(b, &option)
	return option
}

func validateContainerExec(ar  v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	logrus.Infof("Validating pod exec based on pre-configured annotations")
	logrus.WithFields(logrus.Fields{
		"operation": ar.Request.Operation,
		"subresource": ar.Request.SubResource,
		"raw": string(ar.Request.Object.Raw),
	}).Infof("Request Information")
	if ar.Request.Operation != v1beta1.Connect {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if ar.Request.SubResource !=  "exec" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	raw := ar.Request.Object.Raw
	podExecOptions := corev1.PodExecOptions{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &podExecOptions); err != nil {
		logrus.Errorf("failed to perform validation of container exec %v", err)
		return CreateAdmissionResponse("404", err)
	}
	logrus.Infof("Exec options %v", podExecOptions)
	var config = loadConfig()
	if c, ok := config.ExecOptions[podExecOptions.Container]; ok {
		switch podExecOptions.Command[0] {
		case "bash":
			if c.Bash {
				return &v1beta1.AdmissionResponse{Allowed: true}
			} else {
				return CreateAdmissionResponseWithAllowance(false, "404",
					fmt.Errorf("exec into %s via %s in not allowed", podExecOptions.Container, "bash"))
			}
		case "sh":
			if c.Sh {
				return &v1beta1.AdmissionResponse{Allowed: true}
			} else {
				return CreateAdmissionResponseWithAllowance(false, "404",
					fmt.Errorf("exec into %s via %s in not allowed", podExecOptions.Container, "sh"))
			}
		default:
			return &v1beta1.AdmissionResponse{Allowed: true}
		}
	} else {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
}

func ValidatingExecHandler(w http.ResponseWriter, r *http.Request)  {
	serve(w, r, validateContainerExec)
}
