package main

import (
	"flag"
	"fmt"

	"github.com/The-True-Hooha/NimbleFiles/internal/cmd"
)

func init(){
	
}

func main(){
	
	command := cmd.InitCommands()
	flag.Parse()
	 

	if flag.NArg() == 0 {
		fmt.Println("welcome, see the available commands")
		for _, cmd := range command.DisplayCommands(){
			fmt.Printf("  %s: %s\n", cmd.Name, cmd.Description)
		}
		return
	}
	
	name, args := cmd.ParseCommand(flag.Arg(0))
	err := command.Execute(name, args)
	if err != nil {
		fmt.Printf("some error here %s", err)
	}


}