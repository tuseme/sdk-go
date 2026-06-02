# Tuseme Go SDK

Official Go client for the [Tuseme SMS API](https://docs.tuseme.co.ke).

[![Go Reference](https://pkg.go.dev/badge/github.com/tuseme/sdk-go.svg)](https://pkg.go.dev/github.com/tuseme/sdk-go)
[![Go 1.22+](https://img.shields.io/badge/go-1.22+-00ADD8.svg)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Installation

```bash
go get github.com/tuseme/sdk-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    tuseme "github.com/tuseme/sdk-go"
)

func main() {
    client := tuseme.NewClient("tk_test_your_api_key", "sk_test_your_api_secret")

    resp, err := client.Messages.Send(&tuseme.SendRequest{
        Content:  "Hello from Tuseme! Your OTP is 482910.",
        SenderID: "TUSEME-LTD",
        Recipients: []tuseme.Recipient{
            {MSISDN: "+254712345678", Name: "John Doe"},
        },
        Type:     "transactional",
        Priority: "HIGH",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Message ID: %s\nStatus: %s\n", resp.MessageID, resp.Status)
}
```

## Features

- **Zero dependencies** — uses only the Go standard library
- **Thread-safe** — `sync.Mutex`-based token management
- **Automatic authentication** — tokens obtained and refreshed transparently
- **Built-in retries** — exponential backoff for transient failures
- **Functional options** — `WithBaseURL()`, `WithTimeout()`, `WithRetries()`

## Authentication

```go
// Sandbox credentials (for testing)
client := tuseme.NewClient("tk_test_...", "sk_test_...")

// Production credentials
client := tuseme.NewClient("tk_live_...", "sk_live_...")
```

The SDK will:
1. Automatically obtain an access token on the first request
2. Cache the token until it expires
3. Transparently refresh expired tokens

## Usage

### Send SMS

```go
// Single recipient
resp, err := client.Messages.Send(&tuseme.SendRequest{
    Content:    "Your verification code is 123456",
    SenderID:   "TUSEME-LTD",
    Recipients: []tuseme.Recipient{{MSISDN: "+254712345678"}},
    Type:       "transactional",
})

// Multiple recipients with metadata
resp, err := client.Messages.Send(&tuseme.SendRequest{
    Content:  "Flash sale! 50% off today only.",
    SenderID: "TUSEME-LTD",
    Recipients: []tuseme.Recipient{
        {MSISDN: "+254712345678", Name: "Alice"},
        {MSISDN: "+254798765432", Name: "Bob"},
    },
    Type:     "promotional",
    Metadata: map[string]string{"campaign": "flash_sale_q2"},
})
```

### Check Delivery Status

```go
status, err := client.Messages.Get("msg_a1b2c3d4...")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\nDelivered at: %s\n", status.Status, status.DeliveredAt)
```

### List Messages

```go
data, err := client.Messages.List(1, 20)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

## Error Handling

```go
resp, err := client.Messages.Send(req)
if err != nil {
    var tusemeErr *tuseme.TusemeError
    if errors.As(err, &tusemeErr) {
        fmt.Printf("API error (status %d): %s\n", tusemeErr.StatusCode, tusemeErr.Message)
    } else {
        fmt.Printf("Network error: %v\n", err)
    }
    return
}
```

## Configuration

```go
client := tuseme.NewClient(
    "tk_test_...",
    "sk_test_...",
    tuseme.WithBaseURL("https://api.tuseme.co.ke/api/v1"),
    tuseme.WithTimeout(30 * time.Second),
    tuseme.WithRetries(3),
)
```

## License

MIT — see [LICENSE](LICENSE).
