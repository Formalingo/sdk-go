package formalingo_test

import (
	"encoding/json"
	"testing"

	"github.com/Formalingo/sdk-go/client/models"
	"github.com/google/uuid"
	serialization "github.com/microsoft/kiota-abstractions-go/serialization"
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
