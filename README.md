# DiceScript

[![Go Report Card](https://goreportcard.com/badge/github.com/sealdice/dicescript)](https://goreportcard.com/report/github.com/sealdice/dicescript)
![Software License](https://img.shields.io/badge/license-Apache2-brightgreen.svg?style=flat-square)
[![GoDoc](https://godoc.org/github.com/sealdice/dicescript?status.svg)](https://godoc.org/github.com/sealdice/dicescript)

通用TRPG骰点脚本语言。

Simple script language for TRPG dice engine.

特性:
- 易于使用，方便扩展
- 稳定可靠，极高的测试覆盖率
- 免费，并可商用
- 支持JavaScript
- 海豹TRPG骰点核心的第二代解释器

测试页面:

https://sealdice.github.io/dicescript/

## 如何使用

[DiceScript指南](./GUIDE.md)

你可以从这里了解如何使用DiceScript进行骰点，编写自己的TRPG规则，以及如何嵌入到任何你想要的地方。


## 设计原则

先随便写一些想到的，然后再细化

这是未来SealDice项目的一部分，一次从零重构。

目前SealDice使用一个叫做RollVM的解释器来完成脚本语言的解析。

DiceScript将更好的实现骰点功能，语法规范化的同时，具有更好的接口设计和自定义能力。


几个设计原则：

* 从主流语言和跑团软件中借鉴语法，如Golang/Python/JS/Ruby/Fvtt/BCDice，不随意发明
* 兼容中国大陆地区跑团指令的一般习惯
* 要具有较强的配置和扩展能力，符合trpg场景额需求
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
- [x] 骰点运算 流行: d20, 3d20, (4+5)d(20), 2d20k1, 2d20q1
- [x] 骰点运算 - fvtt语法: 2d20kl, 2d20kh, 2d20dl, 2d20dh, d20min10, d20max10
- [x] 骰点运算 - CoC / Fate / WoD / Double Cross
- [ ] 骰点运算 - 自定义算符
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
- [ ] 测试覆盖率 85% / 90%

## 更新记录

[更新记录](./CHANGELOG.md)

## TODO

* ~~d2d(4d4d5)d6 计算过程问题~~

## 开发

如果修改了文法，使用这个工具重新生成:
```
go install github.com/fy0/pigeon@latest
pigeon -nolint -optimize-parser -optimize-ref-expr-by-index -o .\roll.peg.go .\roll.peg
```
