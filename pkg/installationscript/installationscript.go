package installationscript

import (
	"bytes"
	texttemplate "text/template"

	"github.com/authgear/authgear-once-license-server/pkg/uname"
)

var tmpl *texttemplate.Template

func init() {
	t, err := texttemplate.New("").Parse(`#!/bin/sh
set -e

echo "Installing the Authgear ONCE command......"
echo "This script uses sudo, you will be prompted for authentication."
sudo true

download_url="{{ $.DownloadURL }}?uname_s=$(uname -s)&uname_m=$(uname -m)"
tmp_path="$(mktemp)"
curl -fsSL "$download_url" > "$tmp_path"
sudo mv "$tmp_path" /usr/local/bin/authgear-once
sudo chmod u+x /usr/local/bin/authgear-once

{{- $image := "" }}
{{- if $.ImageOverride }}
	{{- $image = printf "--image '%v'" $.ImageOverride }}
{{- end }}

if [ "$(uname -s)" = "Darwin" ]; then
	/usr/local/bin/authgear-once setup {{ $image }} "{{ $.LicenseKey }}"
else
	sudo /usr/local/bin/authgear-once setup {{ $image }} "{{ $.LicenseKey }}"
fi
`)
	if err != nil {
		panic(err)
	}
	tmpl = t
}

type RenderOptions struct {
	DownloadURL   string
	LicenseKey    string
	ImageOverride string
}

func Render(opts RenderOptions) (out string, err error) {
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, opts)
	if err != nil {
		return
	}

	out = buf.String()
	return
}

type RenderDownloadURLOptions struct {
	Uname_s string
	Uname_m string
}

func RenderDownloadURL(downloadURLGoTemplate string, opts RenderDownloadURLOptions) (out string, err error) {
	t, err := texttemplate.New("").Parse(downloadURLGoTemplate)
	if err != nil {
		return
	}

	opts.Uname_s = uname.NormalizeUnameS(opts.Uname_s)
	opts.Uname_m = uname.NormalizeUnameM(opts.Uname_m)

	var buf bytes.Buffer
	err = t.Execute(&buf, opts)
	if err != nil {
		return
	}

	out = buf.String()
	return
}
