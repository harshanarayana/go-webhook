package main

import (
	"bytes"
	"errors"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/http"
	"text/template"
)

const (
	SECRET_LABEL    = "you.shall.not.pass.io"
	ADD_FIRST_LABEL = `[
		{ "op":  "add", "path": "/spec/template/metadata/labels", "value": { "{{ .LabelName }}": "{{ .LabelValue }}" }
	]`
	ADD_LABEL = `[
		{ "op":  "add", "path": "/spec/template/metadata/labels/{{ .LabelName }}", "value": "{{ .LabelValue }}" }
	]`
)

func renderTemplate(tpl string, data interface{}) string {
	t := template.Must(template.New("patch").Parse(tpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

func addSecretLabel(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	klog.Info("Mutating the request to append custom labels for secret if missing")
	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return CreateAdmissionResponse("404", err)
	}
	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if _, ok := pod.Labels[SECRET_LABEL]; !ok {
		secret, err := GenerateRandomString(10)
		if err != nil {
			klog.Fatal("Failed to generate random secret %v", err)
			return CreateAdmissionResponse("404", errors.New("Failed to generate a random secret "))
		} else {
			if len(pod.Labels) == 0 {
				reviewResponse.Patch = []byte(renderTemplate(ADD_FIRST_LABEL, map[string]interface{}{
					"LabelName":  SECRET_LABEL,
					"LabelValue": secret,
				}))
			} else {
				reviewResponse.Patch = []byte(renderTemplate(ADD_LABEL, map[string]interface{}{
					"LabelName":  SECRET_LABEL,
					"LabelValue": secret,
				}))
			}
			pt := v1beta1.PatchTypeJSONPatch
			reviewResponse.PatchType = &pt
		}
	}
	return &reviewResponse
}

func MutatingLabelHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, addSecretLabel)
}
