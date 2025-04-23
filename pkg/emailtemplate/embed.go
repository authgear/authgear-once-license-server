package emailtemplate

import (
	_ "embed"
	htmltemplate "html/template"
	"strings"
)

//go:embed installation_email.gotemplate
var installationEmailString string

var installationEmail *htmltemplate.Template

func init() {
	t, err := htmltemplate.New("").Parse(installationEmailString)
	if err != nil {
		panic(err)
	}
	installationEmail = t
}

type InstallationEmailData struct {
	InstallationOneliner string
}

func RenderInstallationEmail(data InstallationEmailData) string {
	var buf strings.Builder
	err := installationEmail.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
