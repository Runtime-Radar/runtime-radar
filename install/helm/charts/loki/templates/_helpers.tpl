{{/*
Return the proper loki image name
*/}}
{{- define "loki.image" -}}
{{ include "common.image" (dict "context" . "image" .Values.image) }}
{{- end -}}

{{/*
Return the path to the cert file.
*/}}
{{- define "loki.tlsCert" -}}
{{- printf "/etc/loki/certs/%s" (default "tls.crt" .Values.tls.certFilename) -}}
{{- end -}}

{{/*
Return the path to the cert key file.
*/}}
{{- define "loki.tlsCertKey" -}}
{{- printf "/etc/loki/certs/%s" (default "tls.key" .Values.tls.certKeyFilename) -}}
{{- end -}}

{{/*
Return the path to the CA cert file.
*/}}
{{- define "loki.tlsCACert" -}}
{{- printf "/etc/loki/certs/%s" (default "ca.crt" .Values.tls.certCAFilename) -}}
{{- end -}}
