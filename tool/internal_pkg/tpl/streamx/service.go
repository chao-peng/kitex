package streamx

var ServiceTpl = `// Code generated by Kitex {{.Version}}. DO NOT EDIT.
package {{ToLower .ServiceName}}

import (
	{{- range $path, $aliases := .Imports}}
		{{- if not $aliases}}
			"{{$path}}"
		{{- else}}
			{{- range $alias, $is := $aliases}}
				{{$alias}} "{{$path}}"
			{{- end}}
		{{- end}}
	{{- end}}
)
{{- $protocol := .Protocol | getStreamxRef}}

var svcInfo = &serviceinfo.ServiceInfo{
    ServiceName: "{{.RawServiceName}}",
    Methods: map[string]serviceinfo.MethodInfo{
        {{- range .AllMethods}}
        {{- $unary := and (not .ServerStreaming) (not .ClientStreaming)}}
        {{- $clientSide := and .ClientStreaming (not .ServerStreaming)}}
        {{- $serverSide := and (not .ClientStreaming) .ServerStreaming}}
        {{- $bidiSide := and .ClientStreaming .ServerStreaming}}
        {{- $arg := index .Args 0}}
        {{- $mode := ""}}
            {{- if $bidiSide -}} {{- $mode = "serviceinfo.StreamingBidirectional" }}
            {{- else if $serverSide -}} {{- $mode = "serviceinfo.StreamingServer" }}
			{{- else if $clientSide -}} {{- $mode = "serviceinfo.StreamingClient" }}
			{{- else if $unary -}} {{- $mode = "serviceinfo.StreamingUnary" }}
            {{- end}}
		"{{.RawName}}": serviceinfo.NewMethodInfo(
			func(ctx context.Context, handler, reqArgs, resArgs interface{}) error {
				return streamxserver.InvokeStream[{{$protocol}}.Header, {{$protocol}}.Trailer, {{NotPtr $arg.Type}}, {{NotPtr .Resp.Type}}](
					ctx, {{$mode}}, handler.(streamx.StreamHandler), reqArgs.(streamx.StreamReqArgs), resArgs.(streamx.StreamResArgs))
			},
			nil,
			nil,
			false,
			serviceinfo.WithStreamingMode({{$mode}}),
		),
        {{- end}}
    },
    Extra: map[string]interface{}{
        "streaming": true,
        "streamx": true,
    },
}

`
