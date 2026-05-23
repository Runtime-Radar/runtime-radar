{{- define "common.cs.container.logger" -}}
- name: logger
  image: {{ include "common.image" (dict "context" . "image" ((.Values.global).logger).image) }}
  imagePullPolicy: {{ (((.Values.global).logger).image).pullPolicy | quote }}
  command:
    - /fluent-bit/bin/fluent-bit
    - -c
    - /fluent-bit/etc/fluent-bit.yaml
  env:
    - name: POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: COMPONENT
      value: {{ include "common.name" . }}
    - name: BUFFER_CHUNK_SIZE
      value: {{ default "128k" (default ((.Values.global).logger).bufferChunkSize (.Values.logger).bufferChunkSize) }}
    - name: BUFFER_MAX_SIZE
      value: {{ default "128k" (default ((.Values.global).logger).bufferMaxSize (.Values.logger).bufferMaxSize) }}
    {{- with (.Values.logger).env }}
    {{- tpl (toYaml .) . | nindent 4 }}
    {{- end }}
  envFrom:
    - configMapRef:
        name: {{ include "common.cs.configmapName" . }}
  {{- with (.Values.logger).resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- else }}
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
      ephemeral-storage: 50Mi
    limits:
      cpu: 150m
      memory: 192Mi
      ephemeral-storage: 1024Mi
  {{- end }}
  {{- with (.Values.logger).startupProbe }}
  startupProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with (.Values.logger).livenessProbe }}
  livenessProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with (.Values.logger).readinessProbe }}
  readinessProbe: {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- if ne (include "common.cs.devMode.enabled" .) "true" }}
  {{- with (.Values.logger).containerSecurityContext }}
  securityContext:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- end }}
  volumeMounts:
    - name: empty-dir
      mountPath: /tmp
      subPath: tmp-dir
    - name: config-logger
      mountPath: /fluent-bit/etc/fluent-bit.yaml
      readOnly: true
      subPath: fluent-bit.yaml
    {{- if eq (include "common.cs.tls.loki.enabled" .) "true"  }}
    - name: certificates-loki
      mountPath: /etc/ssl/certs/loki.pem
      readOnly: true
      subPath: ca.crt
    {{- end }}
    - name: empty-dir
      mountPath: /var/log/cs
      subPath: log-dir
{{- end -}}
