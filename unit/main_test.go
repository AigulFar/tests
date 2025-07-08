package unit

import (
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func main() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:      "hello-world",
		WaitingFor: wait.ForLog("Hello from Docker!"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer container.Terminate(ctx)

	fmt.Println("Testcontainers works!")
}
