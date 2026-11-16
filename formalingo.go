package formalingo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Formalingo/sdk-go/client"
	abs "github.com/microsoft/kiota-abstractions-go"
	auth "github.com/microsoft/kiota-abstractions-go/authentication"
	http "github.com/microsoft/kiota-http-go"
)

func NewClient(apiKey string, baseURL ...string) (*client.FormalingoClient, error) {
	base := "https://formalingo.com"
	if len(baseURL) > 0 {
		base = baseURL[0]
	}
	parsedBase, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("invalid API base URL")
	}
	origin, err := apiURLOrigin(parsedBase)
	if err != nil {
		return nil, fmt.Errorf("invalid API base URL: %w", err)
	}

	authProvider, err := auth.NewApiKeyAuthenticationProvider(
		fmt.Sprintf("Bearer %s", apiKey),
		"Authorization",
		auth.HEADER_KEYLOCATION,
	)
	if err != nil {
		return nil, fmt.Errorf("creating auth provider: %w", err)
	}

	// Formalingo accepts plain JSON request bodies. Keep Kiota's default
	// middleware, but disable its automatic request-body gzip compression.
	compression := http.NewCompressionOptions(false)
	middleware, err := http.GetDefaultMiddlewaresWithOptions(&compression)
	if err != nil {
		return nil, fmt.Errorf("configuring request middleware: %w", err)
	}
	nativeClient := http.GetDefaultClient(middleware...)
	adapter, err := http.NewNetHttpRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
		&originAuthenticationProvider{provider: authProvider, origin: origin}, nil, nil, nativeClient,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request adapter: %w", err)
	}
	adapter.SetBaseUrl(base)

	return client.NewFormalingoClient(adapter), nil
}

type apiOrigin struct {
	scheme   string
	hostname string
	port     string
}

func apiURLOrigin(uri *url.URL) (apiOrigin, error) {
	scheme := strings.ToLower(uri.Scheme)
	if (scheme != "http" && scheme != "https") || uri.Hostname() == "" || uri.User != nil {
		return apiOrigin{}, fmt.Errorf("an absolute HTTP(S) URL without user information is required")
	}
	port := uri.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	} else {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return apiOrigin{}, fmt.Errorf("invalid port")
		}
		port = strconv.Itoa(number)
	}
	return apiOrigin{scheme: scheme, hostname: strings.ToLower(uri.Hostname()), port: port}, nil
}

// Generated WithUrl builders may override the target. Check the complete API
// origin before the underlying provider can add the API key to a request.
type originAuthenticationProvider struct {
	provider auth.AuthenticationProvider
	origin   apiOrigin
}

func (provider *originAuthenticationProvider) AuthenticateRequest(ctx context.Context, request *abs.RequestInformation, additionalContext map[string]interface{}) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	uri, err := request.GetUri()
	if err != nil {
		return err
	}
	origin, err := apiURLOrigin(uri)
	if err != nil || origin != provider.origin {
		return fmt.Errorf("request URL origin does not match the configured API origin")
	}
	return provider.provider.AuthenticateRequest(ctx, request, additionalContext)
}
