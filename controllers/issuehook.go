package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"osi-task-manager/taskhandler"
	"strings"
)

var (
	//GiteeUserAgent gitee hook request flag
	GiteeUserAgent = "git-oschina-hook"
	//XGiteeToken password or sign
	XGiteeToken = "X-Gitee-Token"
	//XGIteeEventType webhook event type
	XGIteeEventType = "X-Gitee-Event"
	//NoteHookType type of comment
	NoteHookType = "Note Hook"
	//PullReqHookType type of pull request
	PullReqHookType = "Merge Request Hook"
	//PushTagHookType type of push or tag
	PushTagHookType = "Tag Push Hook"
	//IssueHookType type of issue
	IssueHookType = "Issue Hook"
)

// Create data into db according to hook callback
//InternHookEventControllers gitee hook callback
type InternHookEventControllers struct {
	beego.Controller
}

//Post handle gitee webhook
// @router / [post]
func (c *InternHookEventControllers) Post() {
	if ok := c.isLegitimateHookEvent(); !ok {
		c.Ctx.ResponseWriter.WriteHeader(406)
		c.Ctx.WriteString("Illegal incident, discarded")
		return
	}
	eventType := c.Ctx.Request.Header.Get(XGIteeEventType)
	c.Ctx.ResponseWriter.WriteHeader(200)
	c.Ctx.WriteString("Event received: " + eventType)
	switch eventType {
	case NoteHookType: //handle comment hook data
		c.handleNoteDate()
	case PullReqHookType:
		c.handlePullReq()
	case IssueHookType:
		c.handleIssue()
	case PushTagHookType:
		c.handlePushTag()
	default:
		logs.Info(eventType)
	}
}

//isLegitimateHookEvent according to gitee doc judge
func (c *InternHookEventControllers) isLegitimateHookEvent() (ok bool) {
	ok = true
	//judge user agent
	uAgent := c.Ctx.Request.Header.Get("User-Agent")
	if uAgent != GiteeUserAgent {
		ok = false
	}
	ctType := c.Ctx.Request.Header.Get("Content-Type")
	if "application/json" != ctType {
		ok = false
	}
	//judge hook password
	xToken := c.Ctx.Request.Header.Get(XGiteeToken)
	//logs.Info(xToken)
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	if xToken != hookPwd {
		logs.Error("hookPwd Err, xToken: ", xToken)
	}
	return
}

func (c *InternHookEventControllers) handleNoteDate() {
	var hookNote models.CommentPayload
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &hookNote)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := hookNote.Comment.User.UserName
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	hookNote.Password = common.TrimString(hookNote.Password)
	hookPwd = common.TrimString(hookPwd)
	if hookNote.Action == "comment" && hookNote.NoteableType == "Issue" && hookNote.Password == hookPwd {
		logs.Info(string(c.Ctx.Input.RequestBody))
		//handle issue comment
		go taskhandler.HandleIssueComment(hookNote)
	}
}

func (c *InternHookEventControllers) handlePullReq() {
	prHook := models.PrPayload{}
	logs.Info(string(c.Ctx.Input.RequestBody))
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &prHook)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := prHook.Sender.UserName
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	prHook.Password = common.TrimString(prHook.Password)
	if prHook.Password != hookPwd {
		logs.Error("Hook callback pwd verification error, hook: ", prHook)
		return
	}
	prHook.Repository.NameSpace = common.TrimString(prHook.Repository.NameSpace)
	owner := beego.AppConfig.String("repo::owner")
	ownerList := strings.Split(owner, ",")
	olFlag := false
	for _, ol := range ownerList {
		if prHook.Repository.NameSpace == ol {
			olFlag = true
			break
		}
	}
	if !olFlag {
		logs.Error("This hook does not belong to the current organization:, "+
			"owner: ", prHook.Repository.NameSpace, owner)
		return
	}
	labelStr := ""
	if prHook.PullRequest.Labels != nil && len(prHook.PullRequest.Labels) > 0 {
		for _, la := range prHook.PullRequest.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
	}
	if prHook.Action == "assign" {
		taskhandler.UpdatePrAssignee(prHook)
	}
	if prHook.Action == "open" {
		taskhandler.AddHookPr(&prHook, 1)
	}
	if prHook.Action == "update" {
		taskhandler.AddHookPr(&prHook, 2)
	}
	if prHook.Action == "merge" {
		taskhandler.UpdateHookPr(&prHook)
	}
	if prHook.Action == "closed" {
		//taskhandler.DelHookGaussIssue(&issueHook)
	}
}

func (c *InternHookEventControllers) handlePushTag() {

}

func (c *InternHookEventControllers) handleIssue() {
	logs.Info(string(c.Ctx.Input.RequestBody))
	issueHook := models.IssuePayload{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &issueHook)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := issueHook.Sender.UserName
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	issueHook.Password = common.TrimString(issueHook.Password)
	if issueHook.Password != hookPwd {
		logs.Error("Hook callback pwd verification error, hook: ", issueHook)
		return
	}
	if issueHook.Action == "assign" {
		//Update the person in charge of the issue template
		eoi := models.EulerOriginIssue{Owner: issueHook.Repository.NameSpace, RepoPath: issueHook.Repository.Path,
			IssueId: issueHook.Issue.Id, IssueNumber: issueHook.Iid}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eiErr != nil {
			logs.Error(eiErr)
			return
		}
		eoi.IssueAssignee = issueHook.Assignee.UserName
		eoi.GitId = issueHook.Assignee.Id
		upErr := models.UpdateEulerOriginIssue(&eoi, "IssueAssignee", "GitId")
		if upErr != nil {
			logs.Error(upErr)
		}
	}
	if issueHook.Action == "state_change" {
		//handle issue state change
		err = taskhandler.HandleIssueStateChange(&issueHook)
		if err != nil {
			logs.Error(err)
			return
		}
	}
	if issueHook.Action == "open" {
		eoi := models.EulerOriginIssue{Owner: issueHook.Repository.NameSpace, RepoPath: issueHook.Repository.Path,
			IssueId: issueHook.Issue.Id, IssueNumber: issueHook.Iid}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId > 0 {
			logs.Error(eiErr, ",eoi: ", eoi)
			return
		}
		taskhandler.AddHookIssue(&issueHook)
	}
	if issueHook.Action == "delete" {
		taskhandler.DelHookIssue(&issueHook)
	}
}

