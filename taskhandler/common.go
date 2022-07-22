package taskhandler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"osi-task-manager/util"
)

// Get the constant of the original issue
const (
	//GiteOrgInfoURL get gitee org info url
	GiteOrgInfoURL = `https://gitee.com/api/v5/orgs/%v?access_token=%v`
	//GiteOrgReposURL get all repository url
	GiteOrgReposURL = `https://gitee.com/api/v5/orgs/%v/repos?access_token=%v&type=all&page=%v&per_page=%v`
	//GiteRepoIssuesURL get issue list url
	GiteRepoIssuesURL = `https://gitee.com/api/v5/repos/%v/%v/issues?access_token=%v&state=%v&sort=created&direction=desc&page=%v&per_page=%v`
	//GiteRepoBranch get repo branch url
	GiteRepoBranch = `https://gitee.com/api/v5/repos/%v/%v/branches?access_token=%v`
	//RepoInfoURL get repo info url
	RepoInfoURL = "https://api.openeuler.org/pkgmanage/packages/packageInfo?table_name=openEuler_LTS_20.03&pkg_name=%s"
	perPage     = 50
	//IssueType Types of issues crawled
	CIssueType = "开源实习"
)

const (
	// Notification of successful claim task
	IssueClaimSuccess     = `@%v , 感谢您认领此任务, 请及时跟导师沟通, 导师审核通过后才能承担此任务, 否则任务无效.`
	IssueClaimSuccessSend = `%v, 感谢您认领此任务, 请及时跟导师沟通, 导师审核通过后才能承担此任务, 否则任务无效.`

	// The user has not signed up or signed up successfully, unable to receive the task
	IssueClaiNotSigned     = `@%v , 您还未通过开源实习资格申请, 无法领取此任务, 请先完成[实习申请](https://www.openeuler.org/zh/internship/).`
	IssueClaiNotSignedSend = `%v, 还未通过开源实习资格申请, 无法领取此任务, 请先完成[实习申请](https://www.openeuler.org/zh/internship/).`

	MindSporeIssueClaiNotSigned     = `@%v , 您当前还未通过MindSpore社区实习申请, 请按照[开源实习官网]要求先提交申请(https://www.openeuler.org/zh/internship/).`
	MindSporeIssueClaiNotSignedSend = `%v, 当前还未通过MindSpore社区实习申请, 请按照[开源实习官网]要求先提交申请(https://www.openeuler.org/zh/internship/).`

	// The user did not complete the test task
	IssueTestPrCommit     = `@%v , 您还未通过开源实习测试任务, 无法领取此任务, 请先在 %v 仓完成[测试任务](https://gitee.com/openeuler-competition/opensource-internship/issues).`
	IssueTestPrCommitSend = `%v, 还未通过开源实习测试任务, 无法领取此任务, 请先在 %v 仓完成[测试任务](https://gitee.com/openeuler-competition/opensource-internship/issues).`

	// Notification of claim failure
	IssueClaimFailure     = `@%v , 您认领的任务已经达上限,无法再次领取新的任务.`
	IssueClaimFailureSend = `您认领的任务已经达上限,无法再次领取新的任务, 请先提交已认领的任务或者取消已认领的任务.`

	// approve
	IssueStdApproveTask     = `@%v , 恭喜您已成功领取该任务, 请及时处理任务. 认领任务>导师审核认领资格>处理任务>提交任务>导师审核>pr合入>获得积分.`
	IssueStdApproveTaskSend = `%v, 您已成功领取该任务, 请及时处理任务. 认领任务>导师审核认领资格>处理任务>提交任务>导师审核>pr合入>获得积分.`

	//unapprove
	IssueStdUnApproveTask     = `@%v , 您此次任务认领失败, 请认领其他任务, 非常感谢您的参与.`
	IssueStdUnApproveTaskSend = `%v, 此任务认领失败, 请认领其他任务, 非常感谢您的参与.`

	// completed
	IssueClaimCheckSubmit      = `@%v , 您还未通过开源实习资格的申请, 无法提交此任务, 请先完成[实习申请](https://www.openeuler.org/zh/internship/).`
	IssueClaimCheckSubmitSend  = `%v, 提交此任务前, 您还未通过开源实习资格的申请, 无法提交此任务, 请先完成[实习申请](https://www.openeuler.org/zh/internship/).`
	IssueClaimTestPrSubmit     = `@%v , 您还未通过开源实习测试任务, 无法提交此任务, 请先在 %v 仓完成[测试任务](https://gitee.com/openeuler-competition/opensource-internship/issues).`
	IssueClaimTestPrSubmitSend = `%v, 提交此任务前, 还未通过开源实习测试任务, 无法提交任务, 请先在 %v 仓完成[测试任务](https://gitee.com/openeuler-competition/opensource-internship/issues).`
	IssueClaimComplete         = `@%v , 请关注您提交的pr审核进度, 跟进相关负责人审核, pr合入后可获得积分. 注: 提交pr时, 请务必在pr描述里添加此issue编号(#issue编号), 谢谢!`
	IssueClaimCompleteSend     = `%v, 任务已通知导师审核, 请关注当前issue, 及时联系导师, 跟进审核进度, 提交pr时, 请在pr描述里添加此issue编号(#issue编号).`

	// unassign
	IssueGiveUpSuccess        = `@%v , 您已取消此任务认领.`
	IssueGiveUpSuccessSend    = `%v, 此任务您已取消, 可以继续认领其他任务.`
	IssueBlackClaimFailure    = `@%v , 您无法处理此任务, 请知晓.`
	IssueUnassignClaimCount   = `@%v , 您取消认领已超过规定次数, 账号将被冻结1个月, 冻结期间无法领取任务.`
	IssueUnassignedClaimCount = `@%v , 您取消认领已超过规定次数, 账号将被冻结1个月, 冻结期间无法领取任务.`
	IssueUncompleteClaimCount = `由于参赛者: @%v 取消认领已超过规定次数, 账号将被冻结1个月, 冻结期间无法处理任务.`

	// fail
	IssueTutRelSuccessTask    = `@%v , 您已经释放 @%v 认领的任务, 任务将再次进入任务池, 后续他人可以继续认领此任务.`
	IssueStdRelTaskSubmit     = `@%v , 您提交的任务, 已被导师释放, 如有疑问, 请联系导师, 谢谢!`
	IssueStdRelTaskSubmitSend = `%v, 您提交的此任务, 已被导师 @%v 释放, 如有疑问, 请联系导师, 谢谢!`

	// release
	IssueForciGiveUpNoReiew = `@%v , 由于您认领的任务未在规定时间通过导师的审核, 此次认领无效, 任务被释放.`
	IssueForciGiveUpSuccess = `@%v , 由于您认领的任务未在规定时间提交完成, 此次认领无效, 任务被释放.`

	// Claim the same task notification multiple times
	IssueClaimSameTask = `@%v , 您已认领过当前任务, 无法再次认领.`
	IssueClaimSameSend = `%v, 您已经认领过此任务, 无法再次认领.`

	// The task has been preemptively claimed by others
	IssueClaimPree     = `@%v , 您无法认领此任务, 已被他人认领.`
	IssueClaimPreeSend = `%v, 您无法认领此任务, 已被他人认领.`
	// The user cannot claim the current task, the information is wrong
	IssueClaimWrong     = `@%v , 处理失败, 稍后重试.`
	IssueClaimWrongSend = `%v, 处理失败, 稍后重试.`

	// Submit task
	IssueClaimSubmit         = `@%v , 任务认领者才能提交此任务.`
	IssueClaimSubmitComplete = `@%v , 任务已审核完成, 请查看私信或者官网获取结果.`
	// Give up the task
	IssueGiveUpTask     = `@%v , 认领此任务的开发者才能取消此任务.`
	IssueGiveUpTaskSend = `%v, 认领此任务的开发者才能取消此任务. 请确认已报名成为授权用户, 否则命令无效.`
	IssueGiveUpFailure  = `@%v , 任务已审核完成, 无法取消此任务.`

	IssueStateProc = `@%v , 此issue状态, 只能由issue责任人改变, 如需改变, 请先变更issue责任人.`
	// After earning points, send a private message
	IssuePointSend         = `您提交的任务: %v 已审核通过, 当前任务获得的积分为: %v分.`
	IssueBlackSend         = `您认领的任务: %v 已被取消, 请知晓.`
	IssueTutAssignTask     = `@%v , 已报名认证过的导师, 才能审核任务, 谢谢!`
	IssueTutAssignTaskSend = `%v, 您无法审核当前任务, 请先报名认证审核资格, 谢谢!`
	IssueTutUnassignTask   = `@%v , 用户已经放弃此任务, 请知晓, 谢谢!`
	IssueStdRegisterTask   = `@%v , 用户还未报名, 无法认领此任务, 请先通知报名成认证用户, 谢谢!`

	// Submit task
	IssueReviewTaskSend = `%v, 此任务学生已提交完成, 请尽快合入pr, 如果pr不符合要求, 请在issue评论区给出指导意见, 重新修改pr; 或者输入: /intern-fail, 
此任务将被释放到任务池, 其他人可以继续认领此任务.`
	IssueStdUnapproveComplete     = `@%v , 请先联系导师审核此任务的认领资格, 否则提交完成无效, 参考步骤: 认领任务>导师审核认领资格>处理任务>提交任务>导师审核pr通过>获得积分.`
	IssueStdUnapproveCompleteSend = `%v, 请先联系导师审核此任务的认领资格, 否则提交完成无效, 参考步骤: 认领任务>导师审核认领资格>处理任务>提交任务>导师审核pr通过>获得积分.`
	IssueTutFailTask              = `@%v , 已报名认证过的导师, 才能取消他人认领的任务, 谢谢!`
	IssueClosePointsComplete      = `@%v , 您的pr已合入社区, 已获得 %v 积分, 谢谢您的参与.`
	IssueClosedByTry              = `issue: %v, 已经被@%v 关闭, 如果未提交pr, 请及时提交并关联issue编号, 如有其他疑问, 请及时联系您的导师处理, 谢谢!`
	IssueForciGiveUpBlack         = `@%v , 由于您违规操作, 此次认领无效.`
	// Unable to close issue notification
	IssueUnableCloseNot = `@%v , 关闭当前issue条件为=>学生完成pr, 导师在issue评论区输入:/intern-done,正常关闭issue,学生获得积分; 导师issue评论区输入:/close, 任何情况下可以关闭.`
	IssueGetPointNot    = `@%v , 恭喜您提交的任务成果已通过导师审核, 已获得 %v 积分, 谢谢您的参与.`
)

const (
	// test pr
	TestPrCommitComment     = `@%v , 感谢您提交此pr, 我们会尽快评估此pr, 谢谢!`
	TestPrCommitCommentSend = `感谢您提交此pr: %v, 我们会尽快评估此pr, 评审通过后, 请认领具体任务, 谢谢!`
	TestPrRewiewSend        = `参赛者: @%v , 已提交测试任务pr: %v, 请尽快完成pr的合入, 用户才能认领任务, 如果不能通过, 请在评论区告知用户原因, 谢谢!`
	TestPrMergComment       = `@%v , 此pr已审核通过, 可以正常领取任务啦, 谢谢!`
	TestPrMergCommentSend   = `测试任务pr: %v, 已审核通过, 可以正常领取任务啦, 谢谢!`
)

const (
	//IssueRejectState issue state rejected
	IssueRejectState = "rejected"
	//IssueCloseState issue state closed
	IssueCloseState = "closed"
	//IssueProgressState issue  state progressing
	IssueProgressState = "progressing"
	//IssueOpenState issue state open
	IssueOpenState = "open"
)

const (
	IssueComment = "IssueCommentEvent"
	PullRequest  = "PullRequestEvent"
	IssueRequest = "IssueEvent"
)

var (
	TaskScoreRegZh = regexp.MustCompile(`【任务分值】(?s:(.*?))$`)
	TaskScoreRegEn = regexp.MustCompile(`\[任务分值\](?s:(.*?))$`)
	RegexpDigit    = regexp.MustCompile(`([0-9]+)`)
)

type StaticIssueInfo struct {
	WeekIssueCount         int64
	WeekIssueCommentCount  int64
	WeekPullRequestCount   int64
	monthIssueCount        int64
	monthIssueCommentCount int64
	monthPullRequestCount  int64
	TotalIssueCount        int64
	TotalIssueCommentCount int64
	TotalPullRequestCount  int64
}

type StaticIssueTime struct {
	WeekIssueStartTime  string
	WeekIssueEndTime    string
	MonthIssueStartTime string
	MonthIssueEndTime   string
	TotalIssueTime      string
}

//AddCommentToIssue Add a comment to the issue
func AddCommentToIssue(msg, issueNum, owner, repo, access string) {
	url := fmt.Sprintf(`https://gitee.com/api/v5/repos/%v/%v/issues/%v/comments`, owner, repo, issueNum)
	param := fmt.Sprintf(`{"access_token": "%s","body":"%s"}`, access, msg)
	res, err := util.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Add issue comment back:", res)
}

//AddCommentToPr Add a comment to the pr
func AddCommentToPr(msg, owner, repo, access string, prNumber int64) {
	url := fmt.Sprintf(`https://gitee.com/api/v5/repos/%v/%v/pulls/%v/comments`, owner, repo, prNumber)
	param := fmt.Sprintf(`{"access_token": "%s","body":"%s"}`, access, msg)
	res, err := util.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Add pr comment back:", res)
}

//SendPrivateLetters Send a private message to a gitee user
func SendPrivateLetters(access, content, useName string) {
	url := "https://gitee.com/api/v5/notifications/messages"
	param := fmt.Sprintf(`{"access_token":"%s","username":"%s","content":"%s"}`, access, useName, content)
	res, err := util.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Send private message:", res)
}

// Query label
func QueryIssueLabels(token, repo, issueNum, owner string) ([]string, []string) {
	allLabelSlice := make([]string, 0)
	labelSlice := make([]string, 0)
	totalLabel := beego.AppConfig.String("totallabel")
	totalLabelList := strings.Split(totalLabel, ",")
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/issues/%v/labels?access_token=%v", owner, repo, issueNum, token)
	labelData, err := util.HTTPGet(url)
	if err == nil && labelData != nil {
		for _, value := range labelData {
			if _, ok := value["id"]; !ok {
				logs.Error("QueryIssueLabels， labelData, err: ", ok)
				continue
			}
			labelStr := value["name"].(string)
			allLabelSlice = append(allLabelSlice, labelStr)
			labFlag := false
			for _, lab := range totalLabelList {
				if strings.ToLower(labelStr) == strings.ToLower(lab) {
					labFlag = true
					break
				}
			}
			if !labFlag {
				labelSlice = append(labelSlice, labelStr)
			}
		}
	}
	return labelSlice, allLabelSlice
}

func AddIssueLabel(token, repo, issueNum, owner, label string) bool {
	labelStr := label
	labelSlice := strings.Split(label, ",")
	if len(labelSlice) > 0 {
		laSlice := []string{}
		for _, la := range labelSlice {
			laSlice = append(laSlice, fmt.Sprintf("\"%v\"", la))
		}
		if len(laSlice) > 0 {
			labelStr = strings.Join(laSlice, ",")
		}
	}
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/issues/%v/labels?access_token=%v", owner, repo, issueNum, token)
	reqBody := fmt.Sprintf("[%v]", labelStr)
	logs.Info("UpdateIssueLabels, reqBody: ", reqBody)
	resp, err := util.HTTPPostLabel(url, reqBody)
	if err != nil {
		logs.Error("UpdateIssueLabels, Failed to update label,  err: ", err)
		return false
	}
	if len(resp) > 0 {
		if _, ok := resp[0]["id"]; !ok {
			logs.Error("UpdateIssueLabels, Failed to update label, err: ", ok)
			return false
		}
		logs.Info("Update label succeeded, label: ", label)
	}
	return true
}

func UpdateIssueLabels(token, repo, issueNum, owner, label string) bool {
	labelStr := label
	labelSlice := strings.Split(label, ",")
	if len(labelSlice) > 0 {
		laSlice := []string{}
		for _, la := range labelSlice {
			laSlice = append(laSlice, fmt.Sprintf("\"%v\"", la))
		}
		if len(laSlice) > 0 {
			labelStr = strings.Join(laSlice, ",")
		}
	}
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/issues/%v/labels?access_token=%v", owner, repo, issueNum, token)
	reqBody := fmt.Sprintf("[%v]", labelStr)
	logs.Info("UpdateIssueLabels, reqBody: ", reqBody)
	resp, err := util.HTTPPut(url, reqBody)
	if err != nil {
		logs.Error("UpdateIssueLabels, Failed to update label,  err: ", err)
		return false
	}
	if len(resp) > 0 {
		if _, ok := resp[0]["id"]; !ok {
			logs.Error("UpdateIssueLabels, Failed to update label, err: ", ok)
			return false
		}
		logs.Info("Update label succeeded, label: ", label)
	}
	return true
}

// issue status transition
func IssueStateRev(issueState string) (statusName string) {
	if issueState != "" && len(issueState) > 1 {
		if issueState == "待办的" || issueState == "开启的" ||
			strings.ToLower(issueState) == "open" {
			statusName = "open"
		} else if issueState == "进行中" || strings.ToLower(issueState) == "started" ||
			strings.ToLower(issueState) == "progressing" {
			statusName = "progressing"
		} else if issueState == "已完成" || strings.ToLower(issueState) == "closed" {
			statusName = "closed"
		} else if issueState == "已拒绝" || strings.ToLower(issueState) == "rejected" {
			statusName = "rejected"
		} else if issueState == "已挂起" || strings.ToLower(issueState) == "suspended" {
			statusName = "suspended"
		} else {
			statusName = issueState
		}
	}
	return
}

//Get public updates from users
func GetUserPublicUpEvents(userName, accessToken, ownerList string, prevId, limit int64, sTime StaticIssueTime) StaticIssueInfo {
	localPrevId := prevId
	localLimit := limit
	sii := StaticIssueInfo{}
	for {
		url := ""
		if localPrevId > 0 {
			url = fmt.Sprintf("https://gitee.com/api/v5/users/%v/events/public?access_token=%v&prev_id=%v&limit=%v",
				userName, accessToken, localPrevId, localLimit)
		} else {
			url = fmt.Sprintf("https://gitee.com/api/v5/users/%v/events/public?access_token=%v&limit=%v",
				userName, accessToken, localLimit)
		}
		publicData, err := util.HTTPGet(url)
		if err == nil && publicData != nil && len(publicData) > 0 {
			for _, value := range publicData {
				if _, ok := value["id"]; !ok {
					logs.Error("publicData, err: ", ok, "url: ", url)
					continue
				}
				if localPrevId == 0 {
					localPrevId = int64(value["id"].(float64))
				}
				if localPrevId > int64(value["id"].(float64)) {
					localPrevId = int64(value["id"].(float64))
				}
				staticIssueComment(value, ownerList, userName, sTime, &sii)
			}
		} else {
			break
		}
	}
	return sii
}

func staticIssueComment(value map[string]interface{},
	ownerList, userName string, sTime StaticIssueTime, sii *StaticIssueInfo) {
	if value == nil || value["created_at"] == nil {
		logs.Error("created_at is empty to exit, created_at: ", value["created_at"])
		return
	}
	createdAt := value["created_at"].(string)
	ct := int64(0)
	wst := int64(0)
	wet := int64(0)
	mst := int64(0)
	met := int64(0)
	tst := int64(0)
	if len(createdAt) > 0 {
		if len(createdAt) > 19 {
			ct = util.TimeStrToInt(createdAt[:19], "2006-01-02T15:04:05")
		} else {
			ct = util.TimeStrToInt(createdAt, "2006-01-02T15:04:05")
		}
	}
	if len(sTime.WeekIssueStartTime) > 0 {
		if len(sTime.WeekIssueStartTime) <= 10 {
			sTime.WeekIssueStartTime = sTime.WeekIssueStartTime + " 00:00:00"
		}
		wst = util.TimeStrToInt(sTime.WeekIssueStartTime, "2006-01-02 15:04:05")

	}
	if len(sTime.WeekIssueEndTime) > 0 {
		if len(sTime.WeekIssueEndTime) <= 10 {
			sTime.WeekIssueEndTime = sTime.WeekIssueEndTime + " 00:00:00"
		}
		wet = util.TimeStrToInt(sTime.WeekIssueEndTime, "2006-01-02 15:04:05")
	}
	if len(sTime.MonthIssueStartTime) > 0 {
		if len(sTime.MonthIssueStartTime) <= 10 {
			sTime.MonthIssueStartTime = sTime.MonthIssueStartTime + " 00:00:00"
		}
		mst = util.TimeStrToInt(sTime.MonthIssueStartTime, "2006-01-02 15:04:05")

	}
	if len(sTime.MonthIssueEndTime) > 0 {
		if len(sTime.MonthIssueEndTime) <= 10 {
			sTime.MonthIssueEndTime = sTime.MonthIssueEndTime + " 00:00:00"
		}
		met = util.TimeStrToInt(sTime.MonthIssueEndTime, "2006-01-02 15:04:05")
	}
	if len(sTime.TotalIssueTime) > 0 {
		if len(sTime.TotalIssueTime) <= 10 {
			sTime.TotalIssueTime = sTime.TotalIssueTime + " 00:00:00"
		}
		tst = util.TimeStrToInt(sTime.TotalIssueTime, "2006-01-02 15:04:05")

	}
	if value["repo"] == nil || len(value["repo"].(map[string]interface{})) == 0 {
		logs.Error("repo is empty to exit, repo: ", value["repo"])
		return
	}
	repoMap := value["repo"].(map[string]interface{})
	if repoMap["namespace"] == nil || len(repoMap["namespace"].(map[string]interface{})) == 0 {
		logs.Error("namespace is empty to exit, namespace: ", repoMap["namespace"])
		return
	}
	namespaceMap := repoMap["namespace"].(map[string]interface{})
	if namespaceMap["path"] == nil || len(namespaceMap["path"].(string)) == 0 {
		logs.Error("path is empty to exit, path: ", namespaceMap["path"])
		return
	}
	path := namespaceMap["path"].(string)
	pathFlag := false
	ownSlice := strings.Split(ownerList, ",")
	for _, os := range ownSlice {
		if path == os {
			pathFlag = true
		}
	}
	if !pathFlag {
		logs.Error("path: ", path, ",Not in the current organization and not participating in statistics")
		return
	}
	if value["type"] == nil || len(value["type"].(string)) == 0 {
		logs.Error("type is empty to exit, type: ", value["type"])
		return
	}
	switch value["type"].(string) {
	case IssueComment:
		if value["payload"] == nil || len(value["payload"].(map[string]interface{})) == 0 {
			logs.Error("payload is empty to exit, payload: ", value["payload"])
			return
		}
		payloadMap := value["payload"].(map[string]interface{})
		if payloadMap["comment"] == nil || len(payloadMap["comment"].(map[string]interface{})) == 0 {
			logs.Error("comment is empty to exit, comment: ", payloadMap["comment"])
			return
		}
		commentMap := payloadMap["comment"].(map[string]interface{})
		if commentMap["user"] == nil || len(commentMap["user"].(map[string]interface{})) == 0 {
			logs.Error("user is empty to exit, user: ", commentMap["user"])
			return
		}
		userMap := commentMap["user"].(map[string]interface{})
		if userMap["login"] == nil || len(userMap["login"].(string)) == 0 {
			logs.Error("login is empty to exit, login: ", userMap["login"])
			return
		}
		login := userMap["login"].(string)
		if userName == login {
			if wst <= ct && ct <= wet {
				sii.WeekIssueCommentCount += 1
			}
			if mst <= ct && ct <= met {
				sii.monthIssueCommentCount += 1
			}
			if tst <= ct {
				sii.TotalIssueCommentCount += 1
			}
		}
	case PullRequest:
		if value["payload"] == nil || len(value["payload"].(map[string]interface{})) == 0 {
			logs.Error("payload is empty to exit, payload: ", value["payload"])
			return
		}
		payloadMap := value["payload"].(map[string]interface{})
		if payloadMap["head"] == nil || len(payloadMap["head"].(map[string]interface{})) == 0 {
			logs.Error("head is empty to exit, head: ", payloadMap["head"])
			return
		}
		headMap := payloadMap["head"].(map[string]interface{})
		if headMap["user"] == nil || len(headMap["user"].(map[string]interface{})) == 0 {
			logs.Error("user is empty to exit, user: ", headMap["user"])
			return
		}
		userMap := headMap["user"].(map[string]interface{})
		if userMap["login"] == nil || len(userMap["login"].(string)) == 0 {
			logs.Error("login is empty to exit, login: ", userMap["login"])
			return
		}
		login := userMap["login"].(string)
		if userName == login {
			if wst <= ct && ct <= wet {
				sii.WeekPullRequestCount += 1
			}
			if mst <= ct && ct <= met {
				sii.monthPullRequestCount += 1
			}
			if tst <= ct {
				sii.TotalPullRequestCount += 1
			}
		}
	case IssueRequest:
		if value["payload"] == nil || len(value["payload"].(map[string]interface{})) == 0 {
			logs.Error("payload is empty to exit, payload: ", value["payload"])
			return
		}
		payloadMap := value["payload"].(map[string]interface{})
		if payloadMap["user"] == nil || len(payloadMap["user"].(map[string]interface{})) == 0 {
			logs.Error("user is empty to exit, user: ", payloadMap["user"])
			return
		}
		userMap := payloadMap["user"].(map[string]interface{})
		if userMap["login"] == nil || len(userMap["login"].(string)) == 0 {
			logs.Error("login is empty to exit, login: ", userMap["login"])
			return
		}
		login := userMap["login"].(string)
		if userName == login {
			if wst <= ct && ct <= wet {
				sii.WeekIssueCount += 1
			}
			if mst <= ct && ct <= met {
				sii.monthIssueCount += 1
			}
			if tst <= ct {
				sii.TotalIssueCount += 1
			}
		}
	}
	return
}

func TaskDueDateByPoints(point int) (dueDate int) {
	if point <= 5 {
		dueDate = 3650
	} else if point > 5 && point <= 10 {
		dueDate = 3650
	} else if point > 10 && point < 20 {
		dueDate = 3650
	} else if point >= 20 {
		dueDate = 3650
	}
	return
}
