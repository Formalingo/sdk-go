package formalingo

import (
	"context"
	"fmt"

	"github.com/Formalingo/sdk-go/client"
	sdkapi "github.com/Formalingo/sdk-go/client/api"
	"github.com/Formalingo/sdk-go/client/models"
	"github.com/google/uuid"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	auth "github.com/microsoft/kiota-abstractions-go/authentication"
	http "github.com/microsoft/kiota-http-go"
)

func NewClient(apiKey string, baseURL ...string) (*client.FormalingoClient, error) {
	url := "https://app.formalingo.com"
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

// CreateBulkRecipients creates up to 100 recipients with caller-owned retry
// metadata. Reuse the idempotency key only for an exact retry of the same
// request body.
func CreateBulkRecipients(
	ctx context.Context,
	sdkClient *client.FormalingoClient,
	formID uuid.UUID,
	body sdkapi.V1FormsItemRecipientsBulkPostRequestBodyable,
	idempotencyKey string,
) ([]models.RecipientBulkCreateResultable, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}
	if sdkClient == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}
	if body == nil {
		return nil, fmt.Errorf("body cannot be nil")
	}

	headers := abstractions.NewRequestHeaders()
	headers.TryAdd("Idempotency-Key", idempotencyKey)
	requestConfiguration := &abstractions.RequestConfiguration[abstractions.DefaultQueryParameters]{
		Headers: headers,
	}
	response, err := sdkClient.Api().V1().Forms().ById(formID).
		Recipients().Bulk().Post(ctx, body, requestConfiguration)
	if err != nil {
		return nil, err
	}
	if response == nil || response.GetData() == nil {
		return nil, fmt.Errorf("Formalingo returned no bulk recipient data")
	}

	return response.GetData(), nil
}

// CreateDocumentSubmission creates a retry-safe signing submission and returns
// the response data. Reuse an idempotency key only when retrying the same
// logical request.
func CreateDocumentSubmission(
	ctx context.Context,
	sdkClient *client.FormalingoClient,
	documentID uuid.UUID,
	body models.CreateSubmissionBodyable,
	idempotencyKey string,
) (models.CreateSubmissionResultable, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}
	if sdkClient == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}
	if body == nil {
		return nil, fmt.Errorf("body cannot be nil")
	}

	headers := abstractions.NewRequestHeaders()
	headers.TryAdd("Idempotency-Key", idempotencyKey)
	requestConfiguration := &abstractions.RequestConfiguration[abstractions.DefaultQueryParameters]{
		Headers: headers,
	}
	response, err := sdkClient.Api().V1().Documents().ById(documentID).
		Submissions().Post(ctx, body, requestConfiguration)
	if err != nil {
		return nil, err
	}
	if response == nil || response.GetData() == nil {
		return nil, fmt.Errorf("Formalingo returned no document submission data")
	}

	return response.GetData(), nil
}

func validateIdempotencyKey(idempotencyKey string) error {
	if len(idempotencyKey) == 0 || len(idempotencyKey) > 255 {
		return fmt.Errorf("idempotencyKey must contain 1-255 printable ASCII characters without spaces")
	}
	for index := 0; index < len(idempotencyKey); index++ {
		if idempotencyKey[index] < 0x21 || idempotencyKey[index] > 0x7e {
			return fmt.Errorf("idempotencyKey must contain 1-255 printable ASCII characters without spaces")
		}
	}
	return nil
}
