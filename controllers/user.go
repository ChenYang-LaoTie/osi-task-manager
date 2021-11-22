/*
1. Upload student user information;
2. Upload mentor user information;
*/
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"os"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"osi-task-manager/taskhandler"
)

type InternStudentControllers struct {
	beego.Controller
}

type Student struct {
	GiteeId string `json:"giteeId"`
	Email   string `json:"email"`
	Status  int    `json:"status"`
}

type UserInfo struct {
	Token       string    `json:"token"`
	StudentData []Student `json:"studentData"`
}

// @Title internStudent
// @Description internStudent
// @Param	body		body 	models.internStudent	true		"body for user content"
// @Success 200 {int} models.internStudent
// @Failure 403 body is empty
// @router / [post]
func (u *InternStudentControllers) Post() {
	var userData UserInfo
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	json.Unmarshal(u.Ctx.Input.RequestBody, &userData)
	//logs.Info("Cve upload request parameters: ", string(u.Ctx.Input.RequestBody))
	if len(userData.Token) > 1 {
		tokenBool := CheckLoginUser(userData.Token)
		if !tokenBool {
			u.Ctx.ResponseWriter.WriteHeader(406)
			u.Ctx.WriteString("upload failed")
			return
		}
	} else {
		u.Ctx.ResponseWriter.WriteHeader(403)
		u.Ctx.WriteString("upload failed")
		return
	}
	if len(userData.StudentData) == 0 {
		logs.Error("student data is empty")
		u.Ctx.ResponseWriter.WriteHeader(405)
		u.Ctx.WriteString("data failed")
		return
	}
	accessToken := os.Getenv("GITEE_SRC_TOKEN")
	for _, stdData := range userData.StudentData {
		defer common.Catchs()
		logs.Info("Each request parameter: ", stdData)
		gui := taskhandler.GetGitId(accessToken, stdData.GiteeId)
		if gui.GitId == 0 {
			logs.Error("GetGitId, Failed to get gitid information: ", gui)
			continue
		}
		su := models.StdUser{GitId: gui.GitId}
		if stdData.Status == 2 {
			// delete data
			delErr := models.DelStdUser(&su, "GitId")
			if delErr != nil {
				logs.Error("InternStudentControllers, delErr: ", delErr)
			}
		} else {
			quErr := models.QueryStdUser(&su, "GitId")
			if quErr != nil {
				logs.Error("InternStudentControllers, quErr: ", quErr)
			}
			if su.UserId > 0 {
				su.Status = 1
				su.UpdateTime = common.GetCurTime()
				su.EmailAddr = stdData.Email
				su.GitUserId = gui.Login
				su.UserName = gui.Login
				su.TestFlag = 2
				upErr := models.UpdateStdUser(&su, "Status", "UpdateTime", "EmailAddr", "GitUserId", "TestFlag")
				if upErr != nil {
					logs.Error("InternStudentControllers, upErr: ", upErr)
				}
			} else {
				su.EmailAddr = stdData.Email
				su.Status = 1
				su.GitUserId = gui.Login
				su.CreateTime = common.GetCurTime()
				su.TestFlag = 2
				su.UserName = gui.Login
				su.GitId = gui.GitId
				num, inErr := models.InsertStdUser(&su)
				if inErr != nil {
					logs.Error("InternStudentControllers, inErr: ", inErr, num)
				}
			}
		}
	}
	u.Ctx.ResponseWriter.WriteHeader(200)
	u.Ctx.WriteString("success")
	return
}

type InternTutorControllers struct {
	beego.Controller
}

type Tutor struct {
	GiteeId string `json:"giteeId"`
	Email   string `json:"email"`
	Status  int    `json:"status"`
}

type TutorInfo struct {
	Token     string  `json:"token"`
	TutorData []Tutor `json:"tutorData"`
}

// @Title TutorInfo
// @Description TutorInfo
// @Param	body		body 	models.TutorInfo	true		"body for user content"
// @Success 200 {int} models.TutorInfo
// @Failure 403 body is empty
// @router / [post]
func (u *InternTutorControllers) Post() {
	var tutorData TutorInfo
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr, ",Header: ", req.Header)
	json.Unmarshal(u.Ctx.Input.RequestBody, &tutorData)
	//logs.Info("Cve upload request parameters: ", string(u.Ctx.Input.RequestBody))
	if len(tutorData.Token) > 1 {
		tokenBool := CheckLoginUser(tutorData.Token)
		if !tokenBool {
			u.Ctx.ResponseWriter.WriteHeader(406)
			u.Ctx.WriteString("upload failed")
			return
		}
	} else {
		u.Ctx.ResponseWriter.WriteHeader(403)
		u.Ctx.WriteString("upload failed")
		return
	}
	if len(tutorData.TutorData) == 0 {
		logs.Error("tutor data is empty")
		u.Ctx.ResponseWriter.WriteHeader(405)
		u.Ctx.WriteString("data failed")
		return
	}
	accessToken := os.Getenv("GITEE_SRC_TOKEN")
	for _, tutData := range tutorData.TutorData {
		defer common.Catchs()
		logs.Info("Each request parameter: ", tutData)
		gui := taskhandler.GetGitId(accessToken, tutData.GiteeId)
		if gui.GitId == 0 {
			logs.Error("GetGitId, Failed to get gitid information: ", gui)
			continue
		}
		su := models.TutUser{GitId: gui.GitId}
		if tutData.Status == 2 {
			// delete data
			delErr := models.DelTutUser(&su, "GitId")
			if delErr != nil {
				logs.Error("InternStudentControllers1, delErr: ", delErr)
			}
		} else {
			quErr := models.QueryTutUser(&su, "GitId")
			if quErr != nil {
				logs.Error("InternStudentControllers1, quErr: ", quErr)
			}
			if su.UserId > 0 {
				su.Status = 1
				su.UpdateTime = common.GetCurTime()
				su.EmailAddr = tutData.Email
				su.GitUserId = gui.Login
				su.UserName = gui.Login
				upErr := models.UpdateTutUser(&su, "Status", "UpdateTime", "EmailAddr", "GitUserId")
				if upErr != nil {
					logs.Error("InternStudentControllers1, upErr: ", upErr)
				}
			} else {
				su.EmailAddr = tutData.Email
				su.Status = 1
				su.GitUserId = gui.Login
				su.CreateTime = common.GetCurTime()
				su.UserName = gui.Login
				su.GitId = gui.GitId
				num, inErr := models.InsertTutUser(&su)
				if inErr != nil {
					logs.Error("InternStudentControllers1, inErr: ", inErr, num)
				}
			}
		}
	}
	u.Ctx.ResponseWriter.WriteHeader(200)
	u.Ctx.WriteString("success")
	return
}

type InternAdminLoginControllers struct {
	beego.Controller
}

type Result struct {
	Token  string `json:"token"`
	UserId int  `json:"userId"`
}

type LoginResData struct {
	LoginData Result `json:"loginData"`
	Mesg      string `json:"message"`
	Code      int    `json:"code"`
}

func (c *InternAdminLoginControllers) RetData(resp LoginResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title UserLogin
// @Description UserLogin
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {int} models.User.ID
// @Failure 403 body is empty
// @router / [post]
func (u *InternAdminLoginControllers) Post() {
	req := make(map[string]interface{})
	resp := LoginResData{}
	resp.Code = 405
	resp.Mesg = "Login failed"
	resp.LoginData = Result{}
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &req)
	if err != nil {
		logs.Error(err)
		u.RetData(resp)
		return
	}
	//Judge whether it is legal
	if req["userName"] == nil || req["passWord"] == nil {
		resp.Mesg = "Data error: username or password"
		logs.Error("Data error: username or password")
		u.RetData(resp)
		return
	}
	password := fmt.Sprintf("%s", req["passWord"])
	username := fmt.Sprintf("%s", req["userName"])
	loginInfo := models.LoginUserToken{UserName: username, PassWord: password}
	luErr := models.QueryLoginUser(&loginInfo, "UserName", "PassWord")
	if luErr != nil {
		logs.Error("Administrator login information does not exist, luErr: ", luErr)
		resp.Mesg = "Administrator login information does not exist"
		u.RetData(resp)
		return
	}
	if loginInfo.UerId > 0 {
		resp.LoginData.UserId = loginInfo.UerId
		resp.LoginData.Token = loginInfo.AesKey
		resp.Code = 200
		resp.Mesg = "success"
	}
	u.RetData(resp)
	return
}
