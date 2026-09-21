{{/*
Return true if TLS is enabled for CS
*/}}
{{- define "common.cs.tls.enabled" -}}
{{- $globalTLS := hasKey ((.Values.global).tls) "enabled" | ternary ((.Values.global).tls).enabled true -}}
{{- if eq (default $globalTLS (.Values.tls).enabled | toString) "true" -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.tls.component.enabled" -}}
{{- $values := get .context.Values .component | default dict -}}
{{- $global := get .context.Values.global .component | default dict -}}
{{- default ($global.tls).enabled ($values.tls).enabled | empty | not -}}
{{- end -}}

{{- define "common.cs.tls.component.verify" -}}
{{- $values := get .context.Values .component | default dict -}}
{{- $global := get (.context.Values.global | default dict) .component | default dict -}}
{{- $tls := default ($global.tls).enabled ($values.tls).enabled | empty | not -}}
{{- $verify := default ($global.tls).verify ($values.tls).verify | empty | not -}}
{{- and $tls $verify -}}
{{- end -}}

{{- define "common.cs.tls.postgresql.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "postgresql") | eq "true" | and (.Values.postgresql).enabled -}}
{{- end -}}

{{- define "common.cs.tls.postgresql.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "postgresql") | eq "true" | and (.Values.postgresql).enabled -}}
{{- end -}}

{{- define "common.cs.tls.clickhouse.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "clickhouse") | eq "true" | and (.Values.clickhouse).enabled -}}
{{- end -}}

{{- define "common.cs.tls.clickhouse.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "clickhouse") | eq "true" | and (.Values.clickhouse).enabled -}}
{{- end -}}

{{- define "common.cs.tls.redis.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "redis") | eq "true" | and (.Values.redis).enabled -}}
{{- end -}}

{{- define "common.cs.tls.redis.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "redis") | eq "true" | and (.Values.redis).enabled -}}
{{- end -}}

{{- define "common.cs.tls.grafana.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "grafana") -}}
{{- end -}}

{{- define "common.cs.tls.grafana.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "grafana") -}}
{{- end -}}

{{- define "common.cs.tls.prometheus.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "prometheus") -}}
{{- end -}}

{{- define "common.cs.tls.prometheus.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "prometheus") -}}
{{- end -}}

{{- define "common.cs.tls.loki.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "loki") -}}
{{- end -}}

{{- define "common.cs.tls.loki.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "loki") -}}
{{- end -}}

{{/*
Return certificates secret name
*/}}
{{- define "common.cs.tls.secretName" -}}
{{- default (printf "%s-crt" (include "common.cs.basename" .)) ((.Values.global).tls).existingSecret -}}
{{- end -}}

{{/*
Return PostgreSQL certificates secret name
*/}}
{{- define "common.cs.tls.postgresql.secretName" -}}
{{- default "postgresql-crt" (((.Values.global).postgresql).tls).existingSecret -}}
{{- end -}}

{{/*
Return ClickHouse certificates secret name
*/}}
{{- define "common.cs.tls.clickhouse.secretName" -}}
{{- default "clickhouse-crt" (((.Values.global).clickhouse).tls).existingSecret -}}
{{- end -}}

{{/*
Return Redis certificates secret name
*/}}
{{- define "common.cs.tls.redis.secretName" -}}
{{- default "redis-crt" (((.Values.global).redis).tls).existingSecret -}}
{{- end -}}

{{/*
Return Grafana certificates secret name
*/}}
{{- define "common.cs.tls.grafana.secretName" -}}
{{- default "grafana-crt" (((.Values.global).grafana).tls).existingSecret -}}
{{- end -}}

{{/*
Return Loki certificates secret name
*/}}
{{- define "common.cs.tls.loki.secretName" -}}
{{- default "loki-crt" (((.Values.global).loki).tls).existingSecret -}}
{{- end -}}
