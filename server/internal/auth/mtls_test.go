package auth

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// ---------------------------------------------------------------------------
// NewCertificateAuthority
// ---------------------------------------------------------------------------

func TestNewCertificateAuthority_Success(t *testing.T) {
	ca, err := NewCertificateAuthority()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ca == nil {
		t.Fatal("expected non-nil CA")
	}
}

func TestNewCertificateAuthority_CertPEMNonEmpty(t *testing.T) {
	ca, err := NewCertificateAuthority()
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	if len(ca.caCertPEM) == 0 {
		t.Fatal("expected non-empty CA cert PEM")
	}
}

func TestNewCertificateAuthority_KeyPEMNonEmpty(t *testing.T) {
	ca, err := NewCertificateAuthority()
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	if len(ca.caKeyPEM) == 0 {
		t.Fatal("expected non-empty CA key PEM")
	}
}

func TestNewCertificateAuthority_ValidCACert(t *testing.T) {
	ca, err := NewCertificateAuthority()
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	if ca.caCert == nil {
		t.Fatal("expected non-nil parsed CA certificate")
	}
	if !ca.caCert.IsCA {
		t.Fatal("expected CA flag to be true")
	}
	if ca.caCert.Subject.CommonName != "FleetML Root CA" {
		t.Fatalf("expected CN 'FleetML Root CA', got %q", ca.caCert.Subject.CommonName)
	}
}

// ---------------------------------------------------------------------------
// IssueCertificate
// ---------------------------------------------------------------------------

func newTestCA(t *testing.T) *CertificateAuthority {
	t.Helper()
	ca, err := NewCertificateAuthority()
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	return ca
}

func TestIssueCertificate_ValidDeviceID(t *testing.T) {
	ca := newTestCA(t)
	certPEM, keyPEM, err := ca.IssueCertificate("device-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(certPEM) == 0 {
		t.Fatal("expected non-empty cert PEM")
	}
	if len(keyPEM) == 0 {
		t.Fatal("expected non-empty key PEM")
	}
}

func TestIssueCertificate_CNMatchesDeviceID(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, err := ca.IssueCertificate("edge-sensor-42")
	if err != nil {
		t.Fatalf("issue cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode cert PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	if cert.Subject.CommonName != "edge-sensor-42" {
		t.Fatalf("expected CN 'edge-sensor-42', got %q", cert.Subject.CommonName)
	}
}

func TestIssueCertificate_SANContainsDeviceID(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, err := ca.IssueCertificate("device-san-test")
	if err != nil {
		t.Fatalf("issue cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	found := false
	for _, dns := range cert.DNSNames {
		if dns == "device-san-test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected SAN to contain 'device-san-test', got DNS names: %v", cert.DNSNames)
	}
}

func TestIssueCertificate_SignedByCA(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, err := ca.IssueCertificate("device-verify")
	if err != nil {
		t.Fatalf("issue cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	// Verify against the CA.
	pool := x509.NewCertPool()
	pool.AddCert(ca.caCert)
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	if err != nil {
		t.Fatalf("certificate verification failed: %v", err)
	}
}

func TestIssueCertificate_HasClientAuthUsage(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, _ := ca.IssueCertificate("device-usage")

	block, _ := pem.Decode(certPEM)
	cert, _ := x509.ParseCertificate(block.Bytes)

	found := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageClientAuth {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected ExtKeyUsageClientAuth in device certificate")
	}
}

// ---------------------------------------------------------------------------
// IssueServerCertificate
// ---------------------------------------------------------------------------

func TestIssueServerCertificate_ValidHostnames(t *testing.T) {
	ca := newTestCA(t)
	certPEM, keyPEM, err := ca.IssueServerCertificate([]string{"api.fleetml.io", "localhost"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(certPEM) == 0 {
		t.Fatal("expected non-empty server cert PEM")
	}
	if len(keyPEM) == 0 {
		t.Fatal("expected non-empty server key PEM")
	}
}

func TestIssueServerCertificate_SANMatchesHostnames(t *testing.T) {
	ca := newTestCA(t)
	hostnames := []string{"api.fleetml.io", "grpc.fleetml.io", "localhost"}
	certPEM, _, err := ca.IssueServerCertificate(hostnames)
	if err != nil {
		t.Fatalf("issue server cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	for _, expected := range hostnames {
		found := false
		for _, dns := range cert.DNSNames {
			if dns == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected SAN to contain %q, got DNS names: %v", expected, cert.DNSNames)
		}
	}
}

func TestIssueServerCertificate_HasServerAuthUsage(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, _ := ca.IssueServerCertificate([]string{"localhost"})

	block, _ := pem.Decode(certPEM)
	cert, _ := x509.ParseCertificate(block.Bytes)

	found := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected ExtKeyUsageServerAuth in server certificate")
	}
}

func TestIssueServerCertificate_SignedByCA(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, _ := ca.IssueServerCertificate([]string{"localhost"})

	block, _ := pem.Decode(certPEM)
	cert, _ := x509.ParseCertificate(block.Bytes)

	pool := x509.NewCertPool()
	pool.AddCert(ca.caCert)
	_, err := cert.Verify(x509.VerifyOptions{
		Roots:     pool,
		DNSName:   "localhost",
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		t.Fatalf("server certificate verification failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ServerTLSConfig
// ---------------------------------------------------------------------------

func TestServerTLSConfig_ReturnsConfig(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey, err := ca.IssueServerCertificate([]string{"localhost"})
	if err != nil {
		t.Fatalf("issue server cert: %v", err)
	}

	tlsCfg, err := ca.ServerTLSConfig(serverCert, serverKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tlsCfg == nil {
		t.Fatal("expected non-nil TLS config")
	}
}

func TestServerTLSConfig_RequiresClientCerts(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey, _ := ca.IssueServerCertificate([]string{"localhost"})

	tlsCfg, err := ca.ServerTLSConfig(serverCert, serverKey)
	if err != nil {
		t.Fatalf("server TLS config: %v", err)
	}

	if tlsCfg.ClientAuth != 4 { // tls.RequireAndVerifyClientCert == 4
		t.Fatalf("expected RequireAndVerifyClientCert (4), got %d", tlsCfg.ClientAuth)
	}
}

func TestServerTLSConfig_TLS13Minimum(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey, _ := ca.IssueServerCertificate([]string{"localhost"})

	tlsCfg, err := ca.ServerTLSConfig(serverCert, serverKey)
	if err != nil {
		t.Fatalf("server TLS config: %v", err)
	}

	if tlsCfg.MinVersion != 0x0304 { // tls.VersionTLS13 == 0x0304
		t.Fatalf("expected MinVersion TLS 1.3 (0x0304), got 0x%04x", tlsCfg.MinVersion)
	}
}

func TestServerTLSConfig_HasCertificate(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey, _ := ca.IssueServerCertificate([]string{"localhost"})

	tlsCfg, err := ca.ServerTLSConfig(serverCert, serverKey)
	if err != nil {
		t.Fatalf("server TLS config: %v", err)
	}

	if len(tlsCfg.Certificates) == 0 {
		t.Fatal("expected at least one certificate in TLS config")
	}
}

func TestServerTLSConfig_InvalidCertReturnsError(t *testing.T) {
	ca := newTestCA(t)
	_, err := ca.ServerTLSConfig([]byte("bad-cert"), []byte("bad-key"))
	if err == nil {
		t.Fatal("expected error for invalid cert/key pair")
	}
}

func TestServerTLSConfig_HasClientCAs(t *testing.T) {
	ca := newTestCA(t)
	serverCert, serverKey, _ := ca.IssueServerCertificate([]string{"localhost"})

	tlsCfg, err := ca.ServerTLSConfig(serverCert, serverKey)
	if err != nil {
		t.Fatalf("server TLS config: %v", err)
	}

	if tlsCfg.ClientCAs == nil {
		t.Fatal("expected non-nil ClientCAs pool")
	}
}

// ---------------------------------------------------------------------------
// CACertPEM
// ---------------------------------------------------------------------------

func TestCACertPEM_NonEmpty(t *testing.T) {
	ca := newTestCA(t)
	pem := ca.CACertPEM()
	if len(pem) == 0 {
		t.Fatal("expected non-empty CA cert PEM")
	}
}

func TestCACertPEM_ValidPEMFormat(t *testing.T) {
	ca := newTestCA(t)
	pemBytes := ca.CACertPEM()

	block, rest := pem.Decode(pemBytes)
	if block == nil {
		t.Fatal("failed to decode PEM block")
	}
	if block.Type != "CERTIFICATE" {
		t.Fatalf("expected block type CERTIFICATE, got %q", block.Type)
	}
	if len(block.Bytes) == 0 {
		t.Fatal("expected non-empty PEM block bytes")
	}
	// No extra data after the single PEM block.
	trimmed := 0
	for _, b := range rest {
		if b != '\n' && b != '\r' && b != ' ' {
			trimmed++
		}
	}
	if trimmed > 0 {
		t.Fatal("expected no extra data after PEM block")
	}
}

func TestCACertPEM_Parseable(t *testing.T) {
	ca := newTestCA(t)
	pemBytes := ca.CACertPEM()

	block, _ := pem.Decode(pemBytes)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("expected parseable certificate, got %v", err)
	}
	if !cert.IsCA {
		t.Fatal("expected certificate to be a CA")
	}
}

// ---------------------------------------------------------------------------
// ExtractDeviceIDFromCert
// ---------------------------------------------------------------------------

func TestExtractDeviceIDFromCert_ValidCert(t *testing.T) {
	ca := newTestCA(t)
	certPEM, _, err := ca.IssueCertificate("my-edge-device")
	if err != nil {
		t.Fatalf("issue cert: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	deviceID := ExtractDeviceIDFromCert(cert)
	if deviceID != "my-edge-device" {
		t.Fatalf("expected 'my-edge-device', got %q", deviceID)
	}
}

func TestExtractDeviceIDFromCert_NilCert(t *testing.T) {
	// ExtractDeviceIDFromCert accesses cert.Subject.CommonName, so a nil cert
	// will panic. We verify it panics (documenting the behavior).
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil certificate")
		}
	}()
	ExtractDeviceIDFromCert(nil)
}

func TestExtractDeviceIDFromCert_EmptyCN(t *testing.T) {
	// Create a certificate with an empty CommonName.
	cert := &x509.Certificate{}
	deviceID := ExtractDeviceIDFromCert(cert)
	if deviceID != "" {
		t.Fatalf("expected empty device ID, got %q", deviceID)
	}
}

// ---------------------------------------------------------------------------
// Multiple certs from same CA
// ---------------------------------------------------------------------------

func TestMultipleCertificatesDifferentSerials(t *testing.T) {
	ca := newTestCA(t)
	cert1PEM, _, _ := ca.IssueCertificate("device-a")
	cert2PEM, _, _ := ca.IssueCertificate("device-b")

	block1, _ := pem.Decode(cert1PEM)
	c1, _ := x509.ParseCertificate(block1.Bytes)

	block2, _ := pem.Decode(cert2PEM)
	c2, _ := x509.ParseCertificate(block2.Bytes)

	if c1.SerialNumber.Cmp(c2.SerialNumber) == 0 {
		t.Fatal("expected different serial numbers for different certificates")
	}
}
