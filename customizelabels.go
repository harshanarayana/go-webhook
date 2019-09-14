package main

import (
	"bytes"
	"errors"
	klog "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"text/template"
)

const (
	SecretLabel   = "you.shall.not.pass.io"
	AddFirstLabel = `[
		{ "op":  "add", "path": "/spec/template/metadata/labels", "value": { "{{ .LabelName }}": "{{ .LabelValue }}" }},
		{ "op":  "add", "path": "/spec/selector/matchLabels", "value": { "{{ .LabelName }}": "{{ .LabelValue }}" }}
	]`
	AddLabel = `[
		{ "op":  "add", "path": "/spec/template/metadata/labels/{{ .LabelName }}", "value": "{{ .LabelValue }}" },
		{ "op":  "add", "path": "/spec/selector/matchLabels/{{ .LabelName }}", "value": "{{ .LabelValue }}" }
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

	if _, ok := pod.Labels[SecretLabel]; !ok {
		secret, err := GenerateRandomString(10)
		if err != nil {
			klog.Fatalf("Failed to generate random secret %v", err)
			return CreateAdmissionResponse("404", errors.New("Failed to generate a random secret "))
		} else {
			if len(pod.Labels) == 0 {
				reviewResponse.Patch = []byte(renderTemplate(AddFirstLabel, map[string]interface{}{
					"LabelName":  SecretLabel,
					"LabelValue": secret,
				}))
				klog.WithFields(klog.Fields{
					"event": string(reviewResponse.Patch),
				}).Info("Response Patch")
			} else {
				reviewResponse.Patch = []byte(renderTemplate(AddLabel, map[string]interface{}{
					"LabelName":  SecretLabel,
					"LabelValue": secret,
				}))
				klog.WithFields(klog.Fields{
					"event": string(reviewResponse.Patch),
				}).Info("Response Patch")
			}
			pt := v1beta1.PatchTypeJSONPatch
			reviewResponse.PatchType = &pt
		}
	}
	klog.WithFields(klog.Fields{
		"response": reviewResponse.Result,
	}).Info("Response Patch")
	return &reviewResponse
}

func MutatingLabelHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, addSecretLabel)
}
