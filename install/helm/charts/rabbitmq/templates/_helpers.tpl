{{/*
Return the proper RabbitMQ image name
*/}}
{{- define "rabbitmq.image" -}}
{{ include "common.image" (dict "context" . "image" .Values.image) }}
{{- end -}}

{{/*
Get the password key to be retrieved from RabbitMQ secret.
*/}}
{{- define "rabbitmq.secretPasswordKey" -}}
{{- with .Values.auth.existingSecretPasswordKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "RABBIT_PASSWORD" -}}
{{- end -}}
{{- end -}}

{{/*
Get the username key.
Returns the configured existingSecretUserKey, otherwise the default
container-env-style key (also used for the chart-created secret).
*/}}
{{- define "rabbitmq.secretUserKey" -}}
{{- with .Values.auth.existingSecretUserKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "RABBIT_USER" -}}
{{- end -}}
{{- end -}}

{{/*
Return the proper RabbitMQ plugin list
*/}}
{{- define "rabbitmq.plugins" -}}
{{- $plugins := .Values.plugins -}}
{{- if (eq (include "common.metrics.enabled" .) "true") -}}
{{- $plugins = printf "%s %s" $plugins .Values.metrics.plugins -}}
{{- end -}}
{{- printf "%s" $plugins | replace " " ", " -}}
{{- end -}}
