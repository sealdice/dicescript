package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/peterh/liner"
	ds "github.com/sealdice/dicescript"
)

var (
	historyFn = filepath.Join(os.TempDir(), ".dicescript_history")
)

func main() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	line.SetCompleter(func(line string) (c []string) {
		return
	})

	if f, err := os.Open(historyFn); err == nil {
		_, _ = line.ReadHistory(f)
		_ = f.Close()
	}

	attrs := map[string]*ds.VMValue{}

	fmt.Println("DiceScript Shell v0.0.1")
	ccTimes := 0
	vm := ds.NewVM()
	vm.Config.EnableDiceWoD = true
	vm.Config.EnableDiceCoC = true
	vm.Config.EnableDiceFate = true
	vm.Config.EnableDiceDoubleCross = true
	vm.Config.PrintBytecode = true
	vm.Config.CallbackSt = func(_type string, name string, val *ds.VMValue, extra *ds.VMValue, op string, detail string) {
		fmt.Println("st:", _type, name, val.ToString(), extra.ToString(), op, detail)
	}

	vm.Config.IgnoreDiv0 = true
	vm.Config.DefaultDiceSideExpr = "面数 ?? 50"
	vm.Config.OpCountLimit = 30000

	vm.Config.HookValueLoadPre = func(ctx *ds.Context, name string) (string, *ds.VMValue) {
		re := regexp.MustCompile(`^(困难|极难|大成功|常规|失败|困難|極難|常規|失敗)?([^\d]+)(\d+)?$`)
		m := re.FindStringSubmatch(name)
		var cocFlagVarPrefix string

		if len(m) > 0 {
			if m[1] != "" {
				cocFlagVarPrefix = m[1]
				name = name[len(m[1]):]
			}

			// 有末值时覆盖，有初值时
			if m[3] != "" {
				v, _ := strconv.ParseInt(m[3], 10, 64)
				fmt.Println("COC值:", name, cocFlagVarPrefix)
				return name, ds.NewIntVal(ds.IntType(v))
			}
		}

		fmt.Println("COC值:", name, cocFlagVarPrefix)
		return name, nil
	}

	_ = vm.RegCustomDice(`E(\d+)`, func(ctx *ds.Context, groups []string) (*ds.VMValue, string, error) {
		if len(groups) < 2 {
			return nil, "", fmt.Errorf("自定义骰子算符未匹配")
		}
		v, err := strconv.ParseInt(groups[1], 10, 64)
		if err != nil {
			return nil, "", err
		}
		return ds.NewIntVal(ds.IntType(v)), "E" + groups[1], nil
	})

	// 阶乘计算函数
	factorial := func(n int64) int64 {
		if n < 0 {
			return 0
		}
		if n == 0 || n == 1 {
			return 1
		}
		result := int64(1)
		for i := int64(2); i <= n; i++ {
			result *= i
		}
		return result
	}

	// 组合数计算函数 C(a,b) = a! / (b! * (a-b)!)
	combination := func(a, b int64) int64 {
		if b < 0 || a < 0 || b > a {
			return 0
		}
		if b == 0 || b == a {
			return 1
		}
		// 优化：使用 C(a,b) = C(a, a-b) 选择较小的 b
		if b > a-b {
			b = a - b
		}

		result := int64(1)
		for i := int64(0); i < b; i++ {
			result = result * (a - i) / (i + 1)
		}
		return result
	}

	// 注册阶乘算符
	_ = vm.RegCustomDice(`(\d+)!`, func(ctx *ds.Context, groups []string) (*ds.VMValue, string, error) {
		if len(groups) < 2 {
			return nil, "", fmt.Errorf("阶乘算符格式错误")
		}
		n, err := strconv.ParseInt(groups[1], 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("参数解析错误: %v", err)
		}
		result := factorial(n)
		detail := fmt.Sprintf("%d!=%d", n, result)
		return ds.NewIntVal(ds.IntType(result)), detail, nil
	})

	// 注册 Ca,b 算符
	_ = vm.RegCustomDice(`C(\d+),(\d+)`, func(ctx *ds.Context, groups []string) (*ds.VMValue, string, error) {
		if len(groups) < 3 {
			return nil, "", fmt.Errorf("Ca,b 算符格式错误")
		}

		a, err := strconv.ParseInt(groups[1], 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("参数 a 解析错误: %v", err)
		}

		b, err := strconv.ParseInt(groups[2], 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("参数 b 解析错误: %v", err)
		}

		result := combination(a, b)
		detail := ""

		return ds.NewIntVal(ds.IntType(result)), detail, nil
	})

	// vm.ValueStoreNameFunc = func(name string, v *dice.VMValue) {
	//	attrs[name] = v
	// }

	re := regexp.MustCompile(`^(\D+)(\d+)$`)

	vm.GlobalValueLoadFunc = func(name string) *ds.VMValue {
		m := re.FindStringSubmatch(name)
		if len(m) > 1 {
			// val, _ := strconv.ParseInt(m[2], 10, 64)
			// return dice.NewIntVal(val)
			return ds.NewIntVal(0)
		}

		if val, ok := attrs[name]; ok {
			return val
		}
		return nil
	}

	for {
		if text, err := line.Prompt(">>> "); err == nil {
			if strings.TrimSpace(text) == "" {
				continue
			}
			line.AppendHistory(text)

			err := vm.Run(text)
			// fmt.Println(vm.GetAsmText())
			if err != nil {
				fmt.Printf("错误: %s\n", err.Error())
			} else {
				rest := vm.RestInput
				if rest != "" {
					rest = fmt.Sprintf(" 剩余文本: %s", rest)
				}
				fmt.Printf("过程: %s\n", vm.GetDetailText())
				fmt.Printf("结果: %s%s\n", vm.Ret.ToString(), rest)
				fmt.Printf("栈顶: %d 层数:%d 算力: %d\n", vm.StackTop(), vm.Depth(), vm.NumOpCount)
			}

		} else if err == liner.ErrPromptAborted {
			if ccTimes >= 0 {
				fmt.Print("Interrupted")
				break
			} else {
				ccTimes += 1
				fmt.Println("Input Ctrl-c once more to exit")
			}
		} else {
			fmt.Print("Error reading line: ", err)
		}
	}

	if f, err := os.Create(historyFn); err != nil {
		fmt.Println("Error writing history file: ", err)
	} else {
		_, _ = line.WriteHistory(f)
		_ = f.Close()
	}
}
