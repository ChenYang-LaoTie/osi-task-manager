package controllers

import (
	"github.com/astaxie/beego/logs"
	"osi-task-manager/models"
)

func CheckLoginUser(token string) bool {
	lut := models.LoginUserToken{AesKey: token}
	luErr := models.QueryLoginUser(&lut, "AesKey")
	if luErr != nil {
		logs.Error("User is not logged in, luErr: ", luErr)
		return false
	}
	return true
}
