package comparator

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"http pública", "http://8.8.8.8/foo", false},
		{"https pública", "https://8.8.8.8/foo", false},
		{"dominio público", "https://example.com", false},
		{"IPv6 pública", "http://[2001:4860:4860::8888]/", false},
		{"scheme inválido", "ftp://example.com", true},
		{"sin host", "http:///foo", true},
		{"URL inválida", "http://%zz", true},
		{"loopback IPv4", "http://127.0.0.1/", true},
		{"localhost", "http://localhost/", true},
		{"link-local metadata", "http://169.254.169.254/latest/meta-data", true},
		{"privada 10", "http://10.0.0.1/", true},
		{"privada 172.16", "http://172.16.0.1/", true},
		{"privada 192.168", "http://192.168.1.1/", true},
		{"CGNAT", "http://100.64.0.1/", true},
		{"IPv6 loopback", "http://[::1]/", true},
		{"IPv6 privada", "http://[fc00::1]/", true},
		{"IPv6 link-local", "http://[fe80::1]/", true},
		{"IPv6 unspecified", "http://[::]/", true},
		{"multicast", "http://224.0.0.1/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsBlockedIP(t *testing.T) {
	assert.True(t, isBlockedIP(net.ParseIP("127.0.0.1")))
	assert.True(t, isBlockedIP(net.ParseIP("10.0.0.1")))
	assert.True(t, isBlockedIP(net.ParseIP("169.254.169.254")))
	assert.True(t, isBlockedIP(net.ParseIP("100.64.0.1")))
	assert.False(t, isBlockedIP(net.ParseIP("8.8.8.8")))
}
