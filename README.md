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
- [ ] 报错信息优化
- [x] 线程安全
- [x] 变量作用域
- [ ] 测试覆盖率 81% / 90%

## 更新记录

#### 2024.6.5
* 随机种子设置，序列化和反序列
* 计算过程文本可以自定义了
* 计算过程文本的生成函数进行了分离，现在只在GetDetailText被调用时生成
* 将解析和执行分成两个阶段，可以分别单独调用
* 现在能够检查生成的字节码是否进行了骰点运算，这在部分指令中很有用

#### 2024.6.3
* 修改API，新建变量更为简洁
* 移除了undefined
* 修复了ToJSON的一个bug

#### 2024.5.12
* 使用 [自己修改的 pigeon](https://github.com/fy0/pigeon) 重构语法解析

#### 2023.9.15
* NativeObject 一种用于绑定Go对象的类型，行为类似于dict
* 为字典类型增加了 keys() values() items() 方法
* dir 函数，和python的dir函数类似，用于获取对象的所有属性

#### 2023.10.22
* 空值合并算符
* 文档调整

#### 2022.11.29

* 逻辑与/逻辑或/按位与/按位或
* 补全测试用例
* 编写文档

#### 2022.11.28

* Fate/WoD/DoubleCross 相关算符
* 计算过程完善


#### 2022.11.24

* 内置函数支持默认参数
* 简易原型链机制


#### 2022.11.23

* 计算过程显示


#### 2022.11.20

* 字典类型


#### 2022.11.18

* 变量作用域


#### 2022.9.17

* 序列化和反序列化: array native_function，全类型完成


#### 2022.9.12

* 内置函数: ceil floor round int float str
* 序列化和反序列化: int float str undefined null computed function


#### 2022.9.10

* 分片赋值，以及取值语法，支持array和str
* range语法 \[0..2] 为 \[0,1,2]，\[3..1] 为 \[3,2,1]

#### 2022.9.8

* while 语法
* return 语法
* 数组下标赋值
* 线程安全优化
* 现在可以使用 true / false 其值为 1 / 0
* break / continue 支持

#### 2022.9.4

* 多维数组
* computed 计算类型
* function 函数类型

#### 2022.9.3

* if else 语句
* undefined 类型
* a == 1 ? 1 : 2 三目运算符
* a == 1 ? 'A', a == 2 ? 'B', a == 3 : 'C'
* 一元算符 +1 -1
* 数组
* 数组下标
* fvtt语法: \[1,2,3]kh   \[1,2,3]kl

#### 2022.9.2

* 支持浮点数
* 支持字符串
* 支持变量
* RollVM的测试覆盖率提升至95%
* 能够编译到JS

#### 2022.9.1

* 数学和逻辑算符全类型支持(除computed和array之外)
* 初步的单元测试
* 异常机制
* 接口调整
* 实现了d算符
* 实现d算符语法，d20k / d20q / d20kh / d20hl / d20d / d20dl / d20dh

#### 2022.8.30

* 二元算符框架(+-*/等)
* 支持数学四则运算
* 支持比较算符(< <= == != > >=)

#### 2022.8.29

* 指令执行初步架构
* VM接口
* 简易REPL

#### 2022.8.25

* 建项目，初步文法

## TODO

* d2d(4d4d5)d6 计算过程问题

## 开发

如果修改了文法:
```
go install github.com/fy0/pigeon@latest
pigeon -nolint  -optimize-parser -optimize-ref-expr-by-index -o .\roll.peg.go .\roll.peg
```
