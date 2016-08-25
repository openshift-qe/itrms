package main

import (
	"github.com/astaxie/beego"
	"github.com/jasonlvhit/gocron"
	"github.com/wanghaoran1988/itrms/jobs"
	_ "github.com/wanghaoran1988/itrms/routers"
)

func startScheduleJobs() {
	s := gocron.NewScheduler()
	s.Every(60).Seconds().Do(jobs.UpdateImageIDTask)
	<-s.Start()
}
func main() {
	go startScheduleJobs()
	beego.SetStaticPath("bower_components", "bower_components")
	beego.SetStaticPath("dist", "static/dist")
	beego.SetStaticPath("js", "static/js")
	beego.SetStaticPath("css", "static/css")
	beego.SetStaticPath("less", "static/less")
	beego.SetStaticPath("pages", "views")
	beego.SetLogger("file", `{"filename":"logs/irtms.log"}`)
	beego.Run()

}
