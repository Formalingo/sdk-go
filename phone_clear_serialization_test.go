package formalingo_test

import (
    "encoding/json"
    "testing"

    "github.com/Formalingo/sdk-go/client/models"
    serialization "github.com/microsoft/kiota-abstractions-go/serialization"
    jsonserialization "github.com/microsoft/kiota-serialization-json-go"
)

func phonePayload(t *testing.T, model serialization.Parsable) map[string]any {
    t.Helper()
    writer := jsonserialization.NewJsonSerializationWriter()
    defer writer.Close()
    if err := writer.WriteObjectValue("", model); err != nil { t.Fatal(err) }
    raw, err := writer.GetSerializedContent(); if err != nil { t.Fatal(err) }
    var payload map[string]any
    if err := json.Unmarshal(raw, &payload); err != nil { t.Fatal(err) }
    return payload
}

func TestRecipientClearPhoneSerialization(t *testing.T) {
    if _, ok := phonePayload(t, models.NewUpdateRecipientBody())["clearPhone"]; ok { t.Fatal("unset clearPhone must be omitted") }
    model := models.NewUpdateRecipientBody(); value := true; model.SetClearPhone(&value)
    if got := phonePayload(t, model)["clearPhone"]; got != true { t.Fatalf("clearPhone = %v, want true", got) }
}

func TestSignerClearPhoneSerialization(t *testing.T) {
    if _, ok := phonePayload(t, models.NewUpdateSignerBody())["clearPhone"]; ok { t.Fatal("unset clearPhone must be omitted") }
    model := models.NewUpdateSignerBody(); value := true; model.SetClearPhone(&value)
    if got := phonePayload(t, model)["clearPhone"]; got != true { t.Fatalf("clearPhone = %v, want true", got) }
}
