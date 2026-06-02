{{/*
Return the proper image name
{{ include "common.cs.image" ( dict "context" . "image" .Values.path.to.the.image) }}
*/}}
{{- define "common.cs.image" -}}
{{- with (.context.Values.global).imageRegistry -}}
{{- $_ := set $.image "registry" . -}}
{{- end -}}
{{- $tag := default (include "common.cs.csVersion" .context) (.context.Values.global).imageTag -}}
{{- include "common.image" (dict "defaultTag" $tag | mergeOverwrite .) -}}
{{- end -}}
