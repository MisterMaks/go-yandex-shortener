package cert_creator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	certPEMBytes, privateKeyPEMBytes, err := Create()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(certPEMBytes), "-----BEGIN CERTIFICATE-----\n"))
	assert.True(t, strings.HasSuffix(string(certPEMBytes), "\n-----END CERTIFICATE-----\n"))
	assert.True(t, strings.HasPrefix(string(privateKeyPEMBytes), "-----BEGIN RSA PRIVATE KEY-----\n"))
	assert.True(t, strings.HasSuffix(string(privateKeyPEMBytes), "\n-----END RSA PRIVATE KEY-----\n"))
}
