{{- define "common.cs.devMode.enabled" -}}
{{- if any (.Values.global).devMode .Values.devMode (.Values.global).dev .Values.dev }}
    {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.logger.enabled" -}}
{{- $globalLogger := hasKey ((.Values.global).logger) "enabled" | ternary ((.Values.global).logger).enabled true -}}
{{- if eq (default $globalLogger (.Values.logger).enabled | toString) "true" -}}
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

{{- define "common.cs.grafana.http-scheme" -}}
{{- eq (include "common.cs.tls.grafana.enabled" .) "true" | ternary "https" "http" -}}
{{- end -}}

{{- define "common.cs.prometheus.http-scheme" -}}
{{- eq (include "common.cs.tls.prometheus.enabled" .) "true" | ternary "https" "http" -}}
{{- end -}}

{{- define "common.cs.loki.http-scheme" -}}
{{- eq (include "common.cs.tls.loki.enabled" .) "true" | ternary "https" "http" -}}
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
- name: empty-dir
  mountPath: /tmp
  subPath: tmp-dir
- name: empty-dir
  mountPath: /.config
  subPath: gops-dir
{{- if eq (include "common.cs.tls.enabled" .) "true"  }}
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
{{- if eq (include "common.cs.tls.postgresql.verify" .) "true"  }}
- name: certificates-postgresql
  mountPath: db_ca.pem
  subPath: ca.crt
  readOnly: true
{{- end }}
{{- if eq (include "common.cs.tls.redis.verify" .) "true"  }}
- name: certificates-redis
  mountPath: redis_ca.pem
  subPath: ca.crt
  readOnly: true
{{- end }}
{{- if eq (include "common.cs.tls.clickhouse.verify" .) "true"  }}
- name: certificates-clickhouse
  mountPath: /etc/ssl/certs/clickhouse.pem
  readOnly: true
  subPath: ca.crt
{{- end }}
{{- if eq (include "common.cs.logger.enabled" .) "true" }}
- name: empty-dir
  mountPath: /var/log/cs
  subPath: log-dir
{{- end }}
{{- with .Values.volumeMounts }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "common.cs.volumes" -}}
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
{{- if eq (include "common.cs.tls.grafana.verify" .) "true" }}
- name: certificates-grafana
  secret:
    secretName: {{ include "common.cs.tls.grafana.secretName" . }}
{{- end }}
{{- if eq (include "common.cs.tls.loki.verify" .) "true" }}
- name: certificates-loki
  secret:
    secretName: {{ include "common.cs.tls.loki.secretName" . }}
{{- end }}
{{- if eq (include "common.cs.logger.enabled" .) "true" }}
- name: config-logger
  configMap:
    name: {{ include "common.cs.logger.configmapName" . }}
{{- end }}
{{- with .Values.volumes }}
{{ toYaml . }}
{{- end }}
{{- end -}}
