package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func main() {
	// viper.SetConfigName("config")
	// viper.SetConfigType("yaml")
	// viper.AddConfigPath("./test/testViper/cfg")
	viper.SetConfigFile("./test/testViper/cfg/ignore.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}
	name := viper.Get("name")
	fmt.Println(name)
	hbs := viper.Get("hobbies")
	fmt.Println(hbs)
	hbsls := viper.GetStringSlice("hobbies")
	fmt.Println(hbsls)
}
