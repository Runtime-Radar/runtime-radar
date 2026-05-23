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
{{- with .Values.auth.existingSecretPasswordKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "CLICKHOUSE_PASSWORD" -}}
{{- end -}}
{{- end -}}

{{/*
Get the username key.
Returns the configured existingSecretUserKey, otherwise the default
container-env-style key (also used for the chart-created secret).
*/}}
{{- define "clickhouse.secretUserKey" -}}
{{- with .Values.auth.existingSecretUserKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "CLICKHOUSE_USER" -}}
{{- end -}}
{{- end -}}

{{/*
Get the database key.
Returns the configured existingSecretDatabaseKey, otherwise the default
container-env-style key (also used for the chart-created secret).
*/}}
{{- define "clickhouse.secretDatabaseKey" -}}
{{- with .Values.auth.existingSecretDatabaseKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "CLICKHOUSE_DB" -}}
{{- end -}}
{{- end -}}
