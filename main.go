package main

import (
	"github.com/astaxie/beego/orm"
	_ "server-purchase/routers"

	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	orm.Debug, _ = beego.AppConfig.Bool("ormdebug")
	beego.Run()
}
