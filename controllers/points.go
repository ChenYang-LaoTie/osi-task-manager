/*
1. Get a list of points
2. Get the points details of a single user
*/
package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"osi-task-manager/models"
)

type InternPointsListControllers struct {
	beego.Controller
}

func (c *InternPointsListControllers) RetData(resp ResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type UserPointsData struct {
	UserId        int64  `json:"userId"`
	IntegralValue int64  `json:"integralValue"`
	GitUserId     string `json:"GiteeId"`
}

type ResData struct {
	TotalCount int64 `json:"totalCount"`
	UserPoints []UserPointsData
	Mesg       string `json:"message"`
	Code       int    `json:"code"`
}

// @Title Get pointslist
// @Description get pointslist
// @Param	status	int	true (0,1,2)
// @Success 200 {object} pointslist
// @Failure 403 :status is err
// @router / [get]
func (u *InternPointsListControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	resp := ResData{}
	var upd []UserPointsData
	resp.Code = 404
	resp.Mesg = "Data does not exist"
	resp.TotalCount = 0
	resp.UserPoints = upd
	token := u.GetString("token", "")
	tokenBool := CheckLoginUser(token)
	if !tokenBool {
		resp.Code = 403
		resp.Mesg = "Request parameter error"
		u.RetData(resp)
		return
	}
	count := models.QueryUserPointsCount()
	if count > 0 {
		resp.Code = 200
		resp.Mesg = "Data query is successful"
		resp.TotalCount = count
		currentPage, err := u.GetInt("currentPage", 1)
		if err != nil {
			logs.Error("err: ", err, ", currentPage: ", currentPage)
			resp.Code = 403
			resp.Mesg = "Request parameter error"
			resp.TotalCount = 0
			u.RetData(resp)
			return
		}
		pageSize, err := u.GetInt("pageSize", 100)
		if err != nil {
			logs.Error("err: ", err, ", pageSize: ", pageSize)
			resp.Code = 403
			resp.Mesg = "Request parameter error"
			resp.TotalCount = 0
			u.RetData(resp)
			return
		}
		pointsData := models.QueryTotalPointsData(currentPage, pageSize)
		if len(pointsData) > 0 {
			for _, pd := range pointsData {
				var up UserPointsData
				su := models.StdUser{UserId: pd.UserId}
				suErr := models.QueryStdUser(&su, "UserId")
				if suErr != nil {
					logs.Error("QueryStdUser, suErr: ", suErr)
				}
				up.UserId = pd.UserId
				up.GitUserId = su.GitUserId
				up.IntegralValue = pd.IntegralValue
				upd = append(upd, up)
			}
			resp.UserPoints = upd
			u.RetData(resp)
			return
		} else {
			resp.Code = 404
			resp.Mesg = "Data does not exist"
			resp.TotalCount = 0
			logs.Error("Data does not exist")
			u.RetData(resp)
			return
		}
	}
}

type InternUserPointsControllers struct {
	beego.Controller
}

func (c *InternUserPointsControllers) RetData(resp ResUserData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

type PointsData struct {
	IssueNumber   string `json:"issueNumber"`
	IntegralValue int64  `json:"integralValue"`
	GitUrl        string `json:"gitUrl"`
	Owner         string `json:"nameSpace"`
	RepoPath      string `json:"path"`
	CreateTime    string `json:"createTime"`
}

type ResUserData struct {
	TotalCount int64  `json:"totalCount"`
	UserId     int64  `json:"userId"`
	GitUserId  string `json:"GiteeId"`
	UserPoints []PointsData
	Mesg       string `json:"message"`
	Code       int    `json:"code"`
}

// @Title Get userPoints
// @Description get userPoints
// @Param	status	int	true (0,1,2)
// @Success 200 {object} userPoints
// @Failure 403 :status is err
// @router / [get]
func (u *InternUserPointsControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	resp := ResUserData{}
	var upd []PointsData
	resp.Code = 404
	resp.Mesg = "Data does not exist"
	resp.TotalCount = 0
	resp.UserId = 0
	resp.GitUserId = ""
	resp.UserPoints = upd
	token := u.GetString("token", "")
	tokenBool := CheckLoginUser(token)
	if !tokenBool {
		resp.Code = 403
		resp.Mesg = "Request parameter error"
		u.RetData(resp)
		return
	}
	userId, err := u.GetInt64("userId", 0)
	if err != nil || userId == 0 {
		logs.Error("err: ", err, ", userId: ", userId)
		resp.Code = 403
		resp.Mesg = "Request parameter error"
		resp.TotalCount = 0
		u.RetData(resp)
		return
	}
	resp.UserId = userId
	su := models.StdUser{UserId: userId}
	suErr := models.QueryStdUser(&su, "UserId")
	if suErr != nil {
		logs.Error("QueryStdUser, suErr: ", suErr)
	}
	resp.GitUserId = su.GitUserId
	count := models.QueryUserPointsDetailCount(userId)
	if count > 0 {
		resp.Code = 200
		resp.Mesg = "Data query is successful"
		resp.TotalCount = count
		currentPage, err := u.GetInt("currentPage", 1)
		if err != nil {
			logs.Error("err: ", err, ", currentPage: ", currentPage)
			resp.Code = 403
			resp.Mesg = "Request parameter error"
			resp.TotalCount = 0
			u.RetData(resp)
			return
		}
		pageSize, err := u.GetInt("pageSize", 100)
		if err != nil {
			logs.Error("err: ", err, ", pageSize: ", pageSize)
			resp.Code = 403
			resp.Mesg = "Request parameter error"
			resp.TotalCount = 0
			u.RetData(resp)
			return
		}
		pointsData := models.QueryUserPointsDetail(currentPage, pageSize, userId)
		if len(pointsData) > 0 {
			for _, pd := range pointsData {
				var up PointsData
				eoi := models.EulerOriginIssue{OrId: pd.OrId}
				eoiErr := models.QueryEulerOriginIssue(&eoi, "OrId")
				if eoiErr != nil {
					logs.Error("QueryEulerOriginIssue, eoiErr: ", eoiErr)
				}
				up.IntegralValue = pd.IntegralValue
				up.CreateTime = pd.CreateTime
				up.Owner = eoi.Owner
				up.IssueNumber = eoi.IssueNumber
				up.RepoPath = eoi.RepoPath
				up.GitUrl = eoi.GitUrl
				upd = append(upd, up)
			}
			resp.UserPoints = upd
			u.RetData(resp)
			return
		} else {
			resp.Code = 404
			resp.Mesg = "Data does not exist"
			resp.TotalCount = 0
			logs.Error("Data does not exist")
			u.RetData(resp)
			return
		}
	}
}

type InternUserSortPointsControllers struct {
	beego.Controller
}

func (c *InternUserSortPointsControllers) RetData(resp ResData) {
	c.Data["json"] = resp
	c.ServeJSON()
}

func (u *InternUserSortPointsControllers) Get() {
	req := u.Ctx.Request
	addr := req.RemoteAddr
	logs.Info("Method: ", req.Method, "Client request ip address: ", addr,
		", Header: ", req.Header, ", body: ", req.Body)
	resp := ResData{}
	var upd []UserPointsData
	token := u.GetString("token", "")
	tokenBool := CheckLoginUser(token)

	if !tokenBool {
		resp.Code = 403
		resp.Mesg = "Request parameter error"
		u.RetData(resp)
		return
	}

	ownerRepo := u.GetString("sort", "mindspore")

	result, num, err := models.QuerySortPointsCount(ownerRepo)
	if err != nil {
		logs.Info(err)
		resp.Code = 404
		resp.Mesg = "Data does not exist"
		resp.TotalCount = 0
		logs.Error("Data does not exist")
		u.RetData(resp)
		return
	}

	for _, pd := range result {

		var up UserPointsData
		up.UserId = pd.UserId
		up.GitUserId = pd.GiteeId
		up.IntegralValue = pd.IntegralValue
		upd = append(upd, up)
	}

	resp.Code = 200
	resp.Mesg = "Data query is successful"
	resp.TotalCount = num

	resp.UserPoints = upd
	u.RetData(resp)
	return

}
