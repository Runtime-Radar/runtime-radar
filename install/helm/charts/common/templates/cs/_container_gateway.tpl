{{- define "common.cs.container.gateway" -}}
- name: {{ include "common.name" . }}-gateway
  image: {{ include "common.cs.image" (dict "context" . "image" (.Values.gateway).image) }}
  imagePullPolicy: {{ eq (include "common.cs.devMode.enabled" .) "true" | ternary "Always" (default "IfNotPresent" ((.Values.gateway).image).pullPolicy) }}
  {{- with (.Values.gateway).command }}
  command:
    {{- tpl (toYaml .) $ | nindent 4 }}
  {{- end }}
  env:
    {{- with (.Values.gateway).env }}
    {{- tpl (toYaml .) . | nindent 4 }}
    {{- end }}
  envFrom:
    - configMapRef:
        name: {{ include "common.cs.configmapName" . }}
    {{- with (.Values.gateway).envFrom }}
    {{- tpl (toYaml .) $ | nindent 4 }}
    {{- end }}
  {{- with (.Values.gateway).containerPorts }}
  ports:
  {{- range $key, $val := . }}
  - name: {{ $key }}
    containerPort: {{ $val }}
  {{- end }}
  {{- end }}
  {{- with (.Values.gateway).resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with (.Values.gateway).startupProbe }}
  startupProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with (.Values.gateway).livenessProbe }}
  livenessProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with (.Values.gateway).readinessProbe }}
  readinessProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- if ne (include "common.cs.devMode.enabled" .) "true" }}
  {{- with (.Values.gateway).containerSecurityContext }}
  securityContext:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- end }}
  volumeMounts:
    {{- include "common.cs.volumeMounts" . | nindent 4 }}
{{- end -}}
