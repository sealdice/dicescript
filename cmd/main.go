package main

import (
	"fmt"
	"github.com/peterh/liner"
	dice "github.com/sealdice/dicescript"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
		line.ReadHistory(f)
		f.Close()
	}

	attrs := map[string]*dice.VMValue{}

	fmt.Println("DiceScript Shell v0.0.1")
	ccTimes := 0
	vm := dice.NewVM()
	vm.Config.EnableDiceWoD = true
	vm.Config.EnableDiceCoC = true
	vm.Config.EnableDiceFate = true
	vm.Config.EnableDiceDoubleCross = true
	vm.Config.PrintBytecode = true
	vm.Config.CallbackSt = func(_type string, name string, val *dice.VMValue, extra *dice.VMValue, op string, detail string) {
		fmt.Println("st:", _type, name, val.ToString(), extra.ToString(), op, detail)
	}

	vm.Config.IgnoreDiv0 = true
	vm.Config.DefaultDiceSideExpr = "面数 ?? 50"
	vm.Config.OpCountLimit = 30000

	vm.Config.CallbackLoadVar = func(name string) (string, *dice.VMValue) {
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
				return name, dice.VMValueNewInt(v)
			}
		}

		fmt.Println("COC值:", name, cocFlagVarPrefix)
		return name, nil
	}

	_ = vm.RegCustomDice(`E(\d+)`, func(ctx *dice.Context, groups []string) *dice.VMValue {
		return dice.VMValueNewInt(2)
	})

	//vm.ValueStoreNameFunc = func(name string, v *dice.VMValue) {
	//	attrs[name] = v
	//}

	re := regexp.MustCompile(`^(\D+)(\d+)$`)

	vm.GlobalValueLoadFunc = func(name string) *dice.VMValue {
		m := re.FindStringSubmatch(name)
		if len(m) > 1 {
			//val, _ := strconv.ParseInt(m[2], 10, 64)
			//return dice.VMValueNewInt(val)
			return dice.VMValueNewInt(0)
		}

		if val, ok := attrs[name]; ok {
			return val
		}
		return nil
	}

	for true {
		if text, err := line.Prompt(">>> "); err == nil {
			if strings.TrimSpace(text) == "" {
				continue
			}
			line.AppendHistory(text)

			err := vm.Run(text)
			//fmt.Println(vm.GetAsmText())
			if err != nil {
				fmt.Printf("错误: %s\n", err.Error())
			} else {
				rest := vm.RestInput
				if rest != "" {
					rest = fmt.Sprintf(" 剩余文本: %s", rest)
				}
				fmt.Printf("过程: %s\n", vm.Detail)
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
