package main

import (
	"fmt"
	"github.com/abiosoft/ishell/v2"
	"github.com/sealdice/dicescript"
	"strings"
)

func main() {
	shell := ishell.New()
	shell.Println("DiceScript Shell")

	shell.AddCmd(&ishell.Cmd{
		Name:    "run",
		Aliases: []string{"r", "eval"},
		Help:    "执行脚本",
		Func: func(c *ishell.Context) {
			vm := dicescript.NewVM()
			err := vm.Run(strings.Join(c.Args, " "))

			if err != nil {
				c.Err(err)
			} else {
				rest := vm.RestInput
				if rest != "" {
					rest = fmt.Sprintf(" 剩余文本: %s", rest)
				}
				c.Printf("结果: %s%s\n", vm.Ret.ToString(), rest)
			}
		},
	})

	shell.Run()
}
