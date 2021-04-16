package data

import "fmt"

func AlamrStart(d interface{}){
	data, ok := d.(AlarmStartJSON)
	if !ok{
		panic(fmt.Errorf("data was not an alarmStartJSON"))
	}
	fmt.Println(data)
}


