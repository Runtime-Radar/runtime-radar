{{/*
Return the proper Redis image name
*/}}
{{- define "redis.image" -}}
{{ include "common.image" (dict "context" . "image" .Values.image) }}
{{- end -}}

{{/*
Get the password key to be retrieved from Redis&reg; secret.
*/}}
{{- define "redis.secretPasswordKey" -}}
{{- with .Values.auth.existingSecretPasswordKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "REDIS_PASSWORD" -}}
{{- end -}}
{{- end -}}

{{/*
Get the username key.
Returns the configured existingSecretUserKey, otherwise the default
container-env-style key (also used for the chart-created secret).
*/}}
{{- define "redis.secretUserKey" -}}
{{- with .Values.auth.existingSecretUserKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "REDIS_USER" -}}
{{- end -}}
{{- end -}}

{{/*
Return the path to the cert file.
*/}}
{{- define "redis.tlsCert" -}}
{{- printf "/etc/redis/certs/%s" (default "tls.crt" .Values.tls.certFilename) -}}
{{- end -}}

{{/*
Return the path to the cert key file.
*/}}
{{- define "redis.tlsCertKey" -}}
{{- printf "/etc/redis/certs/%s" (default "tls.key" .Values.tls.certKeyFilename) -}}
{{- end -}}

{{/*
Return the path to the CA cert file.
*/}}
{{- define "redis.tlsCACert" -}}
{{- printf "/etc/redis/certs/%s" (default "ca.crt" .Values.tls.certCAFilename) -}}
{{- end -}}
