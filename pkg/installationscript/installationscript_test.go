package installationscript

import (
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		opts     RenderOptions
		expected string
	}{
		{
			name: "Without --image",
			opts: RenderOptions{
				DownloadURL: "https://example.com/download",
				LicenseKey:  "license-123",
			},
			expected: `#!/bin/sh
set -e

echo "Installing the Authgear ONCE command......"
echo "This script uses sudo, you will be prompted for authentication."
sudo true

download_url="https://example.com/download?uname_s=$(uname -s)&uname_m=$(uname -m)"
tmp_path="$(mktemp)"
curl -fsSL "$download_url" > "$tmp_path"
sudo mv "$tmp_path" /usr/local/bin/authgear-once
sudo chmod u+x /usr/local/bin/authgear-once

if [ "$(uname -s)" = "Darwin" ]; then
	/usr/local/bin/authgear-once setup  "license-123"
else
	sudo /usr/local/bin/authgear-once setup  "license-123"
fi
`,
		},
		{
			name: "With --image",
			opts: RenderOptions{
				DownloadURL:   "https://example.com/download",
				LicenseKey:    "license-abc",
				ImageOverride: "some-docker-registry.com/authgear-once:1.0.0",
			},
			expected: `#!/bin/sh
set -e

echo "Installing the Authgear ONCE command......"
echo "This script uses sudo, you will be prompted for authentication."
sudo true

download_url="https://example.com/download?uname_s=$(uname -s)&uname_m=$(uname -m)"
tmp_path="$(mktemp)"
curl -fsSL "$download_url" > "$tmp_path"
sudo mv "$tmp_path" /usr/local/bin/authgear-once
sudo chmod u+x /usr/local/bin/authgear-once

if [ "$(uname -s)" = "Darwin" ]; then
	/usr/local/bin/authgear-once setup --image 'some-docker-registry.com/authgear-once:1.0.0' "license-abc"
else
	sudo /usr/local/bin/authgear-once setup --image 'some-docker-registry.com/authgear-once:1.0.0' "license-abc"
fi
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.opts)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Render() output mismatch\nGot:\n%s\nWant:\n%s", result, tt.expected)
			}
		})
	}
}
