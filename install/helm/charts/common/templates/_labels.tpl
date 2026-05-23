{{/*
Common labels
*/}}
{{- define "common.labels" -}}
helm.sh/chart: {{ include "common.chart" . }}
{{ include "common.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "common.selectorLabels" -}}
{{- with .Values.selectorLabels }}
{{- tpl (toYaml .) . }}
{{- else }}
app.kubernetes.io/name: {{ include "common.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- with .Values.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Pod labels
*/}}
{{- define "common.podLabels" -}}
{{- include "common.labels" . }}
{{- with .Values.podLabels }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end }}

{{/*
Merge labels with .Values.labels
Usage:
{{ include "common.mergeLabels" ( dict "value" .Values.path.to.the.Value1 "context" $ ) }}
*/}}
{{- define "common.mergeLabels" -}}
{{- $labels := include "common.labels" .context | fromYaml -}}
{{- $labels = .value | default dict | merge $labels -}}
{{- with $labels }}
{{- tpl (toYaml .) $ }}
{{- end }}
{{- end }}
