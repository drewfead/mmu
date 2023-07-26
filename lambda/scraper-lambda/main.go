package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/drewfead/mmu/internal/commands"
)

func lambdaHandler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	app := &cli.App{
		Name:     "mmu",
		Usage:    "A utility for scraping websites for data about upcoming theatrical showings and home-video availability",
		Commands: commands.Scrapers,
	}

	err := app.RunContext(ctx, []string{"mmu", request.Body})
	if err != nil {
		return events.LambdaFunctionURLResponse{Body: "error"}, fmt.Errorf("failed to execute app: %v", err)
	}

	return events.LambdaFunctionURLResponse{Body: "success", StatusCode: 200}, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
