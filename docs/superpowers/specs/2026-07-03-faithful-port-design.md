# protoc-gen-custom-errors 忠实迁移设计

## 目标

将用户指定源目录的有效项目代码忠实迁移到当前仓库，保留现有错误代码生成能力，同时彻底移除旧品牌名称，并统一使用当前项目名 `protoc-gen-custom-errors`。

## 迁移范围

- 迁移 Go 生成器源码、代码模板、protobuf 定义及生成文件、测试、Go 模块文件和 Buf 配置。
- 不迁移源仓库的 `.git`、`.idea` 等版本控制或编辑器私有文件。
- 当前仓库已有 README 将改写为完整的安装与使用说明。
- 保持源项目的生成行为、扩展字段和 Kratos errors 兼容逻辑，不进行无关重构。

## 命名规则

- Go 模块路径：`github.com/nice20121222/protoc-gen-custom-errors`
- 插件及二进制名：`protoc-gen-custom-errors`
- protoc 参数：`--custom-errors_out`
- 版本输出及生成文件注释：`protoc-gen-custom-errors`
- README 安装命令：`go install github.com/nice20121222/protoc-gen-custom-errors@latest`
- 所有源码、配置和文档中不得残留旧品牌名称。

## 实现方式

以源项目当前工作树为基准逐文件迁移，并仅做以下适配：替换模块导入路径及产品命名、清理私有文件、修正文档中的旧调用方式和明显笔误。生成器继续读取 `errors/errors.proto` 定义的枚举扩展，经嵌入模板生成 Kratos errors 辅助函数。

## 错误处理与兼容性

保留源项目对 HTTP 状态码范围、默认消息、数字业务码及未配置枚举值的处理方式。模块依赖版本原则上与源项目一致，以降低行为偏差。

## 验证

- 运行 `gofmt` 检查迁移后的 Go 文件。
- 运行 `go test ./...` 验证现有测试及包编译。
- 运行 `go build ./...` 验证插件可构建。
- 全仓搜索旧品牌名称，结果必须为空。
- 检查 README、版本输出、生成注释和模块路径均使用 `protoc-gen-custom-errors`。

## 完成标准

当前仓库包含源项目全部有效功能和测试，插件能成功构建并通过测试，品牌清理检查无残留，且用户可按照 README 安装和调用 `protoc-gen-custom-errors`。
