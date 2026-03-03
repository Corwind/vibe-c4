package main

import (
	"fmt"

	"github.com/example/sample-project/internal/handler"
	"github.com/example/sample-project/internal/service"
)

func main() {
	svc := service.New()
	h := handler.New(svc)
	fmt.Println(h)
}
