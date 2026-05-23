{{/*
Return the proper ClickHouse image name
*/}}
{{- define "clickhouse.image" -}}
{{ include "common.image" (dict "context" . "image" .Values.image) }}
{{- end -}}

{{/*
Return the path to the cert file.
*/}}
{{- define "clickhouse.tlsCert" -}}
{{- printf "/etc/clickhouse-server/certs/%s" (default "tls.crt" .Values.tls.certFilename) -}}
{{- end -}}

{{/*
Return the path to the cert key file.
*/}}
{{- define "clickhouse.tlsCertKey" -}}
{{- printf "/etc/clickhouse-server/certs/%s" (default "tls.key" .Values.tls.certKeyFilename) -}}
{{- end -}}

{{/*
Return the path to the CA cert file.
*/}}
{{- define "clickhouse.tlsCACert" -}}
{{- printf "/etc/clickhouse-server/certs/%s" (default "ca.crt" .Values.tls.certCAFilename) -}}
{{- end -}}

{{/*
Get the ClickHouse password key inside the secret
*/}}
{{- define "clickhouse.secretPasswordKey" -}}
{{- if .Values.auth.existingSecret -}}
    {{- .Values.auth.existingSecretPasswordKey -}}
{{- else }}
    {{- print "admin-password" -}}
{{- end -}}
{{- end -}}

{{/*
Fix ClickHouse database name if it contains special symbols
*/}}
{{- define "clickhouse.database" -}}
{{- if regexMatch "^[a-zA-Z0-9_]*$" .Values.auth.database | not -}}
    {{- printf "`%s`" .Values.auth.database -}}
{{- else }}
    {{- .Values.auth.database -}}
{{- end -}}
{{- end -}}
