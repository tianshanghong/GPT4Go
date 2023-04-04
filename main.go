package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	fileLines := strings.Split(string(fileContent), "\n")

	var funcs []struct {
		Name string
		Code string
	}
	for _, decl := range parsedFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Recv == nil {
				funcName := funcDecl.Name.Name
				startLine := fset.Position(funcDecl.Pos()).Line - 1
				endLine := fset.Position(funcDecl.End()).Line - 1
				funcCode := strings.Join(fileLines[startLine:endLine+1], "\n")
				funcs = append(funcs, struct {
					Name string
					Code string
				}{Name: funcName, Code: funcCode})
			}
		}
	}

	for _, fn := range funcs {
		testCases, err := chatGPTTestCases(ctx, client, packageName, fn.Name, fn.Code)
		if err != nil {
			fmt.Printf("Error generating test cases: %v\n", err)
			continue
		}

		// fmt.Printf("Generated test cases for function %s:\n\n%s\n\nTest cases:\n%s\n", fn.Name, fn.Code, testCases)
		fmt.Printf("Generated test cases for function %s:\n\n%s\n", fn.Name, testCases)
	}
}

func chatGPTTestCases(ctx context.Context, client *openai.Client, packageName, functionName, functionCode string) (string, error) {
	message := fmt.Sprintf("Your task is to generate a runnable test case code for the provided code. Please ensure that the test case code is concise and effective in testing the functionality of the given code. Your response should only include the runnable test case code, without any unnecessary or redundant information. Additionally, please make sure that the test case code is properly formatted and follows best practices for testing.\nPlease generate unit test code for the following Golang function named %s in package %s:\n\n```go\n%s\n```\n", functionName, packageName, functionCode)
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
