{{ range .Errors }}

{{ if .HasComment }}{{ .Comment }}{{ end -}}
func Is{{.CamelValue}}(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == {{ .Name }}_{{ .Value }}.String() && e.Code == {{ .HTTPCode }}
}

{{ if .HasComment }}{{ .Comment }}{{ end -}}
func Error{{ .CamelValue }}(format string, args ...interface{}) *errors.Error {
	if format == "" {
		return errors.New({{ .HTTPCode }}, {{ .Name }}_{{ .Value }}.String(), {{ .QuotedMessage }}).
			WithMetadata(map[string]string{"code":"{{ .NumCode }}"})
	}
	return errors.New({{ .HTTPCode }}, {{ .Name }}_{{ .Value }}.String(), fmt.Sprintf(format, args...)).
		WithMetadata(map[string]string{"code":"{{ .NumCode }}"})
}

{{- end }}
