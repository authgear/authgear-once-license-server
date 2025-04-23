package emailtemplate

import (
	"regexp"
	"testing"
)

func TestInstallationEmailString(t *testing.T) {
	if installationEmailString == "" {
		t.Errorf("expected installationEmailString to be non-empty")
	}
}

func TestInstallationEmail(t *testing.T) {
	if installationEmail == nil {
		t.Errorf("expected installationEmail to be non-nil")
	}
}

func TestRenderInstallationEmail(t *testing.T) {
	data := InstallationEmailData{
		InstallationOneliner: "/bin/bash",
	}

	s := RenderInstallationEmail(data)

	if s == "" {
		t.Errorf("expected result to be non-empty")
	}

	matched, err := regexp.MatchString(data.InstallationOneliner, s)
	if err != nil {
		t.Errorf("expected err to be nil, but it was %v", err)
	}
	if !matched {
		t.Errorf("expected InstallationOneliner to be present")
	}
}
