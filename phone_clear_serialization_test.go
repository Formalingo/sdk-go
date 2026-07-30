package formalingo_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	formalingo "github.com/Formalingo/sdk-go"
	sdkclient "github.com/Formalingo/sdk-go/client"
	sdkapi "github.com/Formalingo/sdk-go/client/api"
	"github.com/Formalingo/sdk-go/client/models"
	"github.com/google/uuid"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	serialization "github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/microsoft/kiota-abstractions-go/store"
	jsonserialization "github.com/microsoft/kiota-serialization-json-go"
)

func phonePayload(t *testing.T, model serialization.Parsable) map[string]any {
	t.Helper()
	writer := jsonserialization.NewJsonSerializationWriter()
	defer writer.Close()
	if err := writer.WriteObjectValue("", model); err != nil {
		t.Fatal(err)
	}
	raw, err := writer.GetSerializedContent()
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestRecipientClearPhoneSerialization(t *testing.T) {
	if _, ok := phonePayload(t, models.NewUpdateRecipientBody())["clearPhone"]; ok {
		t.Fatal("unset clearPhone must be omitted")
	}
	model := models.NewUpdateRecipientBody()
	value := true
	model.SetClearPhone(&value)
	if got := phonePayload(t, model)["clearPhone"]; got != true {
		t.Fatalf("clearPhone = %v, want true", got)
	}
}

func TestSignerClearPhoneSerialization(t *testing.T) {
	if _, ok := phonePayload(t, models.NewUpdateSignerBody())["clearPhone"]; ok {
		t.Fatal("unset clearPhone must be omitted")
	}
	model := models.NewUpdateSignerBody()
	value := true
	model.SetClearPhone(&value)
	if got := phonePayload(t, model)["clearPhone"]; got != true {
		t.Fatalf("clearPhone = %v, want true", got)
	}
}

func TestRecipientCreateResultDispatchSerialization(t *testing.T) {
	model := models.NewRecipientCreateResult()
	dispatchID := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	token := "one-time-token"
	link := "https://www.formalingo.com/f/one-time-token"
	password := "one-time-password"
	model.SetDispatchId(&dispatchID)
	model.SetToken(&token)
	model.SetLink(&link)
	model.SetPlainPassword(&password)

	payload := phonePayload(t, model)
	if got := payload["dispatchId"]; got != dispatchID.String() {
		t.Fatalf("dispatchId = %v, want %s", got, dispatchID)
	}
	if got := payload["token"]; got != token {
		t.Fatalf("token = %v, want %s", got, token)
	}
	if got := payload["link"]; got != link {
		t.Fatalf("link = %v, want %s", got, link)
	}
	if got := payload["plain_password"]; got != password {
		t.Fatalf("plain_password = %v, want %s", got, password)
	}
	if _, ok := payload["passwordHash"]; ok {
		t.Fatal("passwordHash must not be serialized")
	}
}

func TestDocumentSubmissionDispatchSerialization(t *testing.T) {
	submission := models.NewCreateSubmissionResult()
	submissionID := uuid.MustParse("00000000-0000-0000-0000-000000000041")
	dispatchID := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	reused := true
	linksCreated := true
	submission.SetSubmissionId(&submissionID)
	submission.SetDispatchId(&dispatchID)
	submission.SetDispatchReused(&reused)
	submission.SetLinksCreated(&linksCreated)

	payload := phonePayload(t, submission)
	if got := payload["submissionId"]; got != submissionID.String() {
		t.Fatalf("submissionId = %v, want %s", got, submissionID)
	}
	if got := payload["dispatchId"]; got != dispatchID.String() {
		t.Fatalf("dispatchId = %v, want %s", got, dispatchID)
	}
	if got := payload["dispatchReused"]; got != true {
		t.Fatalf("dispatchReused = %v, want true", got)
	}
	if got := payload["linksCreated"]; got != true {
		t.Fatalf("linksCreated = %v, want true", got)
	}

	signer := models.NewCreateSubmissionSignerResult()
	signerID := uuid.MustParse("00000000-0000-0000-0000-000000000043")
	label := "Buyer"
	role := "buyer"
	name := "Alice"
	color := "#13A373"
	order := int32(0)
	link := "https://www.formalingo.com/d/one-time-token"
	signer.SetId(&signerID)
	signer.SetLabel(&label)
	signer.SetRole(&role)
	signer.SetName(&name)
	signer.SetColor(&color)
	signer.SetOrder(&order)
	signer.SetLink(&link)

	signerPayload := phonePayload(t, signer)
	if got := signerPayload["link"]; got != link {
		t.Fatalf("link = %v, want %s", got, link)
	}
	for _, privateField := range []string{"token", "email", "phone", "passwordHash"} {
		if _, ok := signerPayload[privateField]; ok {
			t.Fatalf("%s must not be serialized", privateField)
		}
	}
}

func TestCreateDocumentSubmissionEmitsIdempotencyMetadataAndReturnsDataSigners(t *testing.T) {
	signerResult := models.NewCreateSubmissionSignerResult()
	signerID := uuid.MustParse("00000000-0000-0000-0000-000000000043")
	signerRole := "buyer"
	signerLink := "https://www.formalingo.com/d/one-time-token"
	signerResult.SetId(&signerID)
	signerResult.SetRole(&signerRole)
	signerResult.SetLink(&signerLink)

	result := models.NewCreateSubmissionResult()
	submissionID := uuid.MustParse("00000000-0000-0000-0000-000000000041")
	result.SetSubmissionId(&submissionID)
	result.SetSigners([]models.CreateSubmissionSignerResultable{signerResult})
	response := models.NewCreateSubmissionResponse()
	response.SetData(result)

	adapter := &capturingRequestAdapter{
		baseURL:       "https://example.test",
		response:      response,
		writerFactory: jsonserialization.NewJsonSerializationWriterFactory(),
	}
	client := sdkclient.NewFormalingoClient(adapter)
	signer := models.NewSignerInput()
	name := "Alice"
	email := "alice@example.com"
	signer.SetRole(&signerRole)
	signer.SetName(&name)
	signer.SetEmail(&email)
	body := models.NewCreateSubmissionBody()
	body.SetSigners([]models.SignerInputable{signer})

	submission, err := formalingo.CreateDocumentSubmission(
		context.Background(),
		client,
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		body,
		"document-create-1",
	)
	if err != nil {
		t.Fatal(err)
	}

	request := adapter.requestInfo
	if request == nil {
		t.Fatal("request metadata was not captured")
	}
	if request.Method != abstractions.POST {
		t.Fatalf("method = %s, want POST", request.Method)
	}
	uri, err := request.GetUri()
	if err != nil {
		t.Fatal(err)
	}
	wantURL := "https://example.test/api/v1/documents/00000000-0000-0000-0000-000000000001/submissions"
	if uri.String() != wantURL {
		t.Fatalf("url = %s, want %s", uri, wantURL)
	}
	headerValues := request.Headers.Get("Idempotency-Key")
	if len(headerValues) != 1 || headerValues[0] != "document-create-1" {
		t.Fatalf("Idempotency-Key = %v, want document-create-1", headerValues)
	}
	var requestBody map[string]any
	if err := json.Unmarshal(request.Content, &requestBody); err != nil {
		t.Fatal(err)
	}
	if got := requestBody["deliveryFormat"]; got != "document" {
		t.Fatalf("deliveryFormat = %v, want document", got)
	}
	signers, ok := requestBody["signers"].([]any)
	if !ok || len(signers) != 1 {
		t.Fatalf("signers = %#v, want one signer", requestBody["signers"])
	}
	emittedSigner := signers[0].(map[string]any)
	if emittedSigner["role"] != "buyer" ||
		emittedSigner["name"] != "Alice" ||
		emittedSigner["email"] != "alice@example.com" {
		t.Fatalf("emitted signer = %#v", emittedSigner)
	}
	if got := *submission.GetSigners()[0].GetLink(); got != signerLink {
		t.Fatalf("signer link = %s, want %s", got, signerLink)
	}
}

func TestCreateBulkRecipientsEmitsRequiredIdempotencyMetadataAndReturnsData(t *testing.T) {
	recipient := models.NewRecipientBulkCreateResult()
	recipientID := uuid.MustParse("00000000-0000-0000-0000-000000000043")
	label := "Alice"
	recipient.SetId(&recipientID)
	recipient.SetLabel(&label)
	response := sdkapi.NewV1FormsItemRecipientsBulkPostResponse()
	response.SetData([]models.RecipientBulkCreateResultable{recipient})

	adapter := &capturingRequestAdapter{
		baseURL:       "https://example.test",
		response:      response,
		writerFactory: jsonserialization.NewJsonSerializationWriterFactory(),
	}
	client := sdkclient.NewFormalingoClient(adapter)
	body := sdkapi.NewV1FormsItemRecipientsBulkPostRequestBody()
	confirmBulk := true
	sendNotifications := false
	input := sdkapi.NewV1FormsItemRecipientsBulkPostRequestBody_recipients()
	input.SetLabel(&label)
	body.SetConfirmBulk(&confirmBulk)
	body.SetSendNotifications(&sendNotifications)
	body.SetRecipients(
		[]sdkapi.V1FormsItemRecipientsBulkPostRequestBody_recipientsable{input},
	)

	recipients, err := formalingo.CreateBulkRecipients(
		context.Background(),
		client,
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		body,
		"recipient-bulk-1",
	)
	if err != nil {
		t.Fatal(err)
	}

	request := adapter.requestInfo
	if request == nil {
		t.Fatal("request metadata was not captured")
	}
	uri, err := request.GetUri()
	if err != nil {
		t.Fatal(err)
	}
	wantURL := "https://example.test/api/v1/forms/00000000-0000-0000-0000-000000000001/recipients/bulk"
	if uri.String() != wantURL {
		t.Fatalf("url = %s, want %s", uri, wantURL)
	}
	headerValues := request.Headers.Get("Idempotency-Key")
	if len(headerValues) != 1 || headerValues[0] != "recipient-bulk-1" {
		t.Fatalf("Idempotency-Key = %v, want recipient-bulk-1", headerValues)
	}
	if len(recipients) != 1 || recipients[0].GetLabel() == nil ||
		*recipients[0].GetLabel() != "Alice" {
		t.Fatalf("recipients = %#v, want Alice", recipients)
	}
}

func TestCreateBulkRecipientsRejectsInvalidIdempotencyMetadata(t *testing.T) {
	for _, value := range []string{"", "contains a space", strings.Repeat("a", 256)} {
		t.Run(value, func(t *testing.T) {
			_, err := formalingo.CreateBulkRecipients(
				context.Background(),
				nil,
				uuid.Nil,
				nil,
				value,
			)
			if err == nil {
				t.Fatal("expected invalid idempotency metadata error")
			}
		})
	}
}

func TestCreateDocumentSubmissionRejectsInvalidIdempotencyMetadata(t *testing.T) {
	_, err := formalingo.CreateDocumentSubmission(
		context.Background(),
		nil,
		uuid.Nil,
		nil,
		"contains a space",
	)
	if err == nil {
		t.Fatal("expected invalid idempotency metadata error")
	}
}

type capturingRequestAdapter struct {
	baseURL       string
	requestInfo   *abstractions.RequestInformation
	response      serialization.Parsable
	writerFactory serialization.SerializationWriterFactory
}

func (adapter *capturingRequestAdapter) Send(
	ctx context.Context,
	requestInfo *abstractions.RequestInformation,
	constructor serialization.ParsableFactory,
	errorMappings abstractions.ErrorMappings,
) (serialization.Parsable, error) {
	adapter.requestInfo = requestInfo
	return adapter.response, nil
}

func (adapter *capturingRequestAdapter) SendEnum(
	context.Context,
	*abstractions.RequestInformation,
	serialization.EnumFactory,
	abstractions.ErrorMappings,
) (any, error) {
	return nil, nil
}

func (adapter *capturingRequestAdapter) SendCollection(
	context.Context,
	*abstractions.RequestInformation,
	serialization.ParsableFactory,
	abstractions.ErrorMappings,
) ([]serialization.Parsable, error) {
	return nil, nil
}

func (adapter *capturingRequestAdapter) SendEnumCollection(
	context.Context,
	*abstractions.RequestInformation,
	serialization.EnumFactory,
	abstractions.ErrorMappings,
) ([]any, error) {
	return nil, nil
}

func (adapter *capturingRequestAdapter) SendPrimitive(
	context.Context,
	*abstractions.RequestInformation,
	string,
	abstractions.ErrorMappings,
) (any, error) {
	return nil, nil
}

func (adapter *capturingRequestAdapter) SendPrimitiveCollection(
	context.Context,
	*abstractions.RequestInformation,
	string,
	abstractions.ErrorMappings,
) ([]any, error) {
	return nil, nil
}

func (adapter *capturingRequestAdapter) SendNoContent(
	context.Context,
	*abstractions.RequestInformation,
	abstractions.ErrorMappings,
) error {
	return nil
}

func (adapter *capturingRequestAdapter) GetSerializationWriterFactory() serialization.SerializationWriterFactory {
	return adapter.writerFactory
}

func (adapter *capturingRequestAdapter) EnableBackingStore(store.BackingStoreFactory) {
}

func (adapter *capturingRequestAdapter) SetBaseUrl(baseURL string) {
	adapter.baseURL = baseURL
}

func (adapter *capturingRequestAdapter) GetBaseUrl() string {
	return adapter.baseURL
}

func (adapter *capturingRequestAdapter) ConvertToNativeRequest(
	context.Context,
	*abstractions.RequestInformation,
) (any, error) {
	return nil, nil
}
