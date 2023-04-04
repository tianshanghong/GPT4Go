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
			generateTestCases(ctx, client, path, openai.GPT4)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}
}

func createTestFile(path string, packageName string, testCases []struct {
	Name string
	Code string
}) error {
	testFileName := filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), ".go")+"_test.go")
	var testFileContent strings.Builder

	testFileContent.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	testFileContent.WriteString("import (\n\t\"testing\"\n)\n\n")

	for _, testCase := range testCases {
		testFileContent.WriteString(fmt.Sprintf("// Test case for function %s\n", testCase.Name))
		testFileContent.WriteString(testCase.Code)
		testFileContent.WriteString("\n\n")
	}

	err := ioutil.WriteFile(testFileName, []byte(testFileContent.String()), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Test file generated: %s\n", testFileName)
	return nil
}

func generateTestCases(ctx context.Context, client *openai.Client, path string, model string) {
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
				if funcName == "main" {
					fmt.Printf("Skipping main function")
					continue // Skip the main function
				}
				startLine := fset.Position(funcDecl.Pos()).Line - 1
				endLine := fset.Position(funcDecl.End()).Line - 1
				funcCode := strings.Join(fileLines[startLine:endLine+1], "\n")

				// Check if the function has more than 50 lines
				if endLine-startLine+1 > 100 {
					fmt.Printf("Function %s is longer than 50 lines. It is recommended to make it shorter for better software engineering practices.\n", funcName)
					continue // Skip generating test cases for this function
				}

				funcs = append(funcs, struct {
					Name string
					Code string
				}{Name: funcName, Code: funcCode})
			}
		}
	}

	var testCasesList []struct {
		Name string
		Code string
	}

	for _, fn := range funcs {
		testCaseCode, err := chatGPTTestCases(ctx, client, packageName, fn.Name, fn.Code, model)
		if err != nil {
			fmt.Printf("Error generating test case for function %s: %v\n", fn.Name, err)
			continue
		}

		fmt.Printf("Generated test case for function %s:\n\n%s\n\nTest case code:\n%s\n", fn.Name, fn.Code, testCaseCode)
		testCasesList = append(testCasesList, struct {
			Name string
			Code string
		}{Name: fn.Name, Code: testCaseCode})
	}

	err = createTestFile(path, packageName, testCasesList)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
	}

}

func chatGPTTestCases(ctx context.Context, client *openai.Client, packageName, functionName, functionCode string, model string) (string, error) {
	message := fmt.Sprintf("Your task is to generate a runnable test case code for the provided code. Please ensure that the test case code is concise and effective in testing the functionality of the given code. Your response should only include the runnable test case code, without any unnecessary or redundant information. Additionally, please make sure that the test case code is properly formatted and follows best practices for testing.\nPlease generate unit test code for the following Golang function named %s in package %s:\n\n```go\n%s\n```\n", functionName, packageName, functionCode)
	inputMessages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: message},
	}

	chatCompletions, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		// Model: openai.GPT3Dot5Turbo,
		Model:    model,
		Messages: inputMessages,
	})

	if err != nil {
		return "", err
	}

	return chatCompletions.Choices[0].Message.Content, nil
}
