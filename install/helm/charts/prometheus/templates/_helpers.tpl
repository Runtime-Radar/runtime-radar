{{/*
Return the proper prometheus image name
*/}}
{{- define "prometheus.image" -}}
{{ include "common.image" (dict "context" . "image" .Values.image) }}
{{- end -}}
