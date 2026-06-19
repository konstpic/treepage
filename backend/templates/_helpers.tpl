{{/*
TreePage backend chart helpers.
Compatible naming with legacy uc-chart releases: *-auth, *-server, *-sync.
*/}}
{{- define "treepage-backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "treepage-backend.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "treepage-backend.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "treepage-backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "treepage-backend.labels" -}}
helm.sh/chart: {{ include "treepage-backend.chart" . }}
app.kubernetes.io/name: {{ include "treepage-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Values.global.appVersion | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: treepage
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "treepage-backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "treepage-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "treepage-backend.componentSelectorLabels" -}}
{{ include "treepage-backend.selectorLabels" . }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "treepage-backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "treepage-backend.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "treepage-backend.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repo := .repository -}}
{{- $tag := .tag | default .Values.global.appVersion | default "latest" -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end -}}

{{- define "treepage-backend.secretName" -}}
{{- if .Values.secret.existingSecret -}}
{{- .Values.secret.existingSecret -}}
{{- else -}}
{{- include "treepage-backend.fullname" . -}}-credentials
{{- end -}}
{{- end -}}

{{- define "treepage-backend.postgresHost" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "%s-postgresql" (include "treepage-backend.fullname" .) -}}
{{- else -}}
{{- .Values.postgresql.host -}}
{{- end -}}
{{- end -}}

{{- define "treepage-backend.frontendUrl" -}}
{{- if .Values.global.frontendUrl -}}
{{- .Values.global.frontendUrl -}}
{{- else if .Values.ingress.host -}}
{{- if .Values.ingress.tls.enabled -}}
{{- printf "https://%s" .Values.ingress.host -}}
{{- else -}}
{{- printf "http://%s" .Values.ingress.host -}}
{{- end -}}
{{- else -}}
{{- printf "http://%s-frontend" (include "treepage-backend.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "treepage-backend.syncInternalUrl" -}}
{{- printf "http://%s-sync:%d" (include "treepage-backend.fullname" .) (.Values.sync.service.port | int) -}}
{{- end -}}
