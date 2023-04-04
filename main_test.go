package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// Test case for function createTestFile

func TestCreateTestFile(t *testing.T) {
	tempDir := t.TempDir()
	testCases := []struct {
		Name string
		Code string
	}{
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
	testdataDir := filepath.Join(".", "testdata")
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
	defer os.Remove(path)

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

	_, err := chatGPTTestCases(ctx, client, packageName, functionName, functionCode, openai.GPT3Dot5Turbo)

	assert.NoError(t, err, "There should be no error")
}
