{{/*
TreePage frontend chart helpers.
*/}}
{{- define "treepage-frontend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "treepage-frontend.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "treepage-frontend.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "treepage-frontend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "treepage-frontend.labels" -}}
helm.sh/chart: {{ include "treepage-frontend.chart" . }}
app.kubernetes.io/name: {{ include "treepage-frontend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Values.global.appVersion | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: treepage
app.kubernetes.io/component: frontend
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "treepage-frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "treepage-frontend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: frontend
{{- end -}}

{{- define "treepage-frontend.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repo := .Values.frontend.image.repository -}}
{{- $tag := .Values.frontend.image.tag | default .Values.global.appVersion | default "latest" -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end -}}

{{- define "treepage-frontend.publicUrl" -}}
{{- if .Values.global.frontendUrl -}}
{{- .Values.global.frontendUrl -}}
{{- else if .Values.ingress.host -}}
{{- if .Values.ingress.tls.enabled -}}
{{- printf "https://%s" .Values.ingress.host -}}
{{- else -}}
{{- printf "http://%s" .Values.ingress.host -}}
{{- end -}}
{{- else -}}
{{- "http://localhost" -}}
{{- end -}}
{{- end -}}

{{- define "treepage-frontend.apiUrl" -}}
{{- printf "%s/api" (include "treepage-frontend.publicUrl" .) -}}
{{- end -}}

{{- define "treepage-frontend.authUrl" -}}
{{- printf "%s/api/auth" (include "treepage-frontend.publicUrl" .) -}}
{{- end -}}

{{- define "treepage-frontend.backendAuthService" -}}
{{- .Values.backend.authService | default (printf "%s-auth" .Values.backend.releaseName) -}}
{{- end -}}

{{- define "treepage-frontend.backendServerService" -}}
{{- .Values.backend.serverService | default (printf "%s-server" .Values.backend.releaseName) -}}
{{- end -}}
