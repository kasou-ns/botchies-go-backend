//命名ルールについて
/*
構造体（のメンバ）　→　先頭はUpper
変数　→　先頭はlower(グローバルはUpperなので注意)
json　→　_でつなぐ！
 */

package main

import (
	"wakeUpCall/app/controllers"
	"wakeUpCall/config"
	"wakeUpCall/utils"
)

func main(){
	utils.LoggingSettings(config.Config.LogFile)
	controllers.StartWebServer()
}
