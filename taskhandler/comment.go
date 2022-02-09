package taskhandler

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"osi-task-manager/util"
	"strings"
	"sync"
)

var pointLock sync.Mutex
var gaussLock sync.Mutex
var assignLock sync.Mutex
var unassignLock sync.Mutex
var completeLock sync.Mutex
var approveLock sync.Mutex

type EulerIssueUserRecordTp struct {
	UserId      int64
	OrId        int64
	IssueNumber string
	RepoPath    string
	Owner       string
	Status      int8
}

func CreateIssueBody(eulerToken, path, statusName string, eoi models.EulerOriginIssue) string {
	requestBody := ""
	body := eoi.IssueBody
	body = strings.ReplaceAll(body, "`", "\\`")
	body = strings.ReplaceAll(body, `"`, "\"")
	requestBody = fmt.Sprintf(`{
			"access_token": "%s",
			"repo": "%s", 
			"title": "%s",
			"state": "%s",
			"body": "%s",
			"assignee": "%s",
			"labels": "%s",
			"security_hole": "false"
			}`, eulerToken, path, eoi.Title, statusName, body, eoi.IssueAssignee, eoi.IssueLabel)
	return requestBody
}

func UpdateIssueToGit(eulerToken, owner, path, issueState string, eoi models.EulerOriginIssue) error {
	if eulerToken != "" && owner != "" && path != "" {
		url := "https://gitee.com/api/v5/repos/" + owner + "/issues/" + eoi.IssueNumber
		statusName := IssueStateRev(issueState)
		requestBody := CreateIssueBody(eulerToken, path, statusName, eoi)
		logs.Info("UpdateIssueToGit, isssue_body: ", requestBody)
		if requestBody != "" && len(requestBody) > 1 {
			resp, err := util.HTTPPatch(url, requestBody)
			if err != nil {
				logs.Error("UpdateIssueToGit, Update issue failed, issueNum: ", eoi.IssueNumber, "err: ", err)
				return errors.New("Failed to call gitee to update the issue interface")
			}
			if _, ok := resp["id"]; !ok {
				logs.Error("UpdateIssueToGit, Failed to create issue, err: ", ok)
				return errors.New("Failed to call gitee to update the issue interface")
			}
			// Update the status of issue in db
			if eoi.IssueState != statusName {
				eoi.IssueState = statusName
				upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState")
				if upIssueErr != nil {
					logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
				}
			}
		}
	}
	return nil
}

// Entry function for handling issue status
func HandleIssueStateChange(issueHook *models.IssuePayload) error {
	eulerToken := ""
	issueId := issueHook.Issue.Id
	issueTitle := common.TrimString(issueHook.Issue.Title)
	issueType := common.TrimString(issueHook.Issue.TypeName)
	issueNumber := common.TrimString(issueHook.Issue.Number)
	repoPath := common.TrimString(issueHook.Repository.Path)
	owner := common.TrimString(issueHook.Repository.NameSpace)
	if issueType == CIssueType || strings.HasPrefix(issueTitle, CIssueType) {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueId, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId == 0 {
			logs.Error("QueryEulerOriginIssue, Data does not exist, eiErr: ", eiErr)
			return errors.New("no data")
		}
		eulerToken = common.GetEnvToken(owner)
		if len(eulerToken) < 1 {
			logs.Error("namespace error: ", owner)
			return errors.New("no token")
		}
		switch issueHook.State {
		case IssueOpenState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.GitId != issueHook.Sender.Id {
				is := fmt.Sprintf(IssueStateProc, issueHook.Sender.UserName)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, eoi.IssueState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			upErr := UpdateIssueToGit(eulerToken, owner, repoPath, IssueOpenState, eoi)
			if upErr != nil {
				logs.Error("UpdateIssueToGit, upErr: ", upErr)
			}
		case IssueProgressState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.GitId != issueHook.Sender.Id {
				is := fmt.Sprintf(IssueStateProc, issueHook.Sender.UserName)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, eoi.IssueState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			upErr := UpdateIssueToGit(eulerToken, owner, repoPath, IssueProgressState, eoi)
			if upErr != nil {
				logs.Error("UpdateIssueToGit, upErr: ", upErr)
			}
		case IssueCloseState:
			is := fmt.Sprintf(IssueUnableCloseNot, issueHook.Sender.Login)
			AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
			upErr := UpdateIssueToGit(eulerToken, owner, repoPath, IssueOpenState, eoi)
			if upErr != nil {
				logs.Error("UpdateIssueToGit, upErr: ", upErr)
			}
			eoi.IssueState = IssueOpenState
			upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
			}
		case IssueRejectState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.GitId != issueHook.Sender.Id {
				is := fmt.Sprintf(IssueStateProc, issueHook.Sender.UserName)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, eoi.IssueState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			eoi.IssueState = IssueRejectState
			upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
			}
			userId, delErr := models.RejectEulerOriginIssueAll(&eoi)
			if delErr != nil {
				logs.Error("DeleteEulerOriginIssueAll, Data deletion failed, delErr: ", delErr)
				return delErr
			}
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: eoi.IssueNumber,
				RepoPath: eoi.RepoPath, Owner: owner, Status: 12}
			EulerIssueUserRecord(et)
		}
	}
	return nil
}

func EulerAccountFreeze(issueNumber, repo, owner, eulerToken string, eoi models.EulerOriginIssue) error {
	eu := models.EulerIssueUser{OrId: eoi.OrId}
	euErr := models.QueryEulerIssueUser(&eu, "OrId")
	if euErr == nil {
		su := models.StdUser{UserId: eu.UserId, Status: 1}
		qeuErr := models.QueryStdUser(&su, "UserId", "Status")
		if qeuErr == nil {
			vpErr := VerifyEulerUserClaimPerm(su.GitId, su.UserId, su.GitUserId, issueNumber,
				repo, eulerToken, owner, eoi)
			if vpErr != nil {
				upErr := UpdateIssueToGit(eulerToken, owner, repo, IssueOpenState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
					return upErr
				}
				is := fmt.Sprintf(IssueUncompleteClaimCount, su.GitUserId)
				AddCommentToIssue(is, issueNumber, owner, repo, eulerToken)
				logs.Error("vpErr: ", vpErr)
				return vpErr
			}
		}
	}
	return nil
}

// Parse issue comments
func HandleIssueComment(payload models.CommentPayload) {
	if payload.Issue == nil || payload.Comment == nil {
		return
	}
	if payload.Comment.User == nil {
		return
	}
	// The default timeout for receiving hooks
	logs.Info("payload.Comment: ", payload.Comment, ", Number: ", payload.Issue.Number, "id: ", payload.Issue.Id)
	issueNum := payload.Issue.Number           //issue number string
	issueId := payload.Issue.Id                // issue id int64
	cBody := payload.Comment.Body              //Comment subject
	cuAccount := payload.Comment.User.UserName //gitee domain address
	osiAssignedCmd := beego.AppConfig.DefaultString("osiAssignedCmd", "/intern-assign")
	osiTutApproveCmd := beego.AppConfig.DefaultString("osiTutApproveCmd", "/intern-apprrove")
	osiTutUnapproveCmd := beego.AppConfig.DefaultString("osiTutUnapproveCmd", "/intern-unapprove")
	osiStdCompletedCmd := beego.AppConfig.DefaultString("osiStdCompletedCmd", "/intern-completed")
	osiTutDoneCmd := beego.AppConfig.DefaultString("osiTutDoneCmd", "/intern-done")
	osiUnassignCmd := beego.AppConfig.DefaultString("osiUnassignCmd", "/intern-unassign")
	osiFailCmd := beego.AppConfig.DefaultString("osiFailCmd", "/intern-fail")
	closeIssueCmd := beego.AppConfig.DefaultString("close_issue", "/close")
	if issueNum == "" || cuAccount == "" || cBody == "" {
		logs.Error("Data has null values: issueNum, cuAccount, cBody: ", issueNum, cuAccount, cBody)
		return
	}
	if payload.Issue.State == "closed" || payload.Issue.State == "rejected" ||
		payload.Issue.State == "已完成" || payload.Issue.State == "已拒绝" {
		logs.Error("Cannot edit comment, value: ", payload.Issue)
		return
	}
	issueTitle := common.TrimString(payload.Issue.Title)
	issueType := common.TrimString(payload.Issue.TypeName)
	issueNumber := common.TrimString(payload.Issue.Number)
	repoPath := common.TrimString(payload.Repository.Path)
	owner := common.TrimString(payload.Repository.NameSpace)
	if issueType == CIssueType || strings.HasPrefix(issueTitle, CIssueType) {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueId, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId == 0 {
			logs.Error("QueryEulerOriginIssue, Data does not exist, eiErr: ", eiErr)
			return
		}
		eulerToken := common.GetEnvToken(payload.Repository.NameSpace)
		if len(eulerToken) < 1 {
			logs.Error("namespace error: ", payload.Repository.NameSpace)
			return
		}
		if strings.HasPrefix(cBody, osiAssignedCmd) {
			// first-claimed task
			assignLock.Lock()
			UserClaimTask(payload, eulerToken, owner, eoi)
			assignLock.Unlock()
		} else if strings.HasPrefix(cBody, osiTutApproveCmd) {
			approveLock.Lock()
			TutApproveTask(payload, eulerToken, owner, eoi)
			approveLock.Unlock()
		} else if strings.HasPrefix(cBody, osiTutUnapproveCmd) {
			approveLock.Lock()
			TutUnApproveTask(payload, eulerToken, owner, eoi)
			approveLock.Unlock()
		} else if strings.HasPrefix(cBody, osiStdCompletedCmd) {
			// User submits task
			completeLock.Lock()
			UserSubmitsTask(payload, eulerToken, owner, eoi)
			completeLock.Unlock()
		} else if strings.HasPrefix(cBody, osiUnassignCmd) {
			// Give up the task
			unassignLock.Lock()
			UserGiveUpTask(payload, eulerToken, owner, eoi)
			unassignLock.Unlock()
		} else if strings.HasPrefix(cBody, osiFailCmd) {
			// Give up the task
			unassignLock.Lock()
			TutGiveUpTask(payload, eulerToken, owner, eoi)
			unassignLock.Unlock()
		} else if strings.HasPrefix(cBody, closeIssueCmd) {
			// close cmd
			AssignCloseIssue(payload, eulerToken, owner, eoi)
		} else if strings.HasPrefix(cBody, osiTutDoneCmd) {
			// done cmd
			AssignDoneIssue(payload, eulerToken, owner, eoi)
		}
	}
}

func AssignDoneIssue(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	if eoi.IssueState == "closed" {
		// The issue has been closed and cannot be operated again
		logs.Error("AssignDoneIssue， The issue has been closed and cannot be operated again,issuetmp: ", eoi)
		return
	}
	tu := models.TutUser{GitId: payload.Comment.User.Id, Status: 1}
	tuErr := models.QueryTutUser(&tu, "GitId", "Status")
	// Non-reviewers, cannot modify the status of the issue
	if tu.UserId > 0 || eoi.GitId == payload.Comment.User.Id {
		eafErr := EulerAccountFreeze(payload.Issue.Number, payload.Repository.Path, owner, eulerToken, eoi)
		if eafErr != nil {
			logs.Error("AssignDoneIssue， eafErr: ", eafErr)
			return
		}
		upErr := UpdateIssueToGit(eulerToken, owner, eoi.RepoPath, IssueCloseState, eoi)
		if upErr != nil {
			logs.Error("AssignDoneIssue, upErr: ", upErr)
			return
		}
		// Calculate the points earned by users
		pointLock.Lock()
		CalculateUserPoints(eulerToken, eoi)
		pointLock.Unlock()
		// Modify data status
		eoi.IssueState = IssueCloseState
		eoi.UpdateTime = common.GetCurTime()
		upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState", "UpdateTime")
		if upIssueErr != nil {
			logs.Error("AssignDoneIssue, upIssueErr: ", upIssueErr)
		}
		return
	}
	logs.Error("AssignDoneIssue, tuErr: ", tuErr)
}

func AssignCloseIssue(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	if eoi.IssueState == "closed" {
		// The issue has been closed and cannot be operated again
		logs.Error("AssignCloseIssue， The issue has been closed and cannot be operated again,issuetmp: ", eoi)
		return
	}
	tu := models.TutUser{GitId: payload.Comment.User.Id, Status: 1}
	tuErr := models.QueryTutUser(&tu, "GitId", "Status")
	// Non-reviewers, cannot modify the status of the issue
	if tu.UserId > 0 || eoi.GitId == payload.Comment.User.Id {
		eafErr := EulerAccountFreeze(payload.Issue.Number, payload.Repository.Path, owner, eulerToken, eoi)
		if eafErr != nil {
			logs.Error("eafErr: ", eafErr)
			return
		}
		upErr := UpdateIssueToGit(eulerToken, owner, eoi.RepoPath, IssueCloseState, eoi)
		if upErr != nil {
			logs.Error("UpdateIssueToGit, upErr: ", upErr)
			return
		}
		// Modify data status
		eoi.IssueState = IssueCloseState
		eoi.UpdateTime = common.GetCurTime()
		upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState", "UpdateTime")
		if upIssueErr != nil {
			logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
		}
		return
	}
	logs.Error("AssignCloseIssue, tuErr: ", tuErr)
}

func AddEulerUserUnassignCount(userId, gitId int64, userName string, relFlag int8) {
	eu := models.EulerUnassignUser{UserId: userId, GitId: gitId}
	eus := models.QueryEulerUserUnassigned(&eu)
	if len(eus) > 0 {
		actReleaseFlag := false
		passReleaseFlag := false
		for _, es := range eus {
			if es.ReleaseFlag == 1 && relFlag == 1 {
				actReleaseFlag = true
				es.CountValue += 1
				es.UpdateTime = common.GetCurTime()
				activeReleasesCount := beego.AppConfig.DefaultInt("claimed::active_releases_count", 3)
				if es.CountValue > int8(activeReleasesCount) {
					afterDate, beErr := beego.AppConfig.Int("claimed::active_releases_date")
					if beErr != nil {
						afterDate = 30
					}
					es.UnassignTime = common.GetAfterTime(afterDate)
					upErr := models.UpdateEulerUserUnassigned(&es, "CountValue", "UpdateTime", "UnassignTime")
					logs.Info("upErr: ", upErr)
				} else {
					upErr := models.UpdateEulerUserUnassigned(&es, "CountValue", "UpdateTime")
					logs.Info("upErr: ", upErr)
				}
				break
			}
			if es.ReleaseFlag == 2 && relFlag == 2 {
				passReleaseFlag = true
				es.CountValue += 1
				es.UpdateTime = common.GetCurTime()
				passiveReleasesCount := beego.AppConfig.DefaultInt("claimed::passive_release_count", 3)
				if int(es.CountValue) > passiveReleasesCount {
					afterDate, beErr := beego.AppConfig.Int("claimed::passive_release_date")
					if beErr != nil {
						afterDate = 90
					}
					es.UnassignTime = common.GetAfterTime(afterDate)
					upErr := models.UpdateEulerUserUnassigned(&es, "CountValue", "UpdateTime", "UnassignTime")
					logs.Info("upErr: ", upErr)
				} else {
					upErr := models.UpdateEulerUserUnassigned(&es, "CountValue", "UpdateTime")
					logs.Info("upErr: ", upErr)
				}
				break
			}
		}
		if !actReleaseFlag && relFlag == 1 {
			eu := models.EulerUnassignUser{UserId: userId, GitId: gitId, GitUserId: userName,
				CountValue: 1, CreateTime: common.GetCurTime(), ReleaseFlag: 1}
			num, inErr := models.InsertEulerUserUnassigned(&eu)
			logs.Info("num: ", num, ",inErr: ", inErr)
		}
		if !passReleaseFlag && relFlag == 2 {
			eu := models.EulerUnassignUser{UserId: userId, GitId: gitId, GitUserId: userName,
				CountValue: 1, CreateTime: common.GetCurTime(), ReleaseFlag: 2}
			num, inErr := models.InsertEulerUserUnassigned(&eu)
			logs.Info("num: ", num, ",inErr: ", inErr)
		}
	} else {
		if relFlag == 1 {
			eu := models.EulerUnassignUser{UserId: userId, GitId: gitId, GitUserId: userName,
				CountValue: 1, CreateTime: common.GetCurTime(), ReleaseFlag: 1}
			num, inErr := models.InsertEulerUserUnassigned(&eu)
			logs.Info("num: ", num, ",inErr: ", inErr)
		}
		if relFlag == 2 {
			eu := models.EulerUnassignUser{UserId: userId, GitId: gitId, GitUserId: userName,
				CountValue: 1, CreateTime: common.GetCurTime(), ReleaseFlag: 2}
			num, inErr := models.InsertEulerUserUnassigned(&eu)
			logs.Info("num: ", num, ",inErr: ", inErr)
		}
	}
}

func UserClaimTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	su := models.StdUser{GitId: payload.Comment.User.Id, Status: 1}
	stdErr := models.QueryStdUser(&su, "GitId", "Status")
	if su.UserId == 0 {
		logs.Error("UserClaimTask, The user has not signed up or signed up successfully, "+
			"unable to receive the task, stdErr: ", stdErr)
		is := fmt.Sprintf(IssueClaiNotSigned, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueClaiNotSignedSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		return
	} else {
		UpdateStdUsers(&su, payload.Comment.User.UserName, payload.Comment.User.UserName)
		if su.TestFlag == 1 {
			logs.Error("UserClaimTask, The user did not complete the test task, GitUserId: ", su.GitUserId)
			testRepo := beego.AppConfig.String("repo::test_pr")
			is := fmt.Sprintf(IssueTestPrCommit, payload.Comment.User.UserName, testRepo)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			iss := fmt.Sprintf(IssueTestPrCommitSend, eoi.GitUrl, testRepo)
			SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
			return
		}
		VerifyStdClaimReq(payload, su.UserId, eulerToken, owner, eoi)
	}
}

func VerifyEulerUserClaimPerm(gitId, userId int64, userName, issueNumber, repo, eulerToken, owner string,
	eoi models.EulerOriginIssue) error {
	// Verify that it is in the blacklist
	ebu := models.EulerBlackUser{GitId: gitId, UserId: userId}
	_ = models.QueryEulerBlackUser(&ebu, "GitId", "UserId")
	if ebu.Id > 0 {
		cc := fmt.Sprintf(IssueBlackClaimFailure, userName)
		AddCommentToIssue(cc, issueNumber, owner, repo, eulerToken)
		et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: issueNumber,
			RepoPath: repo, Owner: owner, Status: 13}
		EulerIssueUserRecord(et)
		return errors.New("Blacklisted users cannot claim tasks")
	}
	// Whether the user release task exceeds the limit
	euu := models.EulerUnassignUser{GitId: gitId, UserId: userId}
	eus := models.QueryEulerUserUnassigned(&euu)
	if len(eus) > 0 {
		activeReleasesCount := beego.AppConfig.DefaultInt("claimed::active_releases_count", 6)
		passiveReleaseCount := beego.AppConfig.DefaultInt("claimed::passive_release_count", 3)
		for _, es := range eus {
			if es.ReleaseFlag == 1 && int(es.CountValue) > activeReleasesCount {
				cc := fmt.Sprintf(IssueUnassignClaimCount, userName)
				AddCommentToIssue(cc, issueNumber, owner, repo, eulerToken)
				et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: issueNumber,
					RepoPath: repo, Owner: owner, Status: 14}
				EulerIssueUserRecord(et)
				return errors.New("The number of canceled tasks has reached the line, " +
					"and the task cannot be claimed again")
			}
			if es.ReleaseFlag == 2 && int(es.CountValue) > passiveReleaseCount {
				cc := fmt.Sprintf(IssueUnassignedClaimCount, userName)
				AddCommentToIssue(cc, issueNumber, owner, repo, eulerToken)
				et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: issueNumber,
					RepoPath: repo, Owner: owner, Status: 19}
				EulerIssueUserRecord(et)
				return errors.New("The number of canceled tasks has reached the line, " +
					"and the task cannot be claimed again")
			}
		}
	}
	return nil
}

func VerifyStdClaimReq(payload models.CommentPayload, userId int64, eulerToken, owner string, eoi models.EulerOriginIssue) {
	vpErr := VerifyEulerUserClaimPerm(payload.Comment.User.Id, userId, payload.Comment.User.UserName, payload.Issue.Number,
		payload.Repository.Path, eulerToken, owner, eoi)
	if vpErr != nil {
		logs.Error("vpErr: ", vpErr)
		return
	}
	// Determine whether the user denies the task
	eucr := models.EulerUserClaimRecord{UserId: userId, OrId: eoi.OrId}
	eucErr := models.QueryEulerUserClaimRecord(&eucr, "UserId", "OrId")
	if eucr.Id > 0 {
		logs.Info("Repeat claim task, eucErr: ", eucErr, eucr)
		cc := fmt.Sprintf(IssueClaimSameTask, payload.Comment.User.UserName)
		AddCommentToIssue(cc, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueClaimSameSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 6}
		EulerIssueUserRecord(et)
		return
	}
	issueCount := beego.AppConfig.DefaultInt("claimed::issue_count", 1)
	// Verify whether it is the first-claimed task
	eiu := models.QueryEulerIssueUnfinished(userId, 3)
	ciaimCount := len(eiu)
	if ciaimCount >= issueCount {
		cc := fmt.Sprintf(IssueClaimFailure, payload.Comment.User.UserName)
		AddCommentToIssue(cc, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		SendPrivateLetters(eulerToken, IssueClaimFailureSend, payload.Comment.User.UserName)
		et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 5}
		EulerIssueUserRecord(et)
		return
	} else {
		if ciaimCount > 0 {
			for _, e := range eiu {
				if e.OrId == eoi.OrId {
					ic := fmt.Sprintf(IssueClaimSameTask, payload.Comment.User.UserName)
					AddCommentToIssue(ic, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
					et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
						RepoPath: payload.Repository.Path, Owner: owner, Status: 6}
					EulerIssueUserRecord(et)
					return
				}
			}
		}
		// Claim the task successfully
		StartClaimTask(payload, userId, eulerToken, owner, eoi)
	}
}

func StartClaimTask(payload models.CommentPayload, userId int64, eulerToken, owner string, eoi models.EulerOriginIssue) {
	eu := models.EulerIssueUser{OrId: eoi.OrId}
	euErr := models.QueryEulerIssueUser(&eu, "OrId")
	if eu.Id == 0 || euErr != nil {
		untreatedAfterDate, beErr := beego.AppConfig.Int("claimed::issue_Untreated_date")
		if beErr != nil {
			untreatedAfterDate = 14
		}
		issueUntreatedDate := common.GetAfterTime(untreatedAfterDate)
		eu = models.EulerIssueUser{OrId: eoi.OrId, UserId: userId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, SendEmail: 1, Status: 1,
			CreateTime: common.GetCurTime(), FreezeTime: issueUntreatedDate, PointFlag: 1}
		id, inErr := models.InsertEulerIssueUser(&eu)
		if id > 0 && inErr == nil {
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 1}
			EulerIssueUserRecord(et)
			eucr := models.EulerUserClaimRecord{UserId: userId, OrId: eoi.OrId, CreateTime: common.GetCurTime()}
			models.InsertEulerUserClaimRecord(&eucr)
			is := fmt.Sprintf(IssueClaimSuccess, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			iss := fmt.Sprintf(IssueClaimSuccessSend, eoi.GitUrl)
			SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
			// Edit label
			osiTaskAssign := beego.AppConfig.String("osi_task_assign")
			EditLabel(payload.Repository.Path, payload.Issue.Number, osiTaskAssign, "", eulerToken, owner, eoi)
		} else {
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 7}
			EulerIssueUserRecord(et)
			is := fmt.Sprintf(IssueClaimPree, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			iss := fmt.Sprintf(IssueClaimPreeSend, eoi.GitUrl)
			SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		}
	} else {
		et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 7}
		EulerIssueUserRecord(et)
		is := fmt.Sprintf(IssueClaimPree, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueClaimPreeSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
	}
}

func EulerIssueUserRecord(et EulerIssueUserRecordTp) {
	eir := models.EulerIssueUserRecord{UserId: et.UserId, OrId: et.OrId, IssueNumber: et.IssueNumber,
		RepoPath: et.RepoPath, Owner: et.Owner, Status: et.Status, CreateTime: common.GetCurTime()}
	models.InsertEulerIssueUserRecord(&eir)
}

// UserSubmitsTask User submits task
func UserSubmitsTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	su := models.StdUser{GitId: payload.Comment.User.Id, Status: 1}
	stdErr := models.QueryStdUser(&su, "GitId", "Status")
	if su.UserId == 0 {
		logs.Error("UserClaimTask, The user has not signed up or signed up successfully, "+
			"unable to receive the task, stdErr: ", stdErr)
		is := fmt.Sprintf(IssueClaimCheckSubmit, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueClaimCheckSubmitSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		return
	} else {
		UpdateStdUsers(&su, payload.Comment.User.UserName, payload.Comment.User.UserName)
		if su.TestFlag == 1 {
			logs.Error("UserClaimTask, The user did not complete the test task, GitUserId: ", su.GitUserId)
			testRepo := beego.AppConfig.String("repo::test_pr")
			is := fmt.Sprintf(IssueClaimTestPrSubmit, payload.Comment.User.UserName, testRepo)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			iss := fmt.Sprintf(IssueClaimTestPrSubmitSend, eoi.GitUrl, testRepo)
			SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
			return
		}
		vpErr := VerifyEulerUserClaimPerm(payload.Comment.User.Id, su.UserId, payload.Comment.User.UserName, payload.Issue.Number,
			payload.Repository.Path, eulerToken, owner, eoi)
		if vpErr != nil {
			logs.Error("vpErr: ", vpErr)
			return
		}
		// Determine whether the submitted task and the claimed task are the same user
		eu := models.EulerIssueUser{OrId: eoi.OrId, UserId: su.UserId}
		euErr := models.QueryEulerIssueUser(&eu, "OrId", "UserId")
		if eu.Id == 0 {
			logs.Error("No user claim information is queried,euErr: ", euErr)
			is := fmt.Sprintf(IssueClaimSubmit, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 8}
			EulerIssueUserRecord(et)
		} else {
			if eu.Status == 4 {
				// Edit label
				osiTaskAssigned := beego.AppConfig.String("osi_task_assigned")
				osiTaskRewiew := beego.AppConfig.String("osi_task_rewiew")
				EditLabel(payload.Repository.Path, payload.Issue.Number, osiTaskRewiew, osiTaskAssigned, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueClaimSubmitComplete, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 9}
				EulerIssueUserRecord(et)
			} else if eu.Status == 3 {
				// Edit label
				osiTaskAssigned := beego.AppConfig.String("osi_task_assigned")
				osiTaskRewiew := beego.AppConfig.String("osi_task_rewiew")
				EditLabel(payload.Repository.Path, payload.Issue.Number, osiTaskRewiew, osiTaskAssigned, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueClaimComplete, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 9}
				EulerIssueUserRecord(et)
				iss := fmt.Sprintf(IssueClaimCompleteSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
				tu := models.TutUser{UserId: eu.TutId, Status: 1}
				TutErr := models.QueryTutUser(&tu, "UserId", "Status")
				if tu.UserId > 0 {
					iss := fmt.Sprintf(IssueReviewTaskSend, eoi.GitUrl)
					SendPrivateLetters(eulerToken, iss, tu.UserName)
					return
				} else {
					logs.Error("Failed to query instructor information, TuErr: ", TutErr, eu)
				}
			} else if eu.Status == 1 {
				is := fmt.Sprintf(IssueStdUnapproveComplete, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				iss := fmt.Sprintf(IssueStdUnapproveCompleteSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
				et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 18}
				EulerIssueUserRecord(et)
			} else if eu.Status == 2 {
				eu.Status = 3
				eu.UpdateTime = common.GetCurTime()
				upErr := models.UpdateEulerIssueUser(&eu, "Status", "UpdateTime")
				logs.Info("upErr: ", upErr)
				// Edit label
				osiTaskAssigned := beego.AppConfig.String("osi_task_assigned")
				osiTaskRewiew := beego.AppConfig.String("osi_task_rewiew")
				EditLabel(payload.Repository.Path, payload.Issue.Number, osiTaskRewiew, osiTaskAssigned, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueClaimComplete, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				iss := fmt.Sprintf(IssueClaimCompleteSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
				tu := models.TutUser{UserId: eu.TutId, Status: 1}
				TutErr := models.QueryTutUser(&tu, "UserId", "Status")
				if tu.UserId > 0 {
					iss := fmt.Sprintf(IssueReviewTaskSend, eoi.GitUrl)
					SendPrivateLetters(eulerToken, iss, tu.UserName)
					return
				} else {
					logs.Error("Failed to query instructor information, TuErr: ", TutErr, eu)
				}
				et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 3}
				EulerIssueUserRecord(et)
			}
		}
	}
}

// Give up the task
func UserGiveUpTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	su := models.StdUser{GitId: payload.Comment.User.Id, Status: 1}
	stdErr := models.QueryStdUser(&su, "GitId", "Status")
	if su.UserId == 0 {
		logs.Error("UserClaimTask, The user has not signed up or signed up successfully, "+
			"unable to receive the task, stdErr: ", stdErr)
		is := fmt.Sprintf(IssueGiveUpTask, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueGiveUpTaskSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		return
	}
	UpdateStdUsers(&su, payload.Comment.User.UserName, payload.Comment.User.UserName)
	vpErr := VerifyEulerUserClaimPerm(payload.Comment.User.Id, su.UserId,
		payload.Comment.User.UserName, payload.Issue.Number,
		payload.Repository.Path, eulerToken, owner, eoi)
	if vpErr != nil {
		logs.Error("vpErr: ", vpErr)
		return
	}
	// Determine whether the user denies the task
	eu := models.EulerIssueUser{OrId: eoi.OrId, UserId: su.UserId}
	euErr := models.QueryEulerIssueUser(&eu, "OrId", "UserId")
	if eu.Id == 0 {
		logs.Error("Failed to give up the task, have not received this task,euErr: ", euErr)
		is := fmt.Sprintf(IssueGiveUpTask, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 10}
		EulerIssueUserRecord(et)
	} else {
		if eu.Status == 4 && eoi.IssueState == "closed" {
			is := fmt.Sprintf(IssueGiveUpFailure, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 11}
			EulerIssueUserRecord(et)
		} else if eu.Status <= 3 && eoi.IssueState != "closed" {
			// give up task
			delErr := models.DeleteEulerIssueUser(&eu, "UserId", "OrId")
			if delErr == nil {
				// Edit label
				RelReduceLabel(payload.Repository.Path, payload.Issue.Number, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueGiveUpSuccess, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				AddEulerUserUnassignCount(su.UserId, payload.Comment.User.Id, payload.Comment.User.UserName, 1)
				eir := models.QueryEulerIssueUserRecordset(su.UserId, eoi.OrId, 2)
				if len(eir) < 1 {
					iss := fmt.Sprintf(IssueGiveUpSuccessSend, eoi.GitUrl)
					SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
				}
				et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 2}
				EulerIssueUserRecord(et)
			}
		}
	}
}

func TutGiveUpTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	tu := models.TutUser{GitId: payload.Comment.User.Id, Status: 1}
	tutErr := models.QueryTutUser(&tu, "GitId", "Status")
	if tu.UserId == 0 {
		logs.Error("TutApproveTask, Only the instructor who has passed "+
			"the certification can review the task, stdErr: ", tutErr)
		is := fmt.Sprintf(IssueTutFailTask, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueTutAssignTaskSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		return
	}
	UpdateTutUsers(&tu, payload.Comment.User.UserName, payload.Comment.User.UserName)
	// Determine whether the user denies the task
	eu := models.EulerIssueUser{OrId: eoi.OrId}
	euErr := models.QueryEulerIssueUser(&eu, "OrId")
	if eu.Id == 0 {
		logs.Error("Failed to give up the task, have not received this task,euErr: ", euErr)
		is := fmt.Sprintf(IssueTutUnassignTask, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		et := EulerIssueUserRecordTp{UserId: eu.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 10}
		EulerIssueUserRecord(et)
	} else {
		if eu.Status == 4 && eoi.IssueState == "closed" {
			is := fmt.Sprintf(IssueGiveUpFailure, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			et := EulerIssueUserRecordTp{UserId: eu.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 11}
			EulerIssueUserRecord(et)
		} else if eu.Status <= 3 && eoi.IssueState != "closed" {
			// give up task
			delErr := models.DeleteEulerIssueUser(&eu, "UserId", "OrId")
			if delErr == nil {
				// Edit label
				RelReduceLabel(payload.Repository.Path, payload.Issue.Number, eulerToken, owner, eoi)
				su := models.StdUser{UserId: eu.UserId}
				stdErr := models.QueryStdUser(&su, "UserId")
				if su.UserId == 0 {
					logs.Error("UserClaimTask, The user has not signed up or signed up successfully, "+
						"unable to receive the task, stdErr: ", stdErr)
				}
				is := fmt.Sprintf(IssueTutRelSuccessTask, payload.Comment.User.UserName, su.GitUserId)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				stdIs := fmt.Sprintf(IssueStdRelTaskSubmit, su.GitUserId)
				AddCommentToIssue(stdIs, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				AddEulerUserUnassignCount(su.UserId, payload.Comment.User.Id, payload.Comment.User.UserName, 2)
				eir := models.QueryEulerIssueUserRecordset(su.UserId, eoi.OrId, 20)
				if len(eir) < 1 {
					iss := fmt.Sprintf(IssueStdRelTaskSubmitSend, eoi.GitUrl, payload.Comment.User.UserName)
					SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
				}
				et := EulerIssueUserRecordTp{UserId: su.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 20}
				EulerIssueUserRecord(et)
			}
		}
	}
}

// EditLabel Edit label
func EditLabel(issuePath, issueNumber, osiTaskLabel, repLabel, eulerToken, owner string, eoi models.EulerOriginIssue) {
	labels := ""
	labelList, _ := QueryIssueLabels(eulerToken, issuePath, issueNumber, owner)
	if len(osiTaskLabel) > 0 {
		if len(labelList) > 0 {
			tmpLabel := make([]string, 0)
			isExist := false
			for _, lab := range labelList {
				if len(lab) > 1 {
					if len(repLabel) > 1 && lab == repLabel {
						logs.Info("delete labels, repLabel: ", repLabel, ", add labels: ", osiTaskLabel)
						continue
					}
					tmpLabel = append(tmpLabel, lab)
					if lab == osiTaskLabel {
						isExist = true
					}
				}
			}
			if !isExist {
				tmpLabel = append(tmpLabel, osiTaskLabel)
			}
			if len(tmpLabel) > 0 {
				labels = strings.Join(tmpLabel, ",")
			}
		} else {
			labels = osiTaskLabel
		}
	}
	if len(labels) > 0 {
		UpdateIssueLabels(eulerToken, issuePath, issueNumber, owner, labels)
		eoi.IssueLabel = labels
		eoi.UpdateTime = common.GetCurTime()
		upErr := models.UpdateEulerOriginIssue(&eoi, "IssueLabel", "UpdateTime")
		if upErr != nil {
			logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
		}
	}
}

func ReduceLabel(issuePath, issueNumber, eulerToken, owner string, eoi models.EulerOriginIssue, reduceLabList []string) {
	labels := ""
	_, totalLabelList := QueryIssueLabels(eulerToken, issuePath, issueNumber, owner)
	if len(totalLabelList) > 0 {
		tmpLabel := make([]string, 0)
		for _, lab := range totalLabelList {
			if len(lab) > 1 {
				isExist := false
				if len(reduceLabList) > 0 {
					for _, rl := range reduceLabList {
						if lab == rl {
							isExist = true
							break
						}
					}
				}
				if !isExist {
					tmpLabel = append(tmpLabel, lab)
				}
			}
		}
		if len(tmpLabel) > 0 {
			labels = strings.Join(tmpLabel, ",")
		}
	}
	if len(labels) > 0 {
		UpdateIssueLabels(eulerToken, issuePath, issueNumber, owner, labels)
		eoi.IssueLabel = labels
		eoi.UpdateTime = common.GetCurTime()
		upErr := models.UpdateEulerOriginIssue(&eoi, "IssueLabel", "UpdateTime")
		if upErr != nil {
			logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
		}
	}
}

func TutApproveTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	tu := models.TutUser{GitId: payload.Comment.User.Id, Status: 1}
	stdErr := models.QueryTutUser(&tu, "GitId", "Status")
	if tu.UserId == 0 {
		logs.Error("TutApproveTask, Only the instructor who has passed "+
			"the certification can review the task, stdErr: ", stdErr)
		is := fmt.Sprintf(IssueTutAssignTask, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueTutAssignTaskSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		return
	} else {
		UpdateTutUsers(&tu, payload.Comment.User.UserName, payload.Comment.User.UserName)
		eu := models.EulerIssueUser{OrId: eoi.OrId}
		euErr := models.QueryEulerIssueUser(&eu, "OrId")
		if eu.Id == 0 || euErr != nil {
			is := fmt.Sprintf(IssueTutUnassignTask, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			return
		} else {
			su := models.StdUser{UserId: eu.UserId, Status: 1}
			stdErr := models.QueryStdUser(&su, "UserId", "Status")
			if su.UserId == 0 {
				is := fmt.Sprintf(IssueStdRegisterTask, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				logs.Error("stdErr: ", stdErr)
				return
			}
			if eu.Status > 1 {
				is := fmt.Sprintf(IssueStdApproveTask, su.GitUserId)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				iss := fmt.Sprintf(IssueStdApproveTaskSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, su.UserName)
				osiTaskAssign := beego.AppConfig.String("osi_task_assign")
				osiTaskAssigned := beego.AppConfig.String("osi_task_assigned")
				EditLabel(payload.Repository.Path, payload.Issue.Number, osiTaskAssigned, osiTaskAssign, eulerToken, owner, eoi)
				return
			}
			untreatedAfterDate := TaskDueDateByPoints(eoi.TaskScore)
			issueUntreatedDate := common.GetAfterTime(untreatedAfterDate)
			eu.Status = 2
			eu.UpdateTime = common.GetCurTime()
			eu.AssignTime = issueUntreatedDate
			eu.TutId = tu.UserId
			upErr := models.UpdateEulerIssueUser(&eu, "Status", "UpdateTime", "AssignTime", "TutId")
			if upErr == nil {
				et := EulerIssueUserRecordTp{UserId: eu.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 15}
				EulerIssueUserRecord(et)
				is := fmt.Sprintf(IssueStdApproveTask, su.GitUserId)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				iss := fmt.Sprintf(IssueStdApproveTaskSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, su.UserName)
				// Edit label
				osiTaskAssign := beego.AppConfig.String("osi_task_assign")
				osiTaskAssigned := beego.AppConfig.String("osi_task_assigned")
				EditLabel(payload.Repository.Path, payload.Issue.Number, osiTaskAssigned, osiTaskAssign, eulerToken, owner, eoi)
			} else {
				et := EulerIssueUserRecordTp{UserId: eu.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 17}
				EulerIssueUserRecord(et)
				is := fmt.Sprintf(IssueClaimWrong, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				iss := fmt.Sprintf(IssueClaimWrongSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
			}
		}
	}
}

func TutUnApproveTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	tu := models.TutUser{GitId: payload.Comment.User.Id, Status: 1}
	stdErr := models.QueryTutUser(&tu, "GitId", "Status")
	if tu.UserId == 0 {
		logs.Error("TutApproveTask, Only the instructor who has passed "+
			"the certification can review the task, stdErr: ", stdErr)
		is := fmt.Sprintf(IssueTutAssignTask, payload.Comment.User.UserName)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueTutAssignTaskSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
		return
	} else {
		UpdateTutUsers(&tu, payload.Comment.User.UserName, payload.Comment.User.UserName)
		eu := models.EulerIssueUser{OrId: eoi.OrId}
		euErr := models.QueryEulerIssueUser(&eu, "OrId")
		if eu.Id == 0 || euErr != nil {
			is := fmt.Sprintf(IssueTutUnassignTask, payload.Comment.User.UserName)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			return
		} else {
			delErr := models.DeleteEulerIssueUser(&eu, "Id", "OrId", "UserId")
			if delErr == nil {
				et := EulerIssueUserRecordTp{UserId: eu.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 16}
				EulerIssueUserRecord(et)
				su := models.StdUser{UserId: eu.UserId, Status: 1}
				stdErr := models.QueryStdUser(&su, "UserId", "Status")
				if su.UserId > 0 {
					is := fmt.Sprintf(IssueStdUnApproveTask, su.GitUserId)
					AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
					iss := fmt.Sprintf(IssueStdUnApproveTaskSend, eoi.GitUrl)
					SendPrivateLetters(eulerToken, iss, su.UserName)
				} else {
					logs.Error("stdErr: ", stdErr)
				}
				// Edit label
				RelReduceLabel(payload.Repository.Path, payload.Issue.Number, eulerToken, owner, eoi)
			} else {
				et := EulerIssueUserRecordTp{UserId: eu.UserId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 17}
				EulerIssueUserRecord(et)
				is := fmt.Sprintf(IssueClaimWrong, payload.Comment.User.UserName)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				iss := fmt.Sprintf(IssueClaimWrongSend, eoi.GitUrl)
				SendPrivateLetters(eulerToken, iss, payload.Comment.User.UserName)
			}
		}
	}
}

func RelReduceLabel(repoPath, issueNumber, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Edit label
	totalLabel := beego.AppConfig.String("totallabel")
	osiTaskAssignList := strings.Split(totalLabel, ",")
	ReduceLabel(repoPath, issueNumber, eulerToken, owner, eoi, osiTaskAssignList)
}
