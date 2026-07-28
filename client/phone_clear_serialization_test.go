package client_test

import (
    "encoding/json"
    "testing"

    jsonserialization "github.com/microsoft/kiota-serialization-json-go"
    "github.com/Formalingo/sdk-go/client/models"
)

func assertClearPhone(t *testing.T, model models.UpdateRecipientBodyable, clear bool) {
    writer := jsonserialization.NewJsonSerializationWriter()
    defer writer.Close()
    if err := writer.WriteObjectValue("", model); err != nil { t.Fatal(err) }
    raw, err := writer.GetSerializedContent(); if err != nil { t.Fatal(err) }
    var payload map[string]any
    if err := json.Unmarshal(raw, &payload); err != nil { t.Fatal(err) }
    _, present := payload["clearPhone"]
    if present != clear { t.Fatalf("clearPhone presence = %v, want %v", present, clear) }
    if clear && payload["clearPhone"] != true { t.Fatalf("clearPhone = %v", payload["clearPhone"]) }
}

func TestRecipientClearPhoneSerialization(t *testing.T) {
    omitted := models.NewUpdateRecipientBody(); assertClearPhone(t, omitted, false)
    explicit := models.NewUpdateRecipientBody(); value := true; explicit.SetClearPhone(&value); assertClearPhone(t, explicit, true)
}

func TestSignerClearPhoneSerialization(t *testing.T) {
    writer := jsonserialization.NewJsonSerializationWriter(); defer writer.Close()
    omitted := models.NewUpdateSignerBody(); if err := writer.WriteObjectValue("", omitted); err != nil { t.Fatal(err) }
    raw, _ := writer.GetSerializedContent(); var payload map[string]any; _ = json.Unmarshal(raw, &payload); if _, ok := payload["clearPhone"]; ok { t.Fatal("clearPhone should be omitted") }
    writer = jsonserialization.NewJsonSerializationWriter(); explicit := models.NewUpdateSignerBody(); value := true; explicit.SetClearPhone(&value); if err := writer.WriteObjectValue("", explicit); err != nil { t.Fatal(err) }; raw, _ = writer.GetSerializedContent(); _ = json.Unmarshal(raw, &payload); if payload["clearPhone"] != true { t.Fatal("clearPhone should be true") }
}
