{{- define "common.cs.auth.enabled" -}}
{{- $globalAuth := hasKey ((.Values.global).auth) "enabled" | ternary ((.Values.global).auth).enabled true -}}
{{- if eq (default $globalAuth (.Values.auth).enabled | toString) "true" -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{/*
Return auth secret name
*/}}
{{- define "common.cs.auth.secretName" -}}
{{- with default (.Values.administrator).existingSecret ((.Values.global).administrator).existingSecret -}}
    {{- printf "%s" (tpl . $) -}}
{{- else -}}
    {{- printf "cs-account" -}}
{{- end -}}
{{- end -}}

{{/*
Return PostgreSQL auth existing secret name
*/}}
{{- define "common.cs.auth.postgresql.existingSecret" -}}
{{- include "common.auth._existingSecret" (dict "context" . "name" "postgresql") -}}
{{- end -}}

{{/*
Return PostgreSQL auth secret name
*/}}
{{- define "common.cs.auth.postgresql.secretName" -}}
{{- default "postgresql" (include "common.cs.auth.postgresql.existingSecret" .) -}}
{{- end -}}

{{/*
Return ClickHouse auth existing secret name
*/}}
{{- define "common.cs.auth.clickhouse.existingSecret" -}}
{{- include "common.auth._existingSecret" (dict "context" . "name" "clickhouse") -}}
{{- end -}}

{{/*
Return ClickHouse auth secret name
*/}}
{{- define "common.cs.auth.clickhouse.secretName" -}}
{{- default "clickhouse" (include "common.cs.auth.clickhouse.existingSecret" .) -}}
{{- end -}}

{{/*
Return Redis auth existing secret name
*/}}
{{- define "common.cs.auth.redis.existingSecret" -}}
{{- include "common.auth._existingSecret" (dict "context" . "name" "redis") -}}
{{- end -}}

{{/*
Return Redis auth secret name
*/}}
{{- define "common.cs.auth.redis.secretName" -}}
{{- default "redis" (include "common.cs.auth.redis.existingSecret" .) -}}
{{- end -}}

{{/*
Return RabbitMQ auth existing secret name
*/}}
{{- define "common.cs.auth.rabbitmq.existingSecret" -}}
{{- include "common.auth._existingSecret" (dict "context" . "name" "rabbitmq") -}}
{{- end -}}

{{/*
Return RabbitMQ auth secret name
*/}}
{{- define "common.cs.auth.rabbitmq.secretName" -}}
{{- default "rabbitmq" (include "common.cs.auth.rabbitmq.existingSecret" .) -}}
{{- end -}}
