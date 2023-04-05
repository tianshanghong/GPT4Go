# GPT4Go

GPT4Go is an open-source project that auto-generates test case files for Golang using OpenAI's ChatGPT. The project helps developers quickly create test cases for their functions, ensuring that their code is tested efficiently and comprehensively.

## Features

-   Auto-generates test cases for Golang functions using OpenAI's ChatGPT.
-   Skips generating test cases for already tested functions.
-   Provides suggestions for breaking down large functions for better software engineering practices.
-   Generates well-organized, easy-to-read test case code.

## Installation

1.  Ensure that you have Golang installed on your system. If not, please follow the [official installation guide](https://golang.org/doc/install).
    
2.  Clone the repository:
    

`git clone https://github.com/yourusername/GPT4Go.git`

3.  Change into the project directory:

`cd GPT4Go`

4.  Build the project:

`go build .`

5.  Set the required environment variables:

`export OPENAI_API_KEY=your_openai_api_key export GPT_MODEL=model_name (optional, defaults to gpt-3.5-turbo)`

From the [OpenAI website](https://beta.openai.com/docs/developer-quickstart/api-key), you can obtain an API key for OpenAI. You can also specify the model you want to use for generating test cases. The default model is `gpt-3.5-turbo`, which is the fastest model available, while `gpt-4` is the most accurate model. You can find a list of all available models [here](https://platform.openai.com/docs/models/overview).

## Usage

To generate test cases for a specific directory or file, run the following command:

`./GPT4Go path/to/your/target/directory/or/file`

This command will walk through the specified directory or file and generate test cases for all functions that do not have corresponding test cases in the `_test.go` file.

Please note that you will need an API key for OpenAI to use the ChatGPT functionality. You can obtain one from [OpenAI's website](https://www.openai.com/).

## Contributing

Contributions are welcome! Please feel free to submit pull requests, report bugs, or suggest new features through the [GitHub issues](https://github.com/yourusername/GPT4Go/issues) page.

## License

GPT4Go is licensed under the [MIT License](https://chat.openai.com/LICENSE).

## Acknowledgements

This project is built using OpenAI's ChatGPT, a powerful language model for generating human-like responses. You can learn more about ChatGPT and GPT-4 architecture on the [OpenAI website](https://www.openai.com/).