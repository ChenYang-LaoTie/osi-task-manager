package main

import (
	"osi-task-manager/common"
	"osi-task-manager/models"
	_ "osi-task-manager/routers"
	"github.com/astaxie/beego"
	"osi-task-manager/task"
)

func init() {
	// Initialization log
	common.LogInit()
}

func main() {
	// init db
	dbOk := models.Initdb()
	if !dbOk {
		println("error: Database initialization failed")
		return
	}
	// Initialize a scheduled task
	taskOk := task.InitTask()
	if !taskOk {
		println("error: Timing task initialization failed, the program ends")
		task.StopTask()
		return
	}
	task.StartTask()
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}

