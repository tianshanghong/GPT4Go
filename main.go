package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/tools/go/packages"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable")
		os.Exit(1)
	}

	ctx := context.Background()
	client := openai.NewClient(apiKey)

	targetPath := "."
	if len(os.Args) > 1 {
		targetPath = os.Args[1]
	}

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			fmt.Printf("Generating test cases for %s\n", path)
			generateTestCases(ctx, client, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}
}

func generateTestCases(ctx context.Context, client *openai.Client, path string) {
	cfg := &packages.Config{
		Mode: packages.NeedName,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		fmt.Printf("Error loading package: %v\n", err)
		return
	}

	packageName := pkgs[0].Name

	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return
	}

	var funcs []string
	for _, decl := range parsedFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Recv == nil {
				funcName := funcDecl.Name.Name
				funcs = append(funcs, funcName)
			}
		}
	}

	for _, fn := range funcs {
		testCases, err := chatGPTTestCases(ctx, client, packageName, fn)
		if err != nil {
			fmt.Printf("Error generating test cases: %v\n", err)
			continue
		}

		fmt.Printf("Generated test cases for function %s:\n\n%s\n", fn, testCases)
	}
}

func chatGPTTestCases(ctx context.Context, client *openai.Client, packageName, functionName string) (string, error) {
	message := fmt.Sprintf("Please generate unit test code for a Golang function named %s in package %s. The returned value should be functional golang code.", functionName, packageName)
	inputMessages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: message},
	}

	chatCompletions, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: inputMessages,
	})

	if err != nil {
		return "", err
	}

	return chatCompletions.Choices[0].Message.Content, nil
}
