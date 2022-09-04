# DiceScript

[![Go Report Card](https://goreportcard.com/badge/github.com/sealdice/dicescript)](https://goreportcard.com/report/github.com/sealdice/dicescript)
![Software License](https://img.shields.io/badge/license-Apache2-brightgreen.svg?style=flat-square)
[![GoDoc](https://godoc.org/github.com/sealdice/dicescript?status.svg)](https://godoc.org/github.com/sealdice/dicescript)

最符合国内跑团习惯的TRPG骰点脚本语言。

Simple script language for TRPG dice engine.

特性:
- 易于使用，方便扩展
- 稳定可靠，极高的测试覆盖率
- 免费，并可商用
- 支持JavaScript
- 海豹TRPG骰点核心的第二代解释器

进度:

- [x] 基础类型 int float string
- [x] 基础类型 undefined null
- [x] 一元算符 + -
- [x] 二元算符 +-*/% >,>=,==,!=,<,<=
- [x] 三元算符 ? :
- [x] 骰点运算 流行: d20, 3d20, (4+5)d(20), 2d20k1, 2d20q1 
- [x] 骰点运算 - fvtt语法: 2d20kl, 2d20kh, 2d20dl, 2d20dh, d20min10, d20max10
- [ ] 骰点运算 - 自定义算符
- [ ] 骰点运算 - Fate / WOD / Double Cross
- [x] 高级类型 数组array
- [ ] 高级类型 字典dict
- [x] 高级类型 计算数值computed
- [x] 逻辑语法 if ... else ..
- [ ] 逻辑语法 for
- [ ] 函数支持
- [ ] 内置函数
- [x] 变量支持
- [ ] 序列化和反序列化
- [ ] 计算过程显示
- [ ] 报错信息优化
- [ ] 测试覆盖率 73% / 90%

测试页面:

https://sealdice.github.io/dicescript/

## 如何使用

Golang:
```go
package main

import (
	"fmt"
	dice "github.com/sealdice/dicescript"
)

func main() {
	vm := dice.NewVM()

	// 如果需要使用变量，那么接入一下ValueStoreNameFunc和ValueLoadNameFunc
	// 不需要就跳过
	attrs := map[string]*dice.VMValue{}

	vm.ValueStoreNameFunc = func(name string, v *dice.VMValue) {
		attrs[name] = v
	}
	vm.ValueLoadNameFunc = func(name string) *dice.VMValue {
		if val, ok := attrs[name]; ok {
			return val
		}
		return nil
	}

	// 可以运算了
	err := vm.Run(`d20`)

	// 打印结果
	if err != nil {
		fmt.Printf("错误: %s\n", err.Error())
	} else {
		fmt.Printf("结果: %s\n", vm.Ret.ToString())
	}
}
```

JavaScript // 还会再调整API
```javascript
function roll(text) {
    let ctx = dice.newVM();
    try {
        ctx.Run(text)
        if (ctx.Error) {
            console.log(`错误: ${ctx.Error.Error()}`)
        } else {
            console.log(`结果: ${ctx.Ret.ToString()}`)
        }
    } catch (e) {
        this.items.push(`错误: 未知错误`)
    }
}
```

## 更新记录

#### 2022.9.4

* 多维数组
* computed 计算类型
* 计算属性扩展 &a = 1d1 + x; a.x = 1 此时有 a == 2

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


## 草案: 混沌的思绪

先随便写一些想到的，然后再细化

这是未来SealDice项目的一部分，一次从零重构。

目前SealDice使用一个叫做RollVM的解释器来完成脚本语言的解析。

DiceScript将使用和RollVM相同的技术栈，但会有更好的接口设计，增强自定义能力的同时，剥离耦合的部分。



几个设计原则：

* 从主流语言和跑团软件中借鉴语法，如Golang/Python/JS/Ruby/Fvtt/BCDice，不随意发明
* 兼容中国大陆地区跑团指令的一般习惯
* 要具有不错的配置和扩展能力，且不用重新编译的情况下就能具备
* 一定限度内容忍全角符号
* 兼容gopherjs
* 良好的错误提示文本
* 支持多线程



### 数据类型

#### Number数字类型 

一个 Int 一个 Float

兼任Bool类型



#### String 字符串

字符串，懂得都懂



#### 空值

null 空值
undefined 未定义


#### Computed 计算类型

语法应该会这样选一个：

&砍一刀 = D20 + 4

砍一刀 := D20 + 4

海豹的RollVM中，DND的技能实际上就是这种类型



#### Array 数组

不出意外，应该是使用 [] 来代表数组

如fvtt: {1d20,10}kh 

这里有 [1d20, 10]kh



### 关键字

if

else

for

break



### 运算符



#### 骰子算符

**常驻规则 - 永远存在并可用**

d 常规骰子算符，用法举例 d20  2d20k1 2d20q1  d20优势

**可选规则 - 可开关**

[Fate] f  命运骰，随机骰4次，每骰结果可能是-1 0 1，记为- 0 +

[COC] b/p 奖励骰/惩罚骰

[双十字] c

[WOD/无限] a

[自定义] 这里留一个接口，初步定为符合 [a-zA-Z]\d* 的都可以在这里设置自己的解析结果





#### 数学算符

```
加减乘除余 + -* / %
乘方 ^ ** // 2 ** 3 或 2 ^ 3 即2的3次方
```



### 变量

出于简单考虑，不设计变量作用域（特别复杂的功能还是请用JS来实现）。

仍然是$t开头临时变量，临时变量加入生命期限 expiredTime



### 逻辑



#### 字符串模板

提供类似python的f-string的能力



#### 赋值语句

```
$tA = 1
```



#### 逻辑算符

```
> < == != >= <= && ||
```



#### 条件算符 (?)

```
灵视 >= 40 ? '如果灵视达到40以上，你就能看到这句话'
```



#### 多重条件算符 (? ,)

```
灵视 >= 80 ? '看得很清楚吗？',
灵视 >= 50 ? '不错，再靠近一点……',
灵视 >= 30 ? '仔细听……',
灵视 >= 0 ? '呵，无知之人。'
```



#### 三目运算符 (? :)

```
灵视 >= 40 ? '如果灵视达到40以上，你就能看到这句话' : '无知亦是幸运'
```



#### 条件语句

```
if $t0 > 10 {
    $t1 = "aaa"
} else {
    $t1 = 'bbb'
}
```



#### 循环

尚未决定使用哪种

```
// 参考: Pascal
for $t1 = 1 to 10 {

}
```



```
// 参考: C-like
for $t1 = 1; $t1 < 10; $t1 += 1 {
	break
}
```



```
// 参考: golang
for $t1 < 10 {
	$t1 += 1
}
```



```
// 参考: C-like
while $t1 < 10 {
	$t1 += 1
}
```



### 函数

#### 内置函数

floor ceil round abs



### 语句块(还未想好)

{} 是一个语句块，内含至少1条语句。

通常来讲，一个语句块的值是其最后一条语句。如 { 1; 2; 3 } 的值是3



### 接口

#### 属性模板

#### 事件

#### 计算过程

#### 格式化输出

