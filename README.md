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
- [ ] 计算过程显示
- [ ] 报错信息优化
- [x] 线程安全
- [x] 变量作用域
- [ ] sourcemap
- [ ] 测试覆盖率 74% / 90%

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
	if err := vm.Run(`d20`); err == nil {
		fmt.Printf("结果: %s\n", vm.Ret.ToString())
	} else {
		fmt.Printf("错误: %s\n", err.Error())
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

int 使用 int64，任何int与float运算的操作都会使得结果成为float

float 使用 float64

示例:
```
123
123.456
.456
```


#### String 字符串

字符串是不可变类型。支持四种字符串定义，前三种为：
```
'123'
"123"
`123`
```

其中`字符串支持模板语法，和python的f-string模板语法大致相同，即用{}来取值：

```
name = '张三'
`你好，{name}`
```

特别的，如果你需要在文本中插入一段程序，但又不希望影响输出\[尚未实装]：

```
`你好，{% '这段文本是不会显示出来的' %}`
```

刚才未提到的第四种字符串，用的符号是 \x1e，他跟`的作用完全相同，也支持 f-string

这个是专用于跑团机器人环境的，举个例子，你希望用户能在一段文本中插入变量：

```go
vm.Run("\x1e" + input + "\x1e")
```

这样拿到的结果就是一个字符串了。你可以将上面代码中的"\x1e"换成\`，效果是一样的，但是用户就不能在文本里输入`了

同时，字符串支持分片语法，写法同python，暂不支持步长：
```python
'12345'[2:4]  // 34
```

#### 空值

null 空值
undefined 未定义


#### Computed 计算类型

这种类型的意思是，最终得到的值是一个式子计算的结果，例如:

```
&砍一刀 = D20 + 4
```

```
砍一刀 + 10  // 此时为 D20 + 4 + 10，每次调用时会动态计算一遍
```

同时我们可以实现更高级的功能：

```
&a = this.x + d10
a.x = 5
```

```
a // 获得结果 5[a.x] + d10
```

这对一些二级属性非常有用，可以提前录入公式，只改变其中的变量。

海豹的RollVM中，DND的技能实际上就是这种类型



#### Array 数组

有两种方式定义一个数组。一是\[1,2,3,4,5]，二是\[1..5]会生成一个包含12345的数组，也可以写\[5..1]生成出一个反的数组

支持一个fvtt的特殊语法，如\[1d20, 10]kh，是为两者取高

数组可以装入任意类型，也可以装入多维数组

```
a = [1,2,'test', [4,5,6]]
```

通过下标可以取得数组内容：
```
a[0]
a[3][1]
```

分片语法：
```python
[1,2,3,4,5][2:4]  // [3,4]

a = [1,2,3]; a[2:3] = [4,5,6] // a == [1, 2, 4, 5, 6]
```


### 关键字

if

else

for

break

return

undefined

this

global


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

变量名规则与大多数主流语言，如C/Python/Go/JS基本一致，特别的是允许$字符作为变量名。

实际操作中可以使用特殊变量名来实现一些特别功能，例如SealDice中是这样做的：

如果不以$打头，为当前TRPG的技能或属性，例如: 侦查、聆听、HP等

如以$打头：

$t开头为临时变量，不定期删除
$m开头为个人变量，在当前用户的所有群组有效
$g开头为群组变量，当前群所有用户共享

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

floor ceil round abs int float str

### 注释

使用 // 注释

#### 脚本中定义函数

```
func test(a, b, c) {
    'hello world, ' + this.a
}
```

进行调用
```
test('张三') // 获得文本 hello world, 张三
```

示例，斐波那契数列计算：
```
func fib(n) {
  this.n == 0 ? 0,
  this.n == 1 ? 1,
  this.n == 2 ? 1,
   1 ? fib(this.n-1)+fib(this.n-2)
}
fib(11) // 89
```

另一种写法：
```
func fib(n) {
  if this.n == 0 { 0 }
  else if this.n == 1 { 1 }
  else if this.n == 2 { 1 } else {
    fib(this.n-1) + fib(this.n-2)
  }
}
fib(10) // 55
```

### 语句块(还未想好)

{} 是一个语句块，内含至少1条语句。

通常来讲，一个语句块的值是其最后一条语句。如 { 1; 2; 3 } 的值是3



### 接口

#### 属性模板

#### 事件

#### 计算过程

#### 格式化输出

