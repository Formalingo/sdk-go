package formalingo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Formalingo/sdk-go/client/models"
	"github.com/google/uuid"
)

func TestNewClientSendsPlainJSONOverHTTPS(t *testing.T) {
	type observedRequest struct {
		body            []byte
		contentType     string
		contentEncoding string
		authorization   string
		host            string
		path            string
		method          string
		tls             bool
	}
	requests := make(chan observedRequest, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		requests <- observedRequest{
			body: body, contentType: request.Header.Get("Content-Type"),
			contentEncoding: request.Header.Get("Content-Encoding"),
			authorization:   request.Header.Get("Authorization"), host: request.Host,
			path: request.URL.Path, method: request.Method, tls: request.TLS != nil,
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"success":true,"data":{"id":"00000000-0000-0000-0000-000000000001","isRequired":true,"options":{"supporting_attachment":true},"pageNumber":0,"x":0,"y":0,"width":0,"height":0}}`)
	}))
	t.Cleanup(server.Close)

	// Trust only this test server's certificate for the public factory's client.
	// Restoring the process default keeps this test isolated; do not run it in parallel.
	previousTransport := http.DefaultTransport
	trustedTransport := http.DefaultTransport.(*http.Transport).Clone()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	trustedTransport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	http.DefaultTransport = trustedTransport
	t.Cleanup(func() {
		http.DefaultTransport = previousTransport
		trustedTransport.CloseIdleConnections()
	})

	client, err := NewClient("synthetic-test-key", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	body := models.NewCreateFieldBody()
	label, role := "Proof of address", "Signer 1"
	fieldType, supporting := models.FILE_UPLOAD_CREATEFIELDBODY_TYPE, true
	body.SetLabel(&label)
	body.SetAssigneeRole(&role)
	body.SetTypeEscaped(&fieldType)
	options := models.NewCreateFieldBody_options()
	options.SetSupportingAttachment(&supporting)
	body.SetOptions(options)
	documentID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	response, err := client.Api().V1().Documents().ById(documentID).Fields().Post(context.Background(), body, nil)
	if err != nil {
		t.Fatal(err)
	}
	request := <-requests
	origin, _ := url.Parse(server.URL)
	if !request.tls || request.host != origin.Host || request.method != http.MethodPost || request.path != "/api/v1/documents/"+documentID.String()+"/fields" {
		t.Fatal("factory sent request to an unexpected TLS target")
	}
	if request.authorization != "Bearer synthetic-test-key" {
		t.Fatal("factory omitted API key authentication")
	}
	if request.contentType != "application/json" || request.contentEncoding != "" {
		t.Fatalf("API requires plain JSON; Content-Type = %q, Content-Encoding = %q", request.contentType, request.contentEncoding)
	}
	var serialized map[string]any
	if err := json.Unmarshal(request.body, &serialized); err != nil {
		t.Fatalf("request body is not plain JSON: %v", err)
	}
	if serialized["options"].(map[string]any)["supporting_attachment"] != true {
		t.Fatal("typed supporting attachment flag was omitted")
	}
	for _, name := range []string{"isRequired", "pageNumber", "x", "y", "width", "height"} {
		if _, present := serialized[name]; present {
			t.Fatalf("omitted property %s received a serialized default", name)
		}
	}
	if response == nil || response.GetData() == nil || response.GetData().GetIsRequired() == nil || !*response.GetData().GetIsRequired() {
		t.Fatal("required response did not deserialize")
	}
}

type testRoundTripper func(*http.Request) (*http.Response, error)

func (transport testRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func TestNewClientRejectsOffOriginURLOverridesBeforeSendingCredentials(t *testing.T) {
	for name, override := range map[string]string{
		"hostname": "https://other.invalid/api/v1/documents",
		"scheme":   "http://sdk-proof.invalid/api/v1/documents",
		"port":     "https://sdk-proof.invalid:444/api/v1/documents",
	} {
		t.Run(name, func(t *testing.T) {
			previousTransport := http.DefaultTransport
			calls := 0
			http.DefaultTransport = testRoundTripper(func(request *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{
					StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(strings.NewReader(`{"success":true,"data":[]}`)), Request: request,
				}, nil
			})
			t.Cleanup(func() { http.DefaultTransport = previousTransport })
			client, err := NewClient("synthetic-test-key", "https://sdk-proof.invalid")
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Api().V1().Documents().WithUrl(override).Get(context.Background(), nil)
			if err == nil || !strings.Contains(err.Error(), "origin") {
				t.Fatal("off-origin override must fail before authentication")
			}
			if calls != 0 {
				t.Fatal("off-origin override reached the transport")
			}
		})
	}
}

func TestNewClientAllowsEquivalentHTTPSOrigin(t *testing.T) {
	previousTransport := http.DefaultTransport
	calls := 0
	http.DefaultTransport = testRoundTripper(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Header.Get("Authorization") != "Bearer synthetic-test-key" {
			t.Fatal("same-origin request omitted authentication")
		}
		return &http.Response{
			StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{"success":true,"data":[]}`)), Request: request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	client, err := NewClient("synthetic-test-key", "https://sdk-proof.invalid")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Api().V1().Documents().WithUrl("https://SDK-PROOF.INVALID:443/api/v1/documents").Get(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("equivalent HTTPS origin did not reach the transport")
	}
}
