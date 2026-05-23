{{- define "common.cs.container.app" -}}
- name: {{ include "common.name" . }}
  image: {{ include "common.cs.image" (dict "context" . "image" .Values.image) }}
  imagePullPolicy: {{ eq (include "common.cs.devMode.enabled" .) "true" | ternary "Always" (default "IfNotPresent" (.Values.image).pullPolicy) }}
  {{- with .Values.command }}
  command:
    {{- tpl (toYaml .) $ | nindent 4 }}
  {{- end }}
  env:
    {{- include "common.cs.container.app.env" . | nindent 4 }}
  envFrom:
    {{- include "common.cs.container.app.envFrom" . | nindent 4 }}
  {{- with .Values.containerPorts }}
  ports:
    {{- range $key, $val := . }}
    - name: {{ $key }}
      containerPort: {{ tpl (toString $val) $ }}
    {{- end }}
  {{- end }}
  {{- with .Values.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.startupProbe }}
  startupProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.livenessProbe }}
  livenessProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.readinessProbe }}
  readinessProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- if ne (include "common.cs.devMode.enabled" .) "true" }}
  {{- with .Values.containerSecurityContext }}
  securityContext:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- end }}
  volumeMounts:
    {{- include "common.cs.volumeMounts" . | nindent 4 }}
{{- end -}}

{{- define "common.cs.container.app.env" -}}
{{- if eq (include "common.cs.auth.enabled" .) "true" }}
- name: TOKEN_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "common.cs.keys.secretName" . }}
      key: token
{{- end }}
{{- if eq (include "common.cs.encryption.enabled" .) "true" }}
- name: ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "common.cs.keys.secretName" . }}
      key: encryption
{{- end }}
{{- if (.Values.rabbitmq).enabled }}
{{- with (.Values.rabbitmq).queue }}
- name: RABBIT_QUEUE
  value: {{ . | quote }}
{{- end }}
{{- with (.Values.rabbitmq).runtimeEventsQueue }}
- name: RABBIT_RUNTIME_EVENTS_QUEUE
  value: {{. | quote }}
{{- end }}
{{- with (.Values.rabbitmq).historyEventsQueue }}
- name: RABBIT_HISTORY_EVENTS_QUEUE
  value: {{. | quote }}
{{- end }}
{{- end }}
{{- if .Values.administrator }}
- name: ADMINISTRATOR_USERNAME
  valueFrom:
    secretKeyRef:
      name: {{ include "common.cs.auth.secretName" . }}
      key: username
- name: ADMINISTRATOR_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "common.cs.auth.secretName" . }}
      key: password
{{- end }}
{{- with .Values.env }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end -}}

{{- define "common.cs.container.app.envFrom" -}}
- configMapRef:
    name: {{ include "common.cs.configmapName" . }}
{{- if (.Values.postgresql).enabled }}
- secretRef:
    name: {{ include "common.cs.postgresql.secretName" . }}
- configMapRef:
    name: {{ include "common.cs.postgresql.configmapName" . }}
{{- end }}
{{- if (.Values.redis).enabled }}
- secretRef:
    name: {{ include "common.cs.redis.secretName" . }}
- configMapRef:
    name: {{ include "common.cs.redis.configmapName" . }}
{{- end }}
{{- if (.Values.rabbitmq).enabled }}
- secretRef:
    name: {{ include "common.cs.rabbitmq.secretName" . }}
- configMapRef:
    name: {{ include "common.cs.rabbitmq.configmapName" . }}
{{- end }}
{{- if (.Values.clickhouse).enabled }}
- secretRef:
    name: {{ include "common.cs.clickhouse.secretName" . }}
- configMapRef:
    name: {{ include "common.cs.clickhouse.configmapName" . }}
{{- end }}
{{- with .Values.envFrom }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end -}}
