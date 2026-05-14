# spillway

A dead-letter queue manager for failed jobs across multiple queue backends with retry policy support.

---

## Installation

```bash
go get github.com/yourorg/spillway
```

## Usage

```go
package main

import (
    "github.com/yourorg/spillway"
)

func main() {
    manager := spillway.New(spillway.Config{
        Backend: spillway.Redis("localhost:6379"),
        RetryPolicy: spillway.RetryPolicy{
            MaxAttempts: 5,
            Backoff:     spillway.ExponentialBackoff(2),
        },
    })

    // Register a handler for failed jobs
    manager.Handle("email:send", func(job spillway.Job) error {
        // process the failed job
        return nil
    })

    // Replay all dead-lettered jobs for a queue
    if err := manager.Replay("email:send"); err != nil {
        panic(err)
    }

    manager.Start()
}
```

## Supported Backends

- Redis
- RabbitMQ
- SQS

## License

[MIT](LICENSE)