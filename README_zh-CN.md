
<div align="center">

# 🤖 GPT4Go 🚀

GPT4Go：AI驱动的Golang测试用例生成 🧪

[English](README.md) | 中文

</div>

GPT4Go 是一个使用 OpenAI 的 ChatGPT 🌐 自动为 Golang 生成测试用例文件的开源项目。该项目帮助开发者快速为他们的函数创建测试用例，确保代码得到高效和全面的测试 🧪。

## 🌟 特点

-   🎯 使用 OpenAI 的 ChatGPT 为 Golang 函数自动生成测试用例。
-   🚫 跳过已经测试过的函数，不生成测试用例。
-   💡 为大型函数提供拆分建议，以实现更好的软件工程实践。
-   📚 生成组织良好、易于阅读的测试用例代码。

## 🛠 安装

1.  确保您的系统上已安装 Golang 💻。如果没有，请按照[官方安装指南](https://golang.google.cn/doc/install)操作。
    
2.  克隆代码仓库 📦：

```bash
git clone https://github.com/yourusername/GPT4Go.git
```

3.  切换到项目目录 📂:

```bash
cd GPT4Go
```

4.  构建项目 🔨:

```bash
go build .
```

5.  设置必需的环境变量 🌍:

```bash
export OPENAI_API_KEY=your_openai_api_key
export GPT_MODEL=model_name  # (可选，默认为 gpt-3.5-turbo)
```

在 [OpenAI 网站](https://beta.openai.com/docs/developer-quickstart/api-key) 上，您可以获取 OpenAI 的 API 密钥 🔑。您还可以指定要用于生成测试用例的模型。默认模型是 `gpt-3.5-turbo`，这是最快的模型 🏎，而 `gpt-4` 是最准确的模型 🎯。您可以在[这里](https://platform.openai.com/docs/models/overview)找到所有可用模型的列表。

## 📚 使用方法

要为特定目录或文件生成测试用例，请运行以下命令：

```bash
./GPT4Go path/to/your/target/directory/or/file
```

该命令将遍历指定的目录或文件，并为所有没有相应测试用例的 `_test.go` 文件中的函数生成测试用例。

请注意，您需要 OpenAI 的 API 密钥才能使用 ChatGPT 功能 🔐。您可以从 [OpenAI 的网站](https://www.openai.com/)获取一个。

## 👥 贡献

欢迎贡献！请随时提交拉取请求 📥，报告错误 🐞 或通过 [GitHub 问题](https://github.com/yourusername/GPT4Go/issues) 页面提出新功能建议 💡。

## 📄 许可证

GPT4Go 根据 [MIT 许可证](https://chat.openai.com/LICENSE) 授权 📃。

## 🙏 鸣谢

本项目使用 OpenAI 的 ChatGPT 构建，这是一个用于生成类人回应的强大语言模型 🧠。您可以在 OpenAI 网站 上了解更多关于 ChatGPT 和 GPT-4 架构的信息。
