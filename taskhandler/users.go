package taskhandler

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"osi-task-manager/util"
)

type GitUserInfo struct {
	GitId int64
	Login string
	Name  string
	Url   string
}

// Get gitid related information through username
func GetGitId(gitToken, userName string) (gui GitUserInfo) {
	getUrl := fmt.Sprintf(`https://gitee.com/api/v5/users/%v?access_token=%v`, userName, gitToken)
	gitUserDict, gitErr := util.HTTPGitGet(getUrl)
	if gitErr != nil {
		logs.Error("GetGitId, gitErr:", gitErr)
		return
	}
	if len(gitUserDict) > 0 && gitUserDict["id"] != nil {
		gui.GitId = int64(gitUserDict["id"].(float64))
		if gui.GitId > 0 {
			gui.Login = gitUserDict["login"].(string)
			gui.Name = gitUserDict["name"].(string)
			gui.Url = gitUserDict["url"].(string)
		}
	}
	return
}

func UpdateStdUsers(su *models.StdUser, name, userName string) {
	if su.UserName != userName || su.GitUserId != name {
		su.UpdateTime = common.GetCurTime()
		su.GitUserId = name
		su.UserName = userName
		upErr := models.UpdateStdUser(su, "UpdateTime", "GitUserId", "UserName")
		if upErr == nil {
			logs.Info("UpdateStdUsers success, upErr:", upErr)
		} else {
			logs.Error("UpdateStdUsers failed, upErr: ", upErr)
		}
	}
}

func UpdateTutUsers(tu *models.TutUser, name, userName string) {
	if tu.UserName != userName || tu.GitUserId != name {
		tu.UpdateTime = common.GetCurTime()
		tu.GitUserId = name
		tu.UserName = userName
		upErr := models.UpdateTutUser(tu, "UpdateTime", "GitUserId", "UserName")
		if upErr == nil {
			logs.Info("UpdateTutUsers success, upErr:", upErr)
		} else {
			logs.Error("UpdateTutUsers failed, upErr: ", upErr)
		}
	}
}
