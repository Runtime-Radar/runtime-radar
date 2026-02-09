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
{{- $enabled := $values.enabled | empty | not -}}
{{- $tls := default ($global.tls).enabled ($values.tls).enabled | empty | not -}}
{{- if and $enabled $tls -}}
  {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.tls.component.verify" -}}
{{- $values := get .context.Values .component | default dict -}}
{{- $global := get (.context.Values.global | default dict) .component | default dict -}}
{{- $enabled := $values.enabled | empty | not -}}
{{- $tls := default ($global.tls).verify ($values.tls).verify | empty | not -}}
{{- if and $enabled $tls -}}
  {{- true -}}
{{- end -}}
{{- end -}}

{{- define "common.cs.tls.postgresql.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "postgresql") -}}
{{- end -}}

{{- define "common.cs.tls.postgresql.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "postgresql") -}}
{{- end -}}

{{- define "common.cs.tls.clickhouse.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "clickhouse") -}}
{{- end -}}

{{- define "common.cs.tls.clickhouse.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "clickhouse") -}}
{{- end -}}

{{- define "common.cs.tls.redis.enabled" -}}
{{- include "common.cs.tls.component.enabled" (dict "context" . "component" "redis") -}}
{{- end -}}

{{- define "common.cs.tls.redis.verify" -}}
{{- include "common.cs.tls.component.verify" (dict "context" . "component" "redis") -}}
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
