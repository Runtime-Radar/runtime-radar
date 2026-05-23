{{- define "common.cs.basename" -}}
{{- printf "cs" -}}
{{- end -}}

{{- define "common.cs.configmapName" -}}
{{- printf "%s-config" (include "common.cs.basename" .) -}}
{{- end -}}

{{- define "common.cs.logger.configmapName" -}}
{{- printf "logger-config" -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "common.cs.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "common.name" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Return own cs url
*/}}
{{- define "common.cs.ownCsUrl" -}}
{{- $url := default (.Values.global).ownCsUrl .Values.ownCsUrl -}}
{{- required "global.ownCsUrl argument is required for installation" $url -}}
{{- end -}}

{{/*
Return central cs url
*/}}
{{- define "common.cs.centralCsUrl" -}}
{{- $url := default (.Values.global).centralCsUrl .Values.centralCsUrl -}}
{{- if eq (include "common.cs.isChildCluster" .) "true" -}}
{{- required "global.centralCsUrl argument is required for installation" $url -}}
{{- else -}}
{{- default (include "common.cs.ownCsUrl" .) $url -}}
{{- end -}}
{{- end -}}

{{/*
Return cs version
*/}}
{{- define "common.cs.csVersion" -}}
{{- default (.Values.global).csVersion .Values.csVersion | default .Chart.Version -}}
{{- end -}}

{{/*
Return keys secret name
*/}}
{{- define "common.cs.keys.secretName" -}}
{{- default "cs-keys" ((.Values.global).keys).existingSecret -}}
{{- end -}}

{{- define "common.cs.grafana.address" -}}
{{- default (.Values.grafana).externalHost ((.Values.global).grafana).externalHost | default (printf "%s://grafana.%s.svc.cluster.local:3000" (include "common.cs.grafana.http-scheme" .) .Release.Namespace) | trimSuffix "/" }}
{{- end -}}

{{- define "common.cs.prometheus.address" -}}
{{- default (.Values.prometheus).externalHost ((.Values.global).prometheus).externalHost | default (printf "%s://prometheus.%s.svc.cluster.local:9090" (include "common.cs.prometheus.http-scheme" .) .Release.Namespace) | trimSuffix "/" }}
{{- end -}}

{{- define "common.cs.loki.address" -}}
{{- default (.Values.loki).externalHost ((.Values.global).loki).externalHost | default (printf "%s://loki.%s.svc.cluster.local:3100" (include "common.cs.loki.http-scheme" .) .Release.Namespace) | trimSuffix "/" }}
{{- end -}}

{{/*
Return postgresql secret name.
Precedence: global.postgresql.auth.existingSecret > local postgresql.auth.existingSecret > "postgresql".
*/}}
{{- define "common.cs.postgresql.secretName" -}}
{{- $local := ((.Values.postgresql).auth).existingSecret -}}
{{- $global := (((.Values.global).postgresql).auth).existingSecret -}}
{{- default "postgresql" (default $local $global) -}}
{{- end -}}

{{/*
Return redis secret name.
Precedence: global.redis.auth.existingSecret > local redis.auth.existingSecret > "redis".
*/}}
{{- define "common.cs.redis.secretName" -}}
{{- $local := ((.Values.redis).auth).existingSecret -}}
{{- $global := (((.Values.global).redis).auth).existingSecret -}}
{{- default "redis" (default $local $global) -}}
{{- end -}}

{{/*
Return rabbitmq secret name.
Precedence: global.rabbitmq.auth.existingSecret > local rabbitmq.auth.existingSecret > "rabbitmq".
*/}}
{{- define "common.cs.rabbitmq.secretName" -}}
{{- $local := ((.Values.rabbitmq).auth).existingSecret -}}
{{- $global := (((.Values.global).rabbitmq).auth).existingSecret -}}
{{- default "rabbitmq" (default $local $global) -}}
{{- end -}}

{{/*
Return clickhouse secret name.
Precedence: global.clickhouse.auth.existingSecret > local clickhouse.auth.existingSecret > "clickhouse".
*/}}
{{- define "common.cs.clickhouse.secretName" -}}
{{- $local := ((.Values.clickhouse).auth).existingSecret -}}
{{- $global := (((.Values.global).clickhouse).auth).existingSecret -}}
{{- default "clickhouse" (default $local $global) -}}
{{- end -}}

{{/*
Return postgresql config map name.
*/}}
{{- define "common.cs.postgresql.configmapName" -}}
{{- printf "postgresql" -}}
{{- end -}}

{{/*
Return redis config map name.
*/}}
{{- define "common.cs.redis.configmapName" -}}
{{- printf "redis" -}}
{{- end -}}

{{/*
Return rabbitmq config map name.
*/}}
{{- define "common.cs.rabbitmq.configmapName" -}}
{{- printf "rabbitmq" -}}
{{- end -}}

{{/*
Return clickhouse config map name.
*/}}
{{- define "common.cs.clickhouse.configmapName" -}}
{{- printf "clickhouse" -}}
{{- end -}}
