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
	"regexp"
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
	model := os.Getenv("GPT_MODEL")
	if model == "" {
		fmt.Println("Using GPT-3.5-turbo by default")
		model = openai.GPT3Dot5Turbo
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

		// Check if the file is not a test file
		if !info.IsDir() && filepath.Ext(path) == ".go" && !strings.HasSuffix(filepath.Base(path), "_test.go") {
			fmt.Printf("Generating test cases for %s\n", path)
			generateTestCases(ctx, client, path, model)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
	}
}

type testCase struct {
	Name    string
	Code    string
	Imports []string
}

func createTestFile(path string, packageName string, testCases []testCase) error {
	testFileName := filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), ".go")+"_test.go")

	// Read the existing test file, if it exists
	existingContent, err := ioutil.ReadFile(testFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var testFileContent strings.Builder

	// Write the existing content to the testFileContent
	if len(existingContent) > 0 {
		testFileContent.Write(existingContent)
	} else {
		// If the test file didn't exist, write the package and import statements
		testFileContent.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		testFileContent.WriteString("import (\n\t\"testing\"\n")
	}

	// Collect all unique import statements
	importSet := make(map[string]struct{})
	for _, testCase := range testCases {
		if testCase.Imports != nil {
			for _, imp := range testCase.Imports {
				importSet[imp] = struct{}{}
			}
		}
	}

	// Write unique import statements
	for imp := range importSet {
		testFileContent.WriteString(fmt.Sprintf("\t%s\n", imp))
	}

	if len(existingContent) == 0 {
		testFileContent.WriteString(")\n\n")
	}

	for _, testCase := range testCases {
		testFileContent.WriteString(fmt.Sprintf("// Test case for function %s\n", testCase.Name))
		testFileContent.WriteString(testCase.Code)
		testFileContent.WriteString("\n\n")
	}

	err = ioutil.WriteFile(testFileName, []byte(testFileContent.String()), 0644)
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

	var testCasesList []testCase

	testFileName := filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), ".go")+"_test.go")
	existingContentBytes, err := ioutil.ReadFile(testFileName)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	existingContent := string(existingContentBytes)

	for _, fn := range funcs {
		// Check if the test case for the function already exists in the test file
		testCaseExists := strings.Contains(existingContent, fmt.Sprintf("func Test%s(t *testing.T)", strings.Title(fn.Name)))
		if testCaseExists {
			fmt.Printf("Skipping existing test case for function %s\n", fn.Name)
			continue
		}

		testCaseCode, importContent, err := chatGPTTestCases(ctx, client, packageName, fn.Name, fn.Code, model)
		if err != nil {
			fmt.Printf("Error generating test case for function %s: %v\n", fn.Name, err)
			continue
		}

		fmt.Printf("Generated test case for function %s:\n\n%s\n\nTest case code:\n%s\n\nTest case import content:\n%s\n", fn.Name, fn.Code, testCaseCode, importContent)
		testCasesList = append(testCasesList, testCase{Name: fn.Name, Code: testCaseCode, Imports: importContent})
	}

	err = createTestFile(path, packageName, testCasesList)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
	}
}

func chatGPTTestCases(ctx context.Context, client *openai.Client, packageName, functionName, functionCode string, model string) (string, []string, error) {
	message := fmt.Sprintf("Your task is to generate a runnable test case code for the provided code. Please ensure that the test case covers all possible scenarios and edge cases, and that the code is easy to read and understand. Your response should only include the runnable code. Do not return any original code. Additionally, please make sure that the test case is well-organized and follows best practices for testing.\n\nfunction named: %s\npackage name: %s\nfunction code:\n```go\n%s\n```\n", functionName, packageName, functionCode)
	inputMessages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: message},
	}

	chatCompletions, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		// Model: openai.GPT3Dot5Turbo,
		Model:    model,
		Messages: inputMessages,
	})

	if err != nil {
		return "", nil, err
	}

	response := chatCompletions.Choices[0].Message.Content

	sanitizedCode, importContent, err := sanitizeCode(response)
	if err != nil {
		return "", nil, err
	}

	return sanitizedCode, importContent, nil
}

func sanitizeCode(rawCode string) (string, []string, error) {
	codePattern := "```(?:go)?\n((?s).*?)\n```"
	codeRegex := regexp.MustCompile(codePattern)

	codeMatches := codeRegex.FindStringSubmatch(rawCode)

	code := ""
	if len(codeMatches) < 2 {
		code = rawCode
	} else {
		code = codeMatches[1]
	}

	// Extract package and import statements

	packagePattern := `^package\s+[a-zA-Z_][a-zA-Z0-9_]*\s*\n`
	// importPattern := "import\\s*\\(([^)]+)\\)\n"
	importPattern := `(?m)^import(?:\s+\w+)?\s*(?:(?:\(\n(?:(\s*(?:\w+ )?\"[^"]+\"\s*\n))+\s*\))|(?:\"[^"]+\"))`

	packageRegex := regexp.MustCompile(packagePattern)
	code = packageRegex.ReplaceAllString(code, "")

	importRegex := regexp.MustCompile(importPattern)
	impotMatches := importRegex.FindStringSubmatch(code)
	var importContent []string
	if len(impotMatches) >= 2 {
		importBlock := impotMatches[0]
		code = strings.TrimSpace(strings.Replace(code, importBlock, "", 1))
		importContent = extractImports(importBlock)
	}
	return strings.TrimSpace(code), importContent, nil
}

func extractImports(importBlock string) []string {
	imports := []string{}

	// empty import block
	if importBlock == "" {
		return imports
	}

	// count number of new lines
	// if 1, then single line import
	// if > 1, then multi-line import
	lineNum := strings.Count(importBlock, "\n") + 1

	// single line import, e.g. `import "fmt"`, `import openai "github.com/sashabaranov/go-openai"`
	if lineNum == 1 {
		imp := strings.TrimSpace(strings.TrimPrefix(importBlock, "import"))
		imports = append(imports, imp)
		return imports
	}

	// multi-line import
	for _, line := range strings.Split(importBlock, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "import") {
			continue
		}
		if line == ")" {
			continue
		}

		imports = append(imports, line)
	}

	return imports
}
