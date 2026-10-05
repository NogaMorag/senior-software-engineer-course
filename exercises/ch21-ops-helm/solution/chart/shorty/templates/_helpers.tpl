{{/*
Named templates shared by every file in templates/. A name starting with the
chart name avoids collisions with subcharts.
*/}}

{{/* The release-qualified name, cut to 63 characters (the DNS label limit). */}}
{{- define "shorty.fullname" -}}
{{- if contains .Chart.Name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/* Labels that select the pods. They must never change after the first install. */}}
{{- define "shorty.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* Every object gets the selector labels plus version and owner labels. */}}
{{- define "shorty.labels" -}}
{{ include "shorty.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end -}}
