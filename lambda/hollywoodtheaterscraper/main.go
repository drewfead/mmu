package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/drewfead/mmu/internal/commands"
)

type MyEvent struct {
	Arg string `json:"Arg"`
}

type MyResponse struct {
	Output string `json:"Output:"`
}

func lambda_handler(ctx context.Context, event MyEvent) (MyResponse, error) {
	app := &cli.App{
		Name:     "mmu",
		Usage:    "A utility for scraping websites for data about upcoming theatrical showings and home-video availability",
		Commands: commands.Scrapers,
	}

	err := app.RunContext(ctx, []string{"mmu", event.Arg})
	if err != nil {
		return MyResponse{Output: "error"}, fmt.Errorf("failed to execute app: %v", err)
	}

	return MyResponse{Output: "success"}, nil
}

func main() {
	lambda.Start(lambda_handler)
}
