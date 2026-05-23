{{/*
Common annotations
*/}}
{{- define "common.annotations" -}}
{{- with .Values.annotations }}
{{- tpl (toYaml .) $ }}
{{- end }}
{{- end }}

{{/*
Pod annotations
*/}}
{{- define "common.podAnnotations" -}}
{{- include "common.annotations" . }}
{{- with .Values.podAnnotations }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- if eq (include "common.metrics.enabled" .) "true" }}
{{- with (.Values.metrics).podAnnotations }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Merge annotations with .Values.annotations
Usage:
{{ include "common.mergeAnnotations" ( dict "value" .Values.path.to.the.Value1 "context" $ ) }}
*/}}
{{- define "common.mergeAnnotations" -}}
{{- $annotations := include "common.annotations" .context | fromYaml -}}
{{- $annotations = .value | default dict | merge $annotations -}}
{{- with $annotations }}
{{- tpl (toYaml .) $ }}
{{- end }}
{{- end }}
