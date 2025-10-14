# DiceScript

[![Go Report Card](https://goreportcard.com/badge/github.com/sealdice/dicescript)](https://goreportcard.com/report/github.com/sealdice/dicescript)
![Software License](https://img.shields.io/badge/license-Apache2-brightgreen.svg?style=flat-square)
[![GoDoc](https://godoc.org/github.com/sealdice/dicescript?status.svg)](https://godoc.org/github.com/sealdice/dicescript)

通用TRPG骰点脚本语言。

Simple script language for TRPG dice engine.

特性:
- 支持整数、浮点数、字符串、数组、字典、函数，常见算符以及if和while逻辑语句
- 支持形如d20等的trpg用骰点语法
- 全类型可序列化(包括函数在内)
- 对模板字符串语法做大量优化，可胜任模板引擎
- 易于使用，方便扩展
- 稳定可靠，极高的测试覆盖率
- 免费并可商用
- 可编译到JavaScript

测试页面:

https://sealdice.github.io/dicescript/

这个项目是海豹核心的骰点解释器的完全重构。
从第一次实现中吸取了很多经验和教训，并尝试做得更好。

## 如何使用

[DiceScript指南](./docs/GUIDE.md)

你可以从这里了解如何使用DiceScript进行骰点，编写自己的TRPG规则，以及如何嵌入到任何你想要的地方。

进阶·不修改代码的情况下进行语法扩展:

[自定义骰点语法使用指南(流式解析)](./docs/CustomDiceParser.md)
[自定义骰点算符用法指南(正则模式)](./docs/CustomDiceRegex.md)


## 设计原则

* 从主流语言和跑团软件中借鉴语法，如Golang/Python/JS/Fvtt/BCDice，不随意发明
* 符合国内跑团指令的一般习惯
* 要具有较强的配置和扩展能力，符合trpg场景的需求
* 一定限度内容忍全角符号
* 兼容gopherjs
* 良好的错误提示文本
* 支持多线程


## 进度

- [x] 基础类型 int float string null
- [x] 一元算符 + -
- [x] 二元算符 +-*/% >,>=,==,!=,<,<=,&,|,&&,||
- [x] 三元算符 ? :
- [x] 空值合并算符 ??
- [x] 骰点运算 - 流行语法: d20, 3d20, (4+5)d(20), 2d20k1, 2d20q1
- [x] 骰点运算 - fvtt语法: 2d20kl, 2d20kh, 2d20dl, 2d20dh, d20min10, d20max10
- [x] 骰点运算 - CoC / Fate / WoD / Double Cross
- [x] 骰点运算 - 自定义算符
- [x] 高级类型 数组array
- [x] 高级类型 字典dict
- [x] 高级类型 计算数值computed
- [x] 逻辑语法 if ... else ..
- [x] 逻辑语法 while
- [x] 函数支持
- [x] 内置函数
- [x] 分片语法
- [x] 区间数组
- [x] 变量支持
- [x] 序列化和反序列化
- [x] 计算过程显示
- [x] 角色属性对接
- [x] 报错信息优化
- [x] 线程安全
- [x] 变量作用域
- [ ] 测试覆盖率 86% / 90%

## 更新记录

[更新记录](./docs/CHANGELOG.md)

## TODO

* computed 的repr格式无法读入

## 开发

如果修改了文法，使用这个工具重新生成:
```
go install github.com/fy0/pigeon@latest
pigeon -nolint -optimize-parser -optimize-ref-expr-by-index -o .\roll.peg.go .\roll.peg
```
