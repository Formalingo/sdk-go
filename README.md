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
