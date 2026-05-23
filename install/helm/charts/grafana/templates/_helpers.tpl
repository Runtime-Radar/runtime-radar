{{/*
Return the proper Grafana image name
*/}}
{{- define "grafana.image" -}}
{{ include "common.image" (dict "context" . "image" .Values.image) }}
{{- end -}}

{{/*
Return the path to the cert file.
*/}}
{{- define "grafana.tlsCert" -}}
{{- printf "/etc/grafana/certs/%s" (default "tls.crt" .Values.tls.certFilename) -}}
{{- end -}}

{{/*
Return the path to the cert key file.
*/}}
{{- define "grafana.tlsCertKey" -}}
{{- printf "/etc/grafana/certs/%s" (default "tls.key" .Values.tls.certKeyFilename) -}}
{{- end -}}

{{/*
Return the path to the CA cert file.
*/}}
{{- define "grafana.tlsCACert" -}}
{{- printf "/etc/grafana/certs/%s" (default "ca.crt" .Values.tls.certCAFilename) -}}
{{- end -}}

{{/*
Get the Grafana password key inside the secret
*/}}
{{- define "grafana.secretPasswordKey" -}}
{{- with .Values.auth.existingSecretPasswordKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "GF_SECURITY_ADMIN_PASSWORD" -}}
{{- end -}}
{{- end -}}

{{/*
Get the username key.
Returns the configured existingSecretUserKey, otherwise the default
container-env-style key (also used for the chart-created secret).
*/}}
{{- define "grafana.secretUserKey" -}}
{{- with .Values.auth.existingSecretUserKey -}}
    {{- tpl . $ -}}
{{- else -}}
    {{- "GF_SECURITY_ADMIN_USER" -}}
{{- end -}}
{{- end -}}

{{- define "grafana.rootUrl" -}}
{{- $root := "%(protocol)s://%(domain)s:%(http_port)s" }}
{{- with .Values.subPath }}
  {{- printf "%s/%s/" $root (trimAll "/" .) }}
{{- else }}
  {{- printf "%s/" $root }}
{{- end -}}
{{- end -}}
