{{- define "common.cs.devMode.enabled" -}}
{{- if any (.Values.global).devMode .Values.devMode }}
    {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.auth.enabled" -}}
{{- $globalAuth := hasKey ((.Values.global).auth) "enabled" | ternary ((.Values.global).auth).enabled true -}}
{{- if eq (default $globalAuth (.Values.auth).enabled | toString) "true" -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.encryption.enabled" -}}
{{- if eq (default ((.Values.global).encryption).enabled (.Values.encryption).enabled | toString) "true" -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.isChildCluster" -}}
{{- if eq (default .Values.isChildCluster (.Values.global).isChildCluster | toString) "true" -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.replicas" -}}
{{- eq (include "common.cs.devMode.enabled" .) "true" | ternary 1 (default (.Values.global).replicas .Values.replicas) -}}
{{- end -}}

{{- define "common.cs.http-scheme" -}}
{{- eq (include "common.cs.tls.enabled" .) "true" | ternary "https" "http" -}}
{{- end -}}

{{- define "common.cs.logLevel" -}}
{{- if .Values.logLevel -}}
    {{- .Values.logLevel -}}
{{- else if eq (include "common.cs.devMode.enabled" .) "true" -}}
    {{- "DEBUG" -}}
{{- else if (.Values.global).logLevel -}}
    {{- (.Values.global).logLevel -}}
{{- else -}}
    {{- "INFO" -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.volumeMounts" -}}
{{- $val := hasKey . "values" | ternary (default (dict) .values) .context.Values -}}
volumeMounts:
  - name: empty-dir
    mountPath: /tmp
    subPath: tmp-dir
  - name: empty-dir
    mountPath: /.config
    subPath: gops-dir
  {{- if eq (include "common.cs.tls.enabled" .context) "true"  }}
  - name: certificates
    mountPath: ca.pem
    subPath: ca.crt
    readOnly: true
  - name: certificates
    mountPath: /etc/ssl/certs/ca.pem
    readOnly: true
    subPath: ca.crt
  - name: certificates
    mountPath: cert.pem
    subPath: tls.crt
    readOnly: true
  - name: certificates
    mountPath: key.pem
    subPath: tls.key
    readOnly: true
  {{- end }}
  {{- if eq (include "common.cs.tls.postgresql.verify" .context) "true"  }}
  - name: certificates-postgresql
    mountPath: db_ca.pem
    subPath: ca.crt
    readOnly: true
  {{- end }}
  {{- if eq (include "common.cs.tls.redis.verify" .context) "true"  }}
  - name: certificates-redis
    mountPath: redis_ca.pem
    subPath: ca.crt
    readOnly: true
  {{- end }}
  {{- if eq (include "common.cs.tls.clickhouse.verify" .context) "true"  }}
  - name: certificates-clickhouse
    mountPath: /etc/ssl/certs/clickhouse.pem
    readOnly: true
    subPath: ca.crt
  {{- end }}
  {{- with $val.volumeMounts }}
  {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end -}}

{{- define "common.cs.volumes" -}}
volumes:
  - name: empty-dir
    emptyDir: {}
  {{- if eq (include "common.cs.tls.enabled" .) "true" }}
  - name: certificates
    secret:
      secretName: {{ include "common.cs.tls.secretName" . }}
  {{- end }}
  {{- if eq (include "common.cs.tls.postgresql.verify" .) "true" }}
  - name: certificates-postgresql
    secret:
      secretName: {{ include "common.cs.tls.postgresql.secretName" . }}
  {{- end }}
  {{- if eq (include "common.cs.tls.redis.verify" .) "true" }}
  - name: certificates-redis
    secret:
      secretName: {{ include "common.cs.tls.redis.secretName" . }}
  {{- end }}
  {{- if eq (include "common.cs.tls.clickhouse.verify" .) "true" }}
  - name: certificates-clickhouse
    secret:
      secretName: {{ include "common.cs.tls.clickhouse.secretName" . }}
  {{- end }}
  {{- with .Values.volumes }}
  {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end -}}
