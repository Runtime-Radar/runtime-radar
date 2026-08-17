{{- define "container.tetragon" -}}
- name: {{ include "container.tetragon.name" . }}
  securityContext:
    {{- toYaml .Values.tetragon.containerSecurityContext | nindent 4 }}
  image: {{ include "common.cs.image" (dict "context" . "image" .Values.tetragon.image) }}
  imagePullPolicy: {{ .Values.image.pullPolicy }}
  terminationMessagePolicy: FallbackToLogsOnError
{{- with .Values.tetragon.commandOverride }}
  command:
  {{- toYaml . | nindent 2 }}
{{- end }}
  args:
    - --config-dir=/etc/tetragon/tetragon.conf.d/
{{- with .Values.tetragon.argsOverride }}
  {{- toYaml . | nindent 2 }}
{{- else }}
{{- range $key, $value := .Values.tetragon.extraArgs }}
{{- if $value }}
    - --{{ $key }}={{ $value }}
{{- else }}
    - --{{ $key }}
{{- end }}
{{- end }}
{{- end }}
  volumeMounts:
    {{- with .Values.tetragon.extraVolumeMounts }}
      {{- toYaml . | nindent 4 }}
    {{- end }}
    - mountPath: /etc/tetragon/tetragon.conf.d/
      name: tetragon-config
      readOnly: true
    - mountPath: /sys/fs/bpf
      mountPropagation: Bidirectional
      name: bpf-maps
    - mountPath: "/var/run/cilium"
      name: cilium-run
    - mountPath: {{ .Values.exportDirectory }}
      name: export-logs
    - mountPath: "/procRoot"
      name: host-proc
    - mountPath: /tmp
      name: empty-dir
      subPath: tmp-dir
{{- if and (.Values.tetragon.cri.enabled) (.Values.tetragon.cri.socketHostPath) }}
    - mountPath: {{ quote .Values.tetragon.cri.socketHostPath }}
      name: cri-socket
{{- end }}
{{- range .Values.extraHostPathMounts }}
    - name: {{ .name }}
      mountPath: {{ .mountPath }}
      readOnly: {{ .readOnly }}
{{- if .mountPropagation }}
      mountPropagation: {{ .mountPropagation }}
{{- end }}
{{- end }}
{{- range .Values.extraConfigmapMounts }}
    - name: {{ .name }}
      mountPath: {{ .mountPath }}
      readOnly: {{ .readOnly }}
{{- end }}
  env:
    - name: NODE_NAME
      valueFrom:
        fieldRef:
            fieldPath: spec.nodeName
{{- if .Values.tetragon.extraEnv }}
  {{- toYaml .Values.tetragon.extraEnv | nindent 4 }}
{{- end }}
{{- with .Values.tetragon.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
{{- end }}
{{- if .Values.tetragon.livenessProbe }}
  livenessProbe:
  {{- toYaml .Values.tetragon.livenessProbe | nindent 4 }}
{{- else if .Values.tetragon.healthGrpc.enabled }}
  livenessProbe:
    timeoutSeconds: 60
    grpc:
      port: {{ .Values.tetragon.healthGrpc.port }}
      service: "liveness"
{{- else if .Values.tetragon.grpc.enabled }}
  livenessProbe:
    timeoutSeconds: 60
    exec:
      command:
      - tetra
      - status
      - --server-address
      - {{ .Values.tetragon.grpc.address }}
      - --retries
      - "5"
{{- end -}}
{{- end -}}
