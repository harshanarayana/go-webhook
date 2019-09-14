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
	logrus.Infof("Annotations %v", pod.GetAnnotations())
	if _, ok := pod.GetAnnotations()[EnableExecCheckAnnotation]; !ok {
		logrus.Infof("Annotation %s not found in pod. No exec validation will be performed", EnableExecCheckAnnotation)
		return &v1beta1.AdmissionResponse{Allowed: true}
	} else {
		logrus.Infof("Performing exec validation based on annotation %s", EnableExecCheckAnnotation)
		v, _ := pod.GetAnnotations()[EnableExecCheckAnnotation]
		logrus.Infof("Containers allowed to be exec'ed into %s", v)
		var containers = make(map[string]bool, 0)
		for _, c := range strings.Split(v, ",") {
			containers[c] = true
		}
		if _, aok := containers[podExecOptions.Container]; aok {
			logrus.Infof("Container allowed to be exec'ed into")
			return &v1beta1.AdmissionResponse{Allowed: true}
		} else {
			logrus.Infof("denying exec into container as this is no enabled by the annotation %s", EnableExecCheckAnnotation)
			return CreateAdmissionResponseWithAllowance(false, "404", fmt.Errorf(
				"exec only allowed on contaienrs %v", v))
		}
	}
}

func ValidatingExecHandler(w http.ResponseWriter, r *http.Request)  {
	serve(w, r, validateContainerExec)
}
