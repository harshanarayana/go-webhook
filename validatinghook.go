package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"strings"
)

const (
	EnableExecCheckAnnotation = "maglev.cisco.com/exec-check"
)

func validateContainerExec(ar  v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	logrus.Infof("Validating pod exec based on pre-configured annotations")
	if ar.Request.Operation != v1beta1.Connect {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if ar.Request.SubResource !=  "exec" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	raw := ar.Request.Object.Raw
	podExecOptions := corev1.PodExecOptions{}
	pod := corev1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		logrus.Errorf("failed to perform validation of container exec %v", err)
		return CreateAdmissionResponse("404", err)
	}
	if _, _, err := deserializer.Decode(raw, nil, &podExecOptions); err != nil {
		logrus.Errorf("failed to perform validation of container exec %v", err)
		return CreateAdmissionResponse("404", err)
	}
	if _, ok := pod.Annotations[EnableExecCheckAnnotation]; !ok {
		return &v1beta1.AdmissionResponse{Allowed: true}
	} else {
		v, _ := pod.Annotations[EnableExecCheckAnnotation]
		var containers = make(map[string]bool, 0)
		for _, c := range strings.Split(v, ",") {
			containers[c] = true
		}
		if _, aok := containers[podExecOptions.Container]; aok {
			return &v1beta1.AdmissionResponse{Allowed: true}
		} else {
			return CreateAdmissionResponseWithAllowance(false, "404", fmt.Errorf(
				"exec only allowed on contaienrs %v", v))
		}
	}
}

func ValidatingExecHandler(w http.ResponseWriter, r *http.Request)  {
	serve(w, r, validateContainerExec)
}
