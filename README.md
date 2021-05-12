## Install 

```bash
go get -u github.com/gota33/initializr
```

## Example 

```go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gota33/initializr"
	fiber2 "github.com/gota33/initializr/fiber/v2"
)

func main() {
	ctx, cancel := initializr.GracefulContext()
	defer cancel()

	server, listen, shutdown := fiber2.New(":8080")

	server.Get("hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello")
	})

	if err := initializr.Run(ctx, listen, shutdown); err != nil {
		log.Fatalf("Exit with error: %v", err)
	}
}
```