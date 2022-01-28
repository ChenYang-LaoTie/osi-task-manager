package taskhandler

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"strings"
)

func CheckPrAttribute(nameSpace, repoPath, title string) (flag int) {
	testRepo := beego.AppConfig.String("repo::test_pr")
	owner := beego.AppConfig.String("repo::owner")
	testPr := strings.Split(testRepo, "/")
	prTitle := beego.AppConfig.String("repo::pr_title")
	if testPr[0] == nameSpace && testPr[1] == repoPath && strings.HasPrefix(common.TrimString(title), prTitle) {
		return 1
	} else {
		ownerList := strings.Split(owner, ",")
		for _, ol := range ownerList {
			if ol == nameSpace {
				return 2
			}
		}
	}
	return 3
}

func UpdateTestPrAssignee(prHook models.PrPayload) {
	prNumber := prHook.PullRequest.Number
	repoPath := common.TrimString(prHook.Repository.Path)
	//Update the person in charge of the issue template
	gop := models.TestOriginPr{Owner: prHook.Repository.NameSpace, RepoPath: repoPath,
		PrId: prHook.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryTestOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gopErr != nil {
		logs.Error(gopErr)
		return
	}
	assigneeSlice := []string{}
	if len(prHook.PullRequest.Assignees) > 0 {
		for _, as := range prHook.PullRequest.Assignees {
			assigneeSlice = append(assigneeSlice, as.UserName)
		}
	} else {
		if len(prHook.PullRequest.Assignee) > 0 {
			assigneeSlice = append(assigneeSlice, prHook.PullRequest.Assignee)
		}
	}
	if len(assigneeSlice) > 0 {
		gop.PrAssignee = strings.Join(assigneeSlice, ",")
	}
	upErr := models.UpdateTestOriginPr(&gop, "PrAssignee")
	if upErr != nil {
		logs.Error(upErr)
	}
}

func UpdateEulerPrAssignee(prHook models.PrPayload) {
	prNumber := prHook.PullRequest.Number
	repoPath := common.TrimString(prHook.Repository.Path)
	//Update the person in charge of the issue template
	gop := models.EulerOriginPr{Owner: prHook.Repository.NameSpace, RepoPath: repoPath,
		PrId: prHook.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryEulerOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gopErr != nil {
		logs.Error(gopErr)
		return
	}
	assigneeSlice := []string{}
	if len(prHook.PullRequest.Assignees) > 0 {
		for _, as := range prHook.PullRequest.Assignees {
			assigneeSlice = append(assigneeSlice, as.UserName)
		}
	} else {
		if len(prHook.PullRequest.Assignee) > 0 {
			assigneeSlice = append(assigneeSlice, prHook.PullRequest.Assignee)
		}
	}
	if len(assigneeSlice) > 0 {
		gop.PrAssignee = strings.Join(assigneeSlice, ",")
	}
	upErr := models.UpdateEulerOriginPr(&gop, "PrAssignee")
	if upErr != nil {
		logs.Error(upErr)
	}
}

// UpdatePrAssignee Update pr responsible person
func UpdatePrAssignee(prHook models.PrPayload) {
	repoPath := common.TrimString(prHook.Repository.Path)
	flag := CheckPrAttribute(prHook.Repository.NameSpace, repoPath, prHook.Title)
	if flag == 3 {
		logs.Error("prHook: ", prHook)
		return
	} else if flag == 1 {
		UpdateTestPrAssignee(prHook)
	} else {
		UpdateEulerPrAssignee(prHook)
	}
}

func AddHookPr(prData *models.PrPayload, openFlag int) {
	repoPath := common.TrimString(prData.Repository.Path)
	flag := CheckPrAttribute(prData.Repository.NameSpace, repoPath, prData.Title)
	logs.Info("CheckPrAttribute, flag: ", flag)
	if flag == 3 {
		logs.Error("prHook: ", prData)
		return
	} else if flag == 1 {
		AddHookTestPr(prData, openFlag)
	} else {
		AddHookEulerPr(prData, openFlag)
	}
}

// AddPr Add pr
func AddHookTestPr(prData *models.PrPayload, openFlag int) {
	gitId := prData.PullRequest.User.Id
	prCreater := prData.PullRequest.User.UserName
	su := models.StdUser{GitId: gitId, Status: 1}
	suErr := models.QueryStdUser(&su, "GitId", "Status")
	if suErr != nil {
		logs.Error("suErr: ", suErr)
	}
	if su.UserId == 0 || su.Status != 1 || su.TestFlag == 2 {
		logs.Error("The user does not exist, or the test task has been completed,"+
			" GitUserId: ", prCreater, gitId)
		return
	}
	UpdateStdUsers(&su, prData.PullRequest.User.UserName, prData.PullRequest.User.UserName)
	eulerToken := common.GetEnvToken(prData.Repository.NameSpace)
	if len(eulerToken) < 1 {
		logs.Error("namespace error: ", prData.Repository.NameSpace)
		return
	}
	prNumber := prData.PullRequest.Number
	repoPath := common.TrimString(prData.Repository.Path)
	owner := common.TrimString(prData.Repository.NameSpace)
	gop := models.TestOriginPr{Owner: prData.Repository.NameSpace, RepoPath: repoPath,
		PrId: prData.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryTestOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gopErr != nil {
		CreateHookTestPrOrgData(prData, &gop, 1)
		gop.UserId = su.UserId
		prId, iprErr := models.InsertTestOriginPr(&gop)
		if iprErr != nil {
			logs.Error("InsertTestOriginPr, prId: ", prId, ",iprErr: ", iprErr)
			return
		}
		gop.OrId = prId
	} else {
		updateStr := CreateHookTestPrOrgData(prData, &gop, 2)
		uprErr := models.UpdateTestOriginPr(&gop, updateStr...)
		if uprErr != nil {
			logs.Error("UpdateTestOriginPr, uprErr: ", uprErr)
			return
		}
	}
	if openFlag == 1 {
		// Will write issue comments
		igc := fmt.Sprintf(TestPrCommitComment, prData.PullRequest.User.UserName)
		AddCommentToPr(igc, owner, prData.Repository.Path, eulerToken, prData.PullRequest.Number)
		// Send private message
		igcs := fmt.Sprintf(TestPrCommitCommentSend, gop.GitUrl)
		SendPrivateLetters(eulerToken, igcs, prData.PullRequest.User.UserName)
		assigneeSlice := prData.PullRequest.Assignees
		if len(assigneeSlice) > 0 {
			for _, as := range assigneeSlice {
				igcs := fmt.Sprintf(TestPrRewiewSend, prData.PullRequest.User.UserName, gop.GitUrl)
				SendPrivateLetters(eulerToken, igcs, as.UserName)
			}
		}
	}
}

func AddHookEulerPr(prData *models.PrPayload, openFlag int) {
	gitId := prData.PullRequest.User.Id
	prCreater := prData.PullRequest.User.UserName
	su := models.StdUser{GitId: gitId, Status: 1}
	suErr := models.QueryStdUser(&su, "GitId", "Status")
	if suErr != nil {
		logs.Error("suErr: ", suErr)
	}
	if su.UserId == 0 || su.TestFlag != 2 {
		logs.Error("The user does not exist, or the test task has been completed,"+
			" GitUserId: ", prCreater, gitId)
		return
	}
	UpdateStdUsers(&su, prData.PullRequest.User.UserName, prData.PullRequest.User.UserName)
	eulerToken := common.GetEnvToken(prData.Repository.NameSpace)
	if len(eulerToken) < 1 {
		logs.Error("namespace error: ", prData.Repository.NameSpace)
		return
	}
	prNumber := prData.PullRequest.Number
	repoPath := common.TrimString(prData.Repository.Path)
	owner := common.TrimString(prData.Repository.NameSpace)
	gop := models.EulerOriginPr{Owner: owner, RepoPath: repoPath,
		PrId: prData.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryEulerOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gopErr != nil {
		CreateHookEulerPrOrgData(prData, &gop, 1)
		gop.UserId = su.UserId
		prId, iprErr := models.InsertEulerOriginPr(&gop)
		if iprErr != nil {
			logs.Error("InsertEulerOriginPr, prId: ", prId, ",iprErr: ", iprErr)
			return
		}
		gop.OrId = prId
	} else {
		updateStr := CreateHookEulerPrOrgData(prData, &gop, 2)
		uprErr := models.UpdateEulerOriginPr(&gop, updateStr...)
		if uprErr != nil {
			logs.Error("UpdateTestOriginPr, uprErr: ", uprErr)
			return
		}
	}
}

func CreateHookTestPrOrgData(hi *models.PrPayload, gop *models.TestOriginPr, flag int) []string {
	updateStr := make([]string, 0)
	gop.PrNumber = hi.PullRequest.Number
	gop.PrId = hi.PullRequest.Id
	prState := common.TrimString(hi.State)
	gop.PrState = prState
	updateStr = append(updateStr, "PrState")
	gop.GitUrl = hi.PullRequest.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	gop.Title = hi.PullRequest.Title
	updateStr = append(updateStr, "Title")
	gop.PrBody = hi.PullRequest.Body
	updateStr = append(updateStr, "PrBody")
	if flag == 1 {
		gop.GrabTime = common.GetCurTime()
	}
	labelStr := ""
	if hi.PullRequest.Labels != nil && len(hi.PullRequest.Labels) > 0 {
		for _, la := range hi.PullRequest.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
		gop.PrLabel = labelStr
		updateStr = append(updateStr, "PrLabel")
	}
	gop.PrCreate = hi.PullRequest.User.UserName
	updateStr = append(updateStr, "PrCreate")
	gop.PrUpdate = hi.PullRequest.UpdatedBy.UserName
	updateStr = append(updateStr, "PrUpdate")
	assigneeSlice := []string{}
	if len(hi.PullRequest.Assignees) > 0 {
		for _, as := range hi.PullRequest.Assignees {
			assigneeSlice = append(assigneeSlice, as.UserName)
		}

	} else {
		if len(hi.PullRequest.Assignee) > 0 {
			assigneeSlice = append(assigneeSlice, hi.PullRequest.Assignee)
		}
	}
	if len(assigneeSlice) > 0 {
		gop.PrAssignee = strings.Join(assigneeSlice, ",")
		updateStr = append(updateStr, "PrAssignee")
	}
	gop.RepoUrl = hi.Repository.Url
	updateStr = append(updateStr, "RepoUrl")
	gop.RepoPath = hi.Repository.Path
	gop.Owner = hi.Repository.NameSpace
	gop.Status = 0
	updateStr = append(updateStr, "Status")
	gop.TargetBranch = hi.TargetBranch
	updateStr = append(updateStr, "TargetBranch")
	if len(hi.PullRequest.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		gop.CreateTime = hi.PullRequest.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.PullRequest.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		gop.UpdateTime = hi.PullRequest.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.PullRequest.ClosedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		gop.ClosedTime = hi.PullRequest.ClosedAt.String()
		updateStr = append(updateStr, "ClosedTime")
	}
	if len(hi.PullRequest.MergedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		gop.MergedTime = hi.PullRequest.MergedAt.String()
		updateStr = append(updateStr, "MergedTime")
	}
	logs.Info("gop===>", gop)
	return updateStr
}

func CreateHookEulerPrOrgData(hi *models.PrPayload, gop *models.EulerOriginPr, flag int) []string {
	updateStr := make([]string, 0)
	gop.PrNumber = hi.PullRequest.Number
	gop.PrId = hi.PullRequest.Id
	prState := common.TrimString(hi.State)
	gop.PrState = prState
	updateStr = append(updateStr, "PrState")
	gop.GitUrl = hi.PullRequest.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	gop.Title = hi.PullRequest.Title
	updateStr = append(updateStr, "Title")
	gop.PrBody = hi.PullRequest.Body
	updateStr = append(updateStr, "PrBody")
	if flag == 1 {
		gop.GrabTime = common.GetCurTime()
	}
	labelStr := ""
	if hi.PullRequest.Labels != nil && len(hi.PullRequest.Labels) > 0 {
		for _, la := range hi.PullRequest.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
		gop.PrLabel = labelStr
		updateStr = append(updateStr, "PrLabel")
	}
	gop.PrCreate = hi.PullRequest.User.UserName
	updateStr = append(updateStr, "PrCreate")
	gop.PrUpdate = hi.PullRequest.UpdatedBy.UserName
	updateStr = append(updateStr, "PrUpdate")
	assigneeSlice := []string{}
	if len(hi.PullRequest.Assignees) > 0 {
		for _, as := range hi.PullRequest.Assignees {
			assigneeSlice = append(assigneeSlice, as.UserName)
		}

	} else {
		if len(hi.PullRequest.Assignee) > 0 {
			assigneeSlice = append(assigneeSlice, hi.PullRequest.Assignee)
		}
	}
	if len(assigneeSlice) > 0 {
		gop.PrAssignee = strings.Join(assigneeSlice, ",")
		updateStr = append(updateStr, "PrAssignee")
	}
	gop.RepoUrl = hi.Repository.Url
	updateStr = append(updateStr, "RepoUrl")
	gop.RepoPath = hi.Repository.Path
	gop.Owner = hi.Repository.NameSpace
	gop.Status = 0
	updateStr = append(updateStr, "Status")
	gop.TargetBranch = hi.TargetBranch
	updateStr = append(updateStr, "TargetBranch")
	if len(hi.PullRequest.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		gop.CreateTime = hi.PullRequest.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.PullRequest.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		gop.UpdateTime = hi.PullRequest.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.PullRequest.ClosedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		gop.ClosedTime = hi.PullRequest.ClosedAt.String()
		updateStr = append(updateStr, "ClosedTime")
	}
	if len(hi.PullRequest.MergedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		gop.MergedTime = hi.PullRequest.MergedAt.String()
		updateStr = append(updateStr, "MergedTime")
	}
	logs.Info("gop===>", gop)
	return updateStr
}

func UpdateHookTestPr(prData *models.PrPayload) {
	gitId := prData.PullRequest.User.Id
	prCreater := prData.PullRequest.User.UserName
	su := models.StdUser{GitId: gitId, Status: 1}
	suErr := models.QueryStdUser(&su, "GitId", "Status")
	if suErr != nil {
		logs.Error("suErr: ", suErr)
	}
	if su.UserId == 0 || su.Status != 1 || su.TestFlag == 2 {
		logs.Error("The user does not exist, or the test task has been completed,"+
			" GitUserId: ", prCreater, gitId, su)
		return
	}
	UpdateStdUsers(&su, prData.PullRequest.User.UserName, prData.PullRequest.User.UserName)
	eulerToken := common.GetEnvToken(prData.Repository.NameSpace)
	if len(eulerToken) < 1 {
		logs.Error("namespace error: ", prData.Repository.NameSpace)
		return
	}
	prNumber := prData.PullRequest.Number
	repoPath := common.TrimString(prData.Repository.Path)
	owner := common.TrimString(prData.Repository.NameSpace)
	gop := models.TestOriginPr{Owner: prData.Repository.NameSpace, RepoPath: repoPath,
		PrId: prData.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryTestOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gop.OrId == 0 {
		logs.Error("Data does not exist, gopErr: ", gopErr, gop)
		return
	}
	su.TestFlag = 2
	su.GitUserId = prCreater
	su.UpdateTime = common.GetCurTime()
	upErr := models.UpdateStdUser(&su, "TestFlag", "GitUserId", "UpdateTime")
	if upErr != nil {
		logs.Error("UpdateStdUser, upErr: ", upErr)
		return
	}
	gop.Status = 1
	gop.MergedTime = prData.PullRequest.MergedAt.String()
	gop.PrCreate = prCreater
	uprErr := models.UpdateTestOriginPr(&gop, "Status", "MergedTime", "PrCreate")
	if uprErr != nil {
		logs.Error("UpdateTestOriginPr, uprErr: ", uprErr)
		return
	}
	// Will write issue comments
	igc := fmt.Sprintf(TestPrMergComment, prData.PullRequest.User.UserName)
	AddCommentToPr(igc, owner, prData.Repository.Path, eulerToken, prData.PullRequest.Number)
	// Send private message
	igcs := fmt.Sprintf(TestPrMergCommentSend, gop.GitUrl)
	SendPrivateLetters(eulerToken, igcs, prData.PullRequest.User.UserName)
}

func UpdateHookEulerPr(prData *models.PrPayload) {
	gitId := prData.PullRequest.User.Id
	prCreater := prData.PullRequest.User.UserName
	su := models.StdUser{GitId: gitId, Status: 1}
	suErr := models.QueryStdUser(&su, "GitId", "Status")
	if suErr != nil {
		logs.Error("suErr: ", suErr)
	}
	if su.UserId == 0 || su.TestFlag != 2 {
		logs.Error("The user does not exist, or the test task has been completed,"+
			" GitUserId: ", prCreater, gitId, su)
		return
	}
	UpdateStdUsers(&su, prData.PullRequest.User.UserName, prData.PullRequest.User.UserName)
	eulerToken := common.GetEnvToken(prData.Repository.NameSpace)
	if len(eulerToken) < 1 {
		logs.Error("namespace error: ", prData.Repository.NameSpace)
		return
	}
	prNumber := prData.PullRequest.Number
	repoPath := common.TrimString(prData.Repository.Path)
	owner := common.TrimString(prData.Repository.NameSpace)
	gop := models.EulerOriginPr{Owner: owner, RepoPath: repoPath,
		PrId: prData.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryEulerOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gop.OrId == 0 {
		logs.Error("Data does not exist, gopErr: ", gopErr, gop)
		return
	}
	gop.Status = 1
	gop.MergedTime = prData.PullRequest.MergedAt.String()
	gop.PrCreate = prCreater
	gop.UserId = su.UserId
	uprErr := models.UpdateEulerOriginPr(&gop, "Status", "MergedTime", "PrCreate", "UserId")
	if uprErr != nil {
		logs.Error("UpdateTestOriginPr, uprErr: ", uprErr)
		return
	}
}

func UpdateHookPr(prData *models.PrPayload) {
	repoPath := common.TrimString(prData.Repository.Path)
	flag := CheckPrAttribute(prData.Repository.NameSpace, repoPath, prData.Title)
	if flag == 3 {
		logs.Error("prHook: ", prData)
		return
	} else if flag == 1 {
		UpdateHookTestPr(prData)
	} else {
		UpdateHookEulerPr(prData)
	}
}

func QueryIssueToPr(token, owner, repo, issueNum string) []map[string]interface{} {
	issuePr := make([]map[string]interface{}, 0)
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/issues/%v/pull_requests", owner, issueNum)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logs.Error("NewRequest, url: ", url, ",err: ", err)
		return issuePr
	}
	q := req.URL.Query()
	q.Add("access_token", token)
	q.Add("repo", repo)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("DefaultClient, url: ", url, ",err: ", err)
		return issuePr
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		read, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logs.Error("ReadAll, url: ", url, ",err: ", err)
			return issuePr
		}

		err = json.Unmarshal(read, &issuePr)
		if err != nil {
			logs.Error("Unmarshal, url: ", url, ",err: ", err)
			return issuePr
		}
	}
	return issuePr
}

func CloseIssueAddPoints(token, owner, repo, issueNum, gitUserId string, eoi models.EulerOriginIssue) bool {
	prRes := QueryIssueToPr(token, owner, repo, issueNum)
	if len(prRes) == 0 {
		logs.Error("The issue is not associated with pr, "+
			"no points will be added, repo, issueNum: ", repo, issueNum)
		return false
	}
	resBool := false
	for _, v := range prRes {
		if _, ok := v["id"]; !ok {
			continue
		}
		user := v["user"].(map[string]interface{})
		userLogin := user["login"].(string)
		prNumber := int64(v["number"].(float64))
		mergeState := v["state"].(string)
		mergeable := v["mergeable"].(bool)
		if mergeable && mergeState == "merged" && userLogin == gitUserId {
			// Calculate the points earned by users
			pointLock.Lock()
			CalculateUserPoints(token, eoi)
			pointLock.Unlock()
			resBool = true
			// Modify data status
			baseInfo := v["base"].(map[string]interface{})
			prRepo := baseInfo["repo"].(map[string]interface{})
			repoPath := prRepo["path"].(string)
			namespace := prRepo["namespace"].(map[string]interface{})
			owner := namespace["path"].(string)
			gop := models.EulerOriginPr{Owner: owner, RepoPath: repoPath, PrNumber: prNumber}
			gopErr := models.QueryEulerOriginPr(&gop, "Owner", "RepoPath", "PrNumber")
			if gop.OrId == 0 {
				logs.Error("Data does not exist, gopErr: ", gopErr, gop)
				continue
			}
			gop.Status = 1
			gop.IssueNumber = issueNum
			gop.PrState = mergeState
			gop.UpdateTime = common.GetCurTime()
			uprErr := models.UpdateEulerOriginPr(&gop, "Status", "IssueNumber", "PrState", "UpdateTime")
			if uprErr != nil {
				logs.Error("UpdateTestOriginPr, uprErr: ", uprErr)
				continue
			}
		}
	}
	return resBool
}

// Calculate the points earned by users
func CalculateUserPoints(eulerToken string, eoi models.EulerOriginIssue) {
	// Query user information
	eiu := models.EulerIssueUser{OrId: eoi.OrId}
	eiuErr := models.QueryEulerIssueUser(&eiu, "OrId")
	if eiuErr != nil || eiu.PointFlag == 2 {
		logs.Error("Points cannot be calculated or points have already been calculated, eiuErr: ", eiuErr)
		return
	}
	eiu.Status = 4
	eiu.PointFlag = 2
	eiu.UpdateTime = common.GetCurTime()
	upeiuErr := models.UpdateEulerIssueUser(&eiu, "Status", "UpdateTime", "PointFlag")
	if upeiuErr != nil {
		logs.Error("UpdateEulerIssueUser, upeiuErr: ", upeiuErr)
		return
	}
	pointValue := int64(eoi.TaskScore)
	eid := models.EulerUserIntegDetail{UserId: eiu.UserId, OrId: eoi.OrId}
	eidErr := models.QueryEulerUserIntegDetail(&eid, "UserId", "OrId")
	if eidErr == nil {
		logs.Info("The user has already calculated the points, eid: ", eid)
		return
	} else {
		eid = models.EulerUserIntegDetail{UserId: eiu.UserId, OrId: eoi.OrId,
			IntegralValue: pointValue, CreateTime: common.GetCurTime()}
		id, indErr := models.InsertEulerUserIntegDetail(&eid)
		if id > 0 {
			eic := models.EulerUserIntegCount{UserId: eiu.UserId}
			eicErr := models.QueryEulerUserIntegCount(&eic, "UserId")
			if eicErr != nil {
				eic = models.EulerUserIntegCount{UserId: eiu.UserId,
					IntegralValue: pointValue, CreateTime: common.GetCurTime()}
				eicId, ineicErr := models.InsertEulerUserIntegCount(&eic)
				if ineicErr != nil {
					logs.Error("InsertEulerUserIntegCount, ineicErr: ", ineicErr, eicId)
				}
			} else {
				eic.IntegralValue += pointValue
				upicErr := models.UpdateEulerUserIntegCount(&eic, "IntegralValue")
				if upicErr != nil {
					logs.Error("UpdateEulerUserIntegCount, upicErr: ", upicErr)
				}
			}
			// After earning points, send a private message
			eu := models.StdUser{UserId: eiu.UserId, Status: 1}
			euErr := models.QueryStdUser(&eu, "UserId", "Status")
			if euErr != nil {
				logs.Error("QueryStdUser, euErr: ", euErr)
			} else {
				iss := fmt.Sprintf(IssuePointSend, eoi.GitUrl, pointValue)
				SendPrivateLetters(eulerToken, iss, eu.UserName)
				is := fmt.Sprintf(IssueGetPointNot, eu.GitUserId, eoi.TaskScore)
				AddCommentToIssue(is, eoi.IssueNumber, eoi.Owner, eoi.RepoPath, eulerToken)
			}
		} else {
			logs.Error("InsertEulerUserIntegDetail, indErr:", indErr)
		}
	}
	eiuc := models.EulerIssueUserComplate{UserId: eiu.UserId, OrId: eoi.OrId}
	eiucErr := models.QueryEulerIssueUserComplate(&eiuc, "UserId", "OrId")
	if eiucErr != nil {
		eiuc = models.EulerIssueUserComplate{UserId: eiu.UserId, OrId: eoi.OrId,
			IssueNumber: eoi.IssueNumber, RepoPath: eoi.RepoPath, Owner: eoi.Owner,
			Status: 1, IntegralValue: pointValue, CreateTime: common.GetCurTime()}
		ucId, ucErr := models.InsertEulerIssueUserComplate(&eiuc)
		if ucErr != nil {
			logs.Error("InsertEulerIssueUserComplate, ucErr: ", ucErr, ucId)
		}
	}
}

func CalculateIssuePoints() error {
	eiu := models.QueryClosedIssueNoPoints(4, 1)
	if len(eiu) > 0 {
		for _, eu := range eiu {
			eoi := models.EulerOriginIssue{OrId: eu.OrId, Owner: eu.Owner, RepoPath: eu.RepoPath, IssueNumber: eu.IssueNumber}
			eiErr := models.QueryEulerOriginIssue(&eoi, "OrId", "Owner", "RepoPath", "IssueNumber")
			if eoi.OrId == 0 {
				logs.Error("QueryEulerOriginIssue, Data does not exist, eiErr: ", eiErr)
				continue
			}
			eulerToken := common.GetEnvToken(eu.Owner)
			if len(eulerToken) < 1 {
				logs.Error("namespace error: ", eu.Owner)
				continue
			}
			su := models.StdUser{UserId: eu.UserId, Status: 1}
			euErr := models.QueryStdUser(&su, "UserId", "Status")
			resBool := CloseIssueAddPoints(eulerToken, eu.Owner, eu.RepoPath, eu.IssueNumber, su.GitUserId, eoi)
			if resBool {
				if eu.UserId > 0 {
					if euErr != nil {
						logs.Error("QueryStdUser, euErr: ", euErr)
					} else {
						is := fmt.Sprintf(IssueClosePointsComplete, su.GitUserId, eoi.TaskScore)
						AddCommentToIssue(is, eu.IssueNumber, eu.Owner, eu.RepoPath, eulerToken)
					}
				}
			}
		}
	}
	return nil
}
