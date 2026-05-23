{{/*
Return auth secret name helper
*/}}
{{- define "common.auth._existingSecret" -}}
{{- $global := (get (.context.Values.global | default dict) .name | default dict) }}
{{- with default (.context.Values.auth).existingSecret ($global.auth).existingSecret -}}
    {{- printf "%s" (tpl . $.context) -}}
{{- end -}}
{{- end -}}

{{/*
Return auth secret name
*/}}
{{- define "common.auth.existingSecret" -}}
{{- include "common.auth._existingSecret" (dict "context" . "name" .Chart.Name) }}
{{- end -}}

{{/*
Return the secret containing auth info
*/}}
{{- define "common.auth.secretName" -}}
{{- with include "common.auth.existingSecret" . -}}
    {{- printf "%s" . -}}
{{- else -}}
    {{- printf "%s" (include "common.fullname" .) -}}
{{- end -}}
{{- end -}}
