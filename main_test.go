package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// Test case for function createTestFile
func TestCreateTestFile(t *testing.T) {
	tempDir := t.TempDir()
	testCases := []testCase{
		{
			Name: "TestAdd",
			Code: "func TestAdd(t *testing.T) { ... }",
		},
		{
			Name: "TestMultiply",
			Code: "func TestMultiply(t *testing.T) { ... }",
		},
	}

	err := createTestFile(tempDir+"/sample.go", "main", testCases)
	if err != nil {
		t.Fatalf("createTestFile() error: %v", err)
	}

	generatedFile := tempDir + "/sample_test.go"
	if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
		t.Fatalf("Test file not created: %v", err)
	}

	content, err := ioutil.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Reading generated file error: %v", err)
	}

	expectedContent := `package main

import (
	"testing"
)

// Test case for function TestAdd
func TestAdd(t *testing.T) { ... }

// Test case for function TestMultiply
func TestMultiply(t *testing.T) { ... }

`
	if string(content) != expectedContent {
		t.Fatalf("Generated file content doesn't match expected content")
	}
}

// Test case for function generateTestCases

func TestGenerateTestCases(t *testing.T) {
	ctx := context.Background()
	// Initialize the OpenAI client with the API key.
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test due to missing API key.")
	}
	client := openai.NewClient(apiKey)

	// Create the testdata directory if it doesn't exist
	testdataDir := filepath.Join(".", "__testdata")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		err = os.MkdirAll(testdataDir, os.ModePerm)
		if err != nil {
			t.Fatalf("Error creating testdata directory: %v\n", err)
		}
	}
	path := filepath.Join(testdataDir, "sample.go")

	// Setup and cleanup
	err := ioutil.WriteFile(path, []byte(`
package testdata

func sum(a, b int) int {
	return a + b
}

func main() {
	a := 1
	b := 2
	c := sum(a, b)
}
`),
		os.ModePerm)
	if err != nil {
		t.Fatalf("Error writing testdata file: %v\n", err)
	}
	defer os.RemoveAll(testdataDir)

	// Run the function and capture the output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	generateTestCases(ctx, client, path, openai.GPT3Dot5Turbo)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = old

	// Check if the test case is generated for 'sum' function
	if !strings.Contains(string(out), "Generated test case for function sum") {
		t.Errorf("Expected to find test case for 'sum' function, but it was not generated")
	}
}

// Test case for function chatGPTTestCases
func TestChatGPTTestCases(t *testing.T) {
	packageName := "main"
	functionName := "exampleFunction"
	functionCode := "func exampleFunction() {\n\t// function code\n}\n"

	ctx := context.Background()
	// Initialize the OpenAI client with the API key.
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test due to missing API key.")
	}
	client := openai.NewClient(apiKey)

	_, _, err := chatGPTTestCases(ctx, client, packageName, functionName, functionCode, openai.GPT3Dot5Turbo)

	assert.NoError(t, err, "There should be no error")
}

// Test case for function sanitizeCode
func TestSanitizeCode(t *testing.T) {
	tests := []struct {
		name                  string
		rawCode               string
		expectedCode          string
		expectedImportContent []string
		expectedError         error
	}{
		{
			name:                  "No code block",
			rawCode:               "This is a sample text without a code block",
			expectedCode:          "This is a sample text without a code block",
			expectedImportContent: nil,
			expectedError:         nil,
		},
		{
			name:                  "Code block without go",
			rawCode:               "```\nfunc hello() {\n\tfmt.Println(\"Hello, test!\")\n}\n```",
			expectedCode:          "func hello() {\n\tfmt.Println(\"Hello, test!\")\n}",
			expectedImportContent: nil,
			expectedError:         nil,
		},
		{
			name:                  "Code block with go",
			rawCode:               "```go\nfunc hello() {\n\tfmt.Println(\"Hello, test!\")\n}\n```",
			expectedCode:          "func hello() {\n\tfmt.Println(\"Hello, test!\")\n}",
			expectedImportContent: nil,
			expectedError:         nil,
		},
		{
			name:                  "Code block with package and import",
			rawCode:               "```go\npackage main\n\nimport \"fmt\"\n\nfunc hello() {\n\tfmt.Println(\"Hello, test!\")\n}\n```",
			expectedCode:          "func hello() {\n\tfmt.Println(\"Hello, test!\")\n}",
			expectedImportContent: []string{"\"fmt\""},
			expectedError:         nil,
		},
		{
			name:                  "Multiple code blocks",
			rawCode:               "```go\nfunc hello() {\n\tfmt.Println(\"Hello, test!\")\n}\n```\n\nSome text\n\n```go\nfunc world() {\n\tfmt.Println(\"World, test!\")\n}\n```",
			expectedCode:          "func hello() {\n\tfmt.Println(\"Hello, test!\")\n}",
			expectedImportContent: nil,
			expectedError:         nil,
		},
		{
			name:                  "Multiple imports",
			rawCode:               "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tfmt.Println(\"Hello, world!\")\n}",
			expectedCode:          "func main() {\n\tfmt.Println(\"Hello, world!\")\n}",
			expectedImportContent: []string{`"fmt"`, `"os"`},
			expectedError:         nil,
		},
		{
			name:                  "Single-line import with alias",
			rawCode:               "package main\n\nimport f \"fmt\"\n\nfunc main() {\n\tf.Println(\"Hello, world!\")\n}",
			expectedCode:          "func main() {\n\tf.Println(\"Hello, world!\")\n}",
			expectedImportContent: []string{`f "fmt"`},
			expectedError:         nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, importContent, err := sanitizeCode(test.rawCode)
			if code != test.expectedCode {
				t.Errorf("Expected code: %s, got: %s", test.expectedCode, code)
			}
			equal := true
			for i, v := range importContent {
				if v != test.expectedImportContent[i] {
					equal = false
					break
				}
			}
			if equal != true {
				t.Errorf("Expected importContent: %s, got: %s", test.expectedImportContent, importContent)
			}
			if err != test.expectedError {
				t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
			}
		})
	}
}

// Test case for function extractImports
func TestExtractImports(t *testing.T) {
	tests := []struct {
		name        string
		importBlock string
		wantImports []string
	}{
		{
			name:        "Single line import",
			importBlock: `import "fmt"`,
			wantImports: []string{`"fmt"`},
		},
		{
			name:        "Single line named import",
			importBlock: `import openai "github.com/sashabaranov/go-openai"`,
			wantImports: []string{`openai "github.com/sashabaranov/go-openai"`},
		},
		{
			name: "Multi-line import",
			importBlock: `
		import (
			"fmt"
			"errors"
		)
		`,
			wantImports: []string{`"fmt"`, `"errors"`},
		},
		{
			name: "Multi-line import with extra spaces",
			importBlock: `
		import (
			"fmt"
			"errors"
		)
		`,
			wantImports: []string{`"fmt"`, `"errors"`},
		},
		{
			name: "Multi-line import with empty lines",
			importBlock: `
		import (

			"fmt"

			"errors"
		)
		`,
			wantImports: []string{`"fmt"`, `"errors"`},
		},
		{
			name:        "Empty import block",
			importBlock: "",
			wantImports: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotImports := extractImports(tt.importBlock); !equalSlices(gotImports, tt.wantImports) {
				t.Errorf("extractImports() = %v, want %v", gotImports, tt.wantImports)
			}
		})
	}
}

func equalSlices(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, val := range s1 {
		if val != s2[i] {
			return false
		}
	}

	return true
}

// Test case for function findImportBlock
func TestFindImportBlock(t *testing.T) {
	testCases := []struct {
		code     string
		expected string
	}{
		{
			code: `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
			expected: `import "fmt"`,
		},
		{
			code: `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, world!")
}`,
			expected: `import (
	"fmt"
	"os"
)`,
		},
		{
			code: `package main

func main() {
	fmt.Println("Hello, world!")
}`,
			expected: "",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test Case %d", i+1), func(t *testing.T) {
			result := findImportBlock(testCase.code)
			if result != testCase.expected {
				t.Errorf("Expected: %s, got: %s", testCase.expected, result)
			}
		})
	}
}
