{{/*
Compile all warnings into a single message.
*/}}
{{- define "cs.validateValues" -}}
{{- $messages := list -}}
{{- $messages = append $messages (include "cs.validateValues.hosts" .) -}}
{{- $messages = append $messages (include "cs.validateValues.keys" .) -}}
{{- $messages = append $messages (include "cs.validateValues.tls" .) -}}
{{- if eq .Values.validateSubChartImages true }}
{{- $messages = append $messages (include "cs.validateValues.images" .) -}}
{{- end }}
{{- $messages = without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{- printf "\nVALUES VALIDATION:\n%s" $message | fail -}}
{{- end -}}
{{- end -}}

{{/* Validate values of CS - keys */}}
{{- define "cs.validateValues.keys" -}}
{{- if and .Values.global.keys.encryption (ne .Values.global.keys.encryption "INIT-DO-NOT-USE") }}
{{- if regexMatch "^[a-fA-F0-9]+$" .Values.global.keys.encryption | not }}
cs: global.keys.encryption should be either hex-encoded string or `INIT-DO-NOT-USE`
    Correct string can be generated via `openssl rand -hex 32`
{{- end }}
{{- if ne (len .Values.global.keys.encryption) 64 }}
cs: global.keys.encryption must be 64 hex characters (32 bytes)
{{- end }}
{{- end }}
{{- if and .Values.global.keys.token (ne .Values.global.keys.token "INIT-DO-NOT-USE") }}
{{- if regexMatch "^[a-fA-F0-9]+$" .Values.global.keys.token | not }}
cs: global.keys.token should be either hex-encoded string or `INIT-DO-NOT-USE`
    Correct string can be generated via `openssl rand -hex 32`
{{- end }}
{{- if ne (len .Values.global.keys.token) 64 }}
cs: global.keys.token must be 64 hex characters (32 bytes)
{{- end }}
{{- end }}
{{- if and .Values.global.keys.publicAccessTokenSalt (ne .Values.global.keys.publicAccessTokenSalt "INIT-DO-NOT-USE") }}
{{- if regexMatch "^[a-fA-F0-9]+$" .Values.global.keys.publicAccessTokenSalt | not }}
cs: global.keys.publicAccessTokenSalt should be either hex-encoded string or `INIT-DO-NOT-USE`
    Correct string can be generated via `openssl rand -hex 64`
{{- end }}
{{- if ne (len .Values.global.keys.publicAccessTokenSalt) 128 }}
cs: global.keys.publicAccessTokenSalt must be 128 hex characters (64 bytes)
{{- end }}
{{- end }}
{{- end -}}

{{/* Validate values of CS - hosts */}}
{{- define "cs.validateValues.hosts" -}}
{{- if and .Values.postgresql.externalHost (ne .Values.postgresql.deploy false) }}
cs: postgresql.externalHost is not supported with postgresql.deploy
    Set postgresql.deploy to false to proceed with external host
{{- end -}}
{{- if and .Values.redis.externalHost (ne .Values.redis.deploy false) }}
cs: redis.externalHost is not supported with redis.deploy
    Set redis.deploy to false to proceed with external host
{{- end -}}
{{- if and .Values.rabbitmq.externalHost (ne .Values.rabbitmq.deploy false) }}
cs: rabbitmq.externalHost is not supported with rabbitmq.deploy
    Set rabbitmq.deploy to false to proceed with external host
{{- end -}}
{{- if and .Values.clickhouse.externalHost (ne .Values.clickhouse.deploy false) }}
cs: clickhouse.externalHost is not supported with clickhouse.deploy
    Set clickhouse.deploy to false to proceed with external host
{{- end -}}
{{- end -}}

{{/* Validate values of CS - TLS */}}
{{- define "cs.validateValues.tls" -}}
{{- if and .Values.global.postgresql.tls.enabled .Values.global.postgresql.tls.verify .Values.postgresql.externalHost (not .Values.postgresql.tls.certCA) }}
cs: tls
    A valid .Values.postgresql.tls.certCA entry required to verify TLS with external host!
{{- end}}
{{- if and .Values.global.redis.tls.enabled .Values.global.redis.tls.verify .Values.redis.externalHost (not .Values.redis.tls.certCA) }}
cs: tls
    A valid .Values.redis.tls.certCA entry required to verify TLS with external host!
{{- end}}
{{- if and .Values.global.clickhouse.tls.enabled .Values.global.clickhouse.tls.verify .Values.clickhouse.externalHost (not .Values.clickhouse.tls.certCA) }}
cs: tls
    A valid .Values.clickhouse.tls.certCA entry required to verify TLS with external host!
{{- end}}
{{- end -}}

{{/* Validate values of CS - images */}}
{{- define "cs.validateValues.images" -}}
{{- $annotations := dict }}
{{- range $key, $value := .Chart.Annotations }}
{{- $_ := set $annotations $key $value }}
{{- end }}

{{- if ne ($annotation := get $annotations "runtime-monitor.tetragon") ($image := include "common.image" (dict "context" . "image" (get .Values "runtime-monitor").tetragon.image)) }}
cs: runtime-monitor.tetragon.image is incorrect
    Image in annotation is not the same as in values of chart or subchart {{ printf "('%s' != '%s')" $annotation $image }}
{{- end }}
{{- if ne ($annotation := get $annotations "postgresql") ($image := include "common.image" (dict "context" . "image" .Values.postgresql.image)) }}
cs: postgresql.image is incorrect
    Image in annotation is not the same as in values of chart or subchart {{ printf "('%s' != '%s')" $annotation $image }}
{{- end }}
{{- if ne ($annotation := get $annotations "postgresql.metrics") ($image := include "common.image" (dict "context" . "image" .Values.postgresql.metrics.image)) }}
cs: postgresql.metrics.image is incorrect
    Image in annotation is not the same as in values of chart or subchart {{ printf "('%s' != '%s')" $annotation $image }}
{{- end }}
{{- if ne ($annotation := get $annotations "redis") ($image := include "common.image" (dict "context" . "image" .Values.redis.image)) }}
cs: redis.image is incorrect
    Image in annotation is not the same as in values of chart or subchart {{ printf "('%s' != '%s')" $annotation $image }}
{{- end }}
{{- if ne ($annotation := get $annotations "rabbitmq") ($image := include "common.image" (dict "context" . "image" .Values.rabbitmq.image)) }}
cs: rabbitmq.image is incorrect
    Image in annotation is not the same as in values of chart or subchart {{ printf "('%s' != '%s')" $annotation $image }}
{{- end }}
{{- if ne ($annotation := get $annotations "clickhouse") ($image := include "common.image" (dict "context" . "image" .Values.clickhouse.image)) }}
cs: clickhouse.image is incorrect
    Image in annotation is not the same as in values of chart or subchart {{ printf "('%s' != '%s')" $annotation $image }}
{{- end }}
{{- end -}}
