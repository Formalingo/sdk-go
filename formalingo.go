package formalingo

import (
	"fmt"

	"github.com/Formalingo/sdk-go/client"
	auth "github.com/microsoft/kiota-abstractions-go/authentication"
	http "github.com/microsoft/kiota-http-go"
)

func NewClient(apiKey string, baseURL ...string) (*client.FormalingoClient, error) {
	url := "https://formalingo.com"
	if len(baseURL) > 0 {
		url = baseURL[0]
	}

	authProvider, err := auth.NewApiKeyAuthenticationProvider(
		fmt.Sprintf("Bearer %s", apiKey),
		"Authorization",
		auth.HEADER_KEYLOCATION,
	)
	if err != nil {
		return nil, fmt.Errorf("creating auth provider: %w", err)
	}

	adapter, err := http.NewNetHttpRequestAdapter(authProvider)
	if err != nil {
		return nil, fmt.Errorf("creating request adapter: %w", err)
	}
	adapter.SetBaseUrl(url)

	return client.NewFormalingoClient(adapter), nil
}
