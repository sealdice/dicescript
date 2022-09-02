package main

import (
	"fmt"
	"github.com/peterh/liner"
	"github.com/sealdice/dicescript"
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

	line.SetCtrlCAborts(true)
	line.SetCompleter(func(line string) (c []string) {
		return
	})

	if f, err := os.Open(historyFn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	attrs := map[string]*dicescript.VMValue{}

	fmt.Println("DiceScript Shell v0.0.0")
	ccTimes := 0
	for true {
		if text, err := line.Prompt(">>> "); err == nil {
			if strings.TrimSpace(text) == "" {
				continue
			}
			line.AppendHistory(text)

			vm := dicescript.NewVM()
			vm.Flags.PrintBytecode = true
			vm.ValueStoreNameFunc = func(name string, v *dicescript.VMValue) {
				attrs[name] = v
			}
			vm.ValueLoadNameFunc = func(name string) *dicescript.VMValue {
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
