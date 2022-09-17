package main

import (
	"fmt"
	"github.com/peterh/liner"
	dice "github.com/sealdice/dicescript"
	"os"
	"path/filepath"
	"strings"
)

var (
	historyFn = filepath.Join(os.TempDir(), ".dicescript_history")
)

func main() {
	line := liner.NewLiner()
	defer line.Close()

	a, b := dice.VMValueNewFloat(3.2).ToJSON()
	fmt.Println("!!!", string(a), b)

	//a, b = dice.VMValueNewComputed("1 + this.x + d10").ToJSON()
	//fmt.Println("!!!", string(a), b)
	v, _ := dice.NewVMWithStore(nil)
	v.Run(`func a(x) { return 5 }; a`)
	aa, _ := v.Ret.ToJSON()
	fmt.Println("!!!!", string(aa), v.Ret)

	v, _ = dice.NewVMWithStore(nil)
	v.Run(`[1,2,3]`)
	aa, _ = v.Ret.ToJSON()
	fmt.Println("!!!!", string(aa), v.Ret)
	dice.VMValueFromJSON(aa)

	line.SetCtrlCAborts(true)
	line.SetCompleter(func(line string) (c []string) {
		return
	})

	if f, err := os.Open(historyFn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	attrs := map[string]*dice.VMValue{}

	fmt.Println("DiceScript Shell v0.0.0")
	ccTimes := 0
	for true {
		if text, err := line.Prompt(">>> "); err == nil {
			if strings.TrimSpace(text) == "" {
				continue
			}
			line.AppendHistory(text)

			vm := dice.NewVM()
			vm.Flags.PrintBytecode = true
			vm.ValueStoreNameFunc = func(name string, v *dice.VMValue) {
				attrs[name] = v
			}
			vm.ValueLoadNameFunc = func(name string) *dice.VMValue {
				if val, ok := attrs[name]; ok {
					return val
				}
				return nil
			}

			err := vm.Run(text)
			if err != nil {
				fmt.Printf("错误: %s\n", err.Error())
			} else {
				rest := vm.RestInput
				if rest != "" {
					rest = fmt.Sprintf(" 剩余文本: %s", rest)
				}
				fmt.Printf("结果: %s%s\n", vm.Ret.ToString(), rest)
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
