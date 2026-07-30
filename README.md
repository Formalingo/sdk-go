# Formalingo Go SDK

Official Go SDK for the [Formalingo API](https://formalingo.com/docs), generated with [Microsoft Kiota](https://learn.microsoft.com/en-us/openapi/kiota/).

## Installation

```bash
go get github.com/Formalingo/sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    formalingo "github.com/Formalingo/sdk-go"
)

func main() {
    client, err := formalingo.NewClient("af_live_YOUR_KEY")
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    // List forms
    forms, err := client.Api().V1().Forms().Get(ctx, nil)

    // Create a form
    body := &models.CreateFormBody{
        Title:       to.Ptr("Customer Satisfaction Survey"),
        Description: to.Ptr("Help us improve our service."),
    }
    form, err := client.Api().V1().Forms().Post(ctx, body, nil)

    // Get a specific form
    details, err := client.Api().V1().Forms().ByFormId("FORM_ID").Get(ctx, nil)

    // Delete a form
    err = client.Api().V1().Forms().ByFormId("FORM_ID").Delete(ctx, nil)
}
```

## Custom Base URL

```go
client, err := formalingo.NewClient("af_live_YOUR_KEY", "http://localhost:3000")
```

## Bulk Create Recipients Safely

```go
recipient := sdkapi.NewV1FormsItemRecipientsBulkPostRequestBody_recipients()
label, email := "Alice", "alice@example.com"
recipient.SetLabel(&label)
recipient.SetEmail(&email)
body := sdkapi.NewV1FormsItemRecipientsBulkPostRequestBody()
confirmBulk := true
body.SetConfirmBulk(&confirmBulk)
body.SetRecipients(
    []sdkapi.V1FormsItemRecipientsBulkPostRequestBody_recipientsable{recipient},
)

recipients, err := formalingo.CreateBulkRecipients(
    ctx,
    client,
    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
    body,
    "recipient-bulk-create-7f3f",
)
```

The required caller-owned key makes ambiguous retries safe. Reuse it only with the exact same serialized request body.
On `idempotency_request_in_progress`, retry the exact body with the same key. A different body returns `idempotency_key_conflict`; recipient erasure returns `idempotency_replay_unavailable`.

## Create a Document Submission

```go
signer := models.NewSignerInput()
role, name, email := "signer_1", "Alice", "alice@example.com"
signer.SetRole(&role)
signer.SetName(&name)
signer.SetEmail(&email)
body := models.NewCreateSubmissionBody()
body.SetSigners([]models.SignerInputable{signer})

submission, err := formalingo.CreateDocumentSubmission(
    ctx,
    client,
    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
    body,
    "document-create-7f3f",
)
fmt.Println(*submission.GetSigners()[0].GetLink())
```

## Publishing

Go modules require no registry account. Publishing is just git tags:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Consumers install with:

```bash
go get github.com/Formalingo/sdk-go@v0.1.0
```

The module auto-appears on [pkg.go.dev](https://pkg.go.dev) after first request.

## Documentation

- [Go SDK Guide](https://formalingo.com/docs/sdks/go)
- [API Reference](https://formalingo.com/docs/api-reference)

## License

MIT
