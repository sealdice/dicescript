 # 自定义骰点算符用法指南(正则模式)

  ## 基本概念

  - RegCustomDice(pattern, handler)：注册一个自定义骰点。pattern 是 整段表达式起始处要匹配的正则；handler 用来兑现
  （roll）结果。
  - 回调签名：func(ctx *Context, groups []string) (*VMValue, string, error)
      - groups[0] 是整段匹配文本，其余元素对应正则的捕获组。
      - 返回值：
          - *VMValue：最终压栈的结果（会被 clone，因此可以复用零状态实例）。
          - string：展示在计算过程中的细节文本；留空则使用原始匹配文本。
          - error：非 nil 即终止执行并将错误抛出给调用者。

  ## 标准示例

  vm := dicescript.NewVM()

  err := vm.RegCustomDice(`E(\d+)`, func(ctx *dicescript.Context, groups []string) (*dicescript.VMValue, string, error)
  {
      if len(groups) < 2 {
          return nil, "", fmt.Errorf("缺少数值参数")
      }
      v, err := strconv.ParseInt(groups[1], 10, 64)
      if err != nil {
          return nil, "", err
      }
      // 返回结果值、展示文本、错误
      return dicescript.NewIntVal(dicescript.IntType(v)*2), "custom:" + groups[0], nil
  })
  if err != nil {
      log.Fatal(err)
  }

  if err := vm.Run("E5 + 1"); err != nil {
      log.Fatal(err)
  }
  fmt.Println(vm.Ret.ToString())       // => 11
  fmt.Println(vm.GetDetailText())      // 包含 custom:E5 的细节

  要点：

  - 正则必须匹配表达式的起始位置；一旦匹配成功，解析器会自动“吃掉”对应长度的源码字节，并在当前 detail 区域写入 handler
  的 string 返回值。
  - 若希望展示备用文字，可在 handler 中自行拼接（如上例的 custom:E5）。
  - 错误处理由 handler 返回的 error 控制，一旦非 nil 将停止本次求值。

  ## 测试与细节

  - 项目包含 rollvm_test.go 中的 TestCustomDice* 用例，可作参考。
  - 回调中的 ctx 是当前子 VM，支持读取全局/局部变量或继续执行其他表达式（注意 reentrancy）。
  - 返回值会被 clone，因此即使你复用了同一个 *VMValue 也不会污染内部状态。
