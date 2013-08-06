package main

import (
	"fmt"

	"github.com/kvu787/go-schedule/crawler/database"
)

func main() {
	i, err := database.GetSwitch()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(i)
	err = database.UpdateSwitch(2, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	i, err = database.GetSwitch()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(i)
}
