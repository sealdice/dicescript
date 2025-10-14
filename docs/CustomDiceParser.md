# 自定义骰点解析器使用指南（流式解析）

本文介绍如何使用新版 API 在 DiceScript 中注册一个自定义骰点解析器，允许在解析阶段逐字符地消费或回退输入，构建属于自己的迷你 parser。

## 核心接口

```go
// 注册自定义解析器 + 执行器。
func (ctx *Context) RegCustomDiceParser(
    parser CustomDiceParserFunc,
    handler CustomDiceHandler,
) error

// 解析器签名
func CustomDiceParserFunc(
    ctx *Context,
    stream *CustomDiceStream,
) (*CustomDiceParseResult, error)

// 解析结果
type CustomDiceParseResult struct {
    Groups  []string
    Display string
    Payload any
    Matched bool
}

// 执行器签名
func CustomDiceHandler(
    ctx *Context,
    groups []string,
    payload any,
) (*VMValue, string, error)
```

- `stream` 暴露逐字符 API（详见下文），负责**尝试**识别你想支持的骰点语法。
- 返回 `nil` 或 `Matched=false` 表示未匹配，解析器会继续尝试下一条自定义规则或内置语法。
- `Payload` 可以保存解析阶段构造的任意结构，稍后会原样传给 `handler`，方便避免二次解析。

## `CustomDiceStream` 能做什么？

下表展示核心方法（具体以实现为准）：

| 方法 | 作用 |
| --- | --- |
| `Peek() (rune, bool)` | 查看下一个字符但不前进 |
| `Read() (rune, bool)` | 读取下一个字符并前进 |
| `Unread()` | 将最近一次 `Read` 的字符放回 |
| `ResetAttempt()` | 放弃本次尝试，光标回到起始位置 |
| `Commit()` | （可选）标记本次解析已确认成功 |
| `Consumed()` | 返回已消费的字节数 |
| `Current() string` | 返回当前已消费的原始文本 |
| `ReadDigits()` | 便捷函数：连续读取数字字符 |

可根据需要扩展更多便捷方法（如读取标识符、跳过空白等）。

## 示例：语法 `C{基数}T{阈值}`

```go
type dicePayload struct {
    BaseStr      string
    ThresholdStr string
}

ctx.RegCustomDiceParser(
    func(ctx *dicescript.Context, stream *dicescript.CustomDiceStream) (*dicescript.CustomDiceParseResult, error) {
        r, ok := stream.Read()
        if !ok || r != 'C' {
            stream.ResetAttempt()
            return &dicescript.CustomDiceParseResult{Matched: false}, nil
        }

        baseStr, ok := stream.ReadDigits()
        if !ok {
            stream.ResetAttempt()
            return &dicescript.CustomDiceParseResult{Matched: false}, nil
        }

        r, ok = stream.Read()
        if !ok || (r != 'T' && r != 't') {
            stream.ResetAttempt()
            return &dicescript.CustomDiceParseResult{Matched: false}, nil
        }

        thresholdStr, ok := stream.ReadDigits()
        if !ok {
            stream.ResetAttempt()
            return &dicescript.CustomDiceParseResult{Matched: false}, nil
        }

        stream.Commit()
        payload := &dicePayload{BaseStr: baseStr, ThresholdStr: thresholdStr}
        groups := []string{stream.Current(), baseStr, thresholdStr}
        return &dicescript.CustomDiceParseResult{
            Groups:  groups,
            Payload: payload,
            Matched: true,
        }, nil
    },
    func(ctx *dicescript.Context, groups []string, payload any) (*dicescript.VMValue, string, error) {
        info := payload.(*dicePayload)
        base, err := strconv.ParseInt(info.BaseStr, 10, 64)
        if err != nil {
            return nil, "", err
        }
        threshold, err := strconv.ParseInt(info.ThresholdStr, 10, 64)
        if err != nil {
            return nil, "", err
        }
        result := base + threshold
        detail := fmt.Sprintf("C%sT%s=%d", groups[1], groups[2], result)
        return dicescript.NewIntVal(dicescript.IntType(result)), detail, nil
    },
)
```

要点：
- `ReadDigits()` 仅作示例，实际可通过循环 `Read` + `unicode.IsDigit` 实现。
- `groups[0]` 建议放完整匹配文本，其余元素可按需求自定义。
- 解析失败时记得 `ResetAttempt()`，否则后续操作可能接在错误位置。

## 错误处理与回退

- 若 `parser` 返回 `nil` 或 `Matched=false` 将继续交由其他规则匹配，不会报错。
- `err != nil` 会立即终止本次求值并抛出错误。
- 解析成功后务必使 `Consumed()` 返回正数（默认即为已读取字节数），否则框架会忽略本次匹配。

## 调试建议

1. 打开 `vm.Config.PrintBytecode = true`，确认生成的指令中出现 `dice.custom`。
2. 调用 `vm.GetDetailText()`，检查输出是否包含 `handler` 返回的描述文本。
3. 为自定义解析器编写单元测试，可参考 `rollvm_test.go` 中的 `TestCustomDiceParserStream`。

## 常见问题

| 问题 | 解决方案 |
| --- | --- |
| 读取过多字符 | 使用 `Unread()` 或 `ResetAttempt()` 回退，并返回 `Matched=false` |
| 需要携带复杂上下文 | 利用 `Payload` 传递结构体，无需字符串再解析 |
| 想尝试多种语法 | 注册多条 `RegCustomDiceParser`，按顺序逐一尝试 |

借助 `CustomDiceStream`，你可以轻松将 rule-based 或手写解析逻辑嵌入 DiceScript，构建高度定制化的骰点算符。
