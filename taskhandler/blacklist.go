package taskhandler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"sync"
)

var SynWait sync.WaitGroup
var BlWait sync.WaitGroup

func ForciblyCancelClaim(ei models.EulerIssueUser, relFlag int8) {
	if relFlag == 3 {
		defer BlWait.Done()
	} else {
		defer SynWait.Done()
	}
	eulerToken := common.GetEnvToken(ei.Owner)
	if len(eulerToken) < 1 {
		logs.Error("QueryInitialAuditClaimIssue, owner: ", ei.Owner)
	}
	// give up task
	delErr := models.DeleteEulerIssueUser(&ei, "UserId", "OrId", "Status")
	if delErr == nil {
		// add issue count
		models.AddEulerOrgIssueCount(ei.OrId)
		// Edit label
		eoi, eoiErr := GetIssueInfo(ei.OrId)
		if eoiErr != nil {
			logs.Error("GetIssueInfo, eoiErr: ", eoiErr)
		}
		RelReduceLabel(ei.RepoPath, ei.IssueNumber, eulerToken, ei.Owner, eoi)
		su := models.StdUser{UserId: ei.UserId, Status: 1}
		suErr := models.QueryStdUser(&su, "UserId", "Status")
		if suErr != nil {
			logs.Error("QueryStdUser, euErr: ", suErr)
		}
		// release count
		AddEulerUserUnassignCount(su.UserId, su.GitId, su.GitUserId, 2)
		is := ""
		if relFlag == 1 {
			is = fmt.Sprintf(IssueForciGiveUpNoReiew, su.GitUserId)
		} else if relFlag == 2 {
			is = fmt.Sprintf(IssueForciGiveUpSuccess, su.GitUserId)
		} else {
			is = fmt.Sprintf(IssueForciGiveUpBlack, su.GitUserId)
		}
		AddCommentToIssue(is, eoi.IssueNumber, ei.Owner, eoi.RepoPath, eulerToken)
		iss := fmt.Sprintf(IssueBlackSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, su.UserName)
		et := EulerIssueUserRecordTp{UserId: ei.UserId, OrId: eoi.OrId, IssueNumber: eoi.IssueNumber,
			RepoPath: eoi.RepoPath, Owner: ei.Owner, Status: 21}
		EulerIssueUserRecord(et)
	}
}

func GetIssueInfo(orId int64) (models.EulerOriginIssue, error) {
	eoi := models.EulerOriginIssue{OrId: orId}
	eiErr := models.QueryEulerOriginIssue(&eoi, "OrId")
	if eiErr != nil {
		logs.Error(eiErr)
		return eoi, eiErr
	}
	return eoi, eiErr
}

func RelFrozenBlackTask() error {
	// Query blacklist task
	ebu := models.QueryEulerBlackUserAll(1)
	if len(ebu) > 0 {
		// Release the problem claimed by the developer
		for _, bl := range ebu {
			// Find user information
			eiu := models.QueryEulerIssueUnfinished(bl.UserId, 3)
			if len(eiu) > 0 {
				for _, ei := range eiu {
					BlWait.Add(1)
					go ForciblyCancelClaim(ei, 3)
				}
			}
			bl.Status = 2
			upErr := models.UpdateEulerBlackUser(&bl, "Status")
			logs.Info("upErr: ", upErr)
		}
	}
	BlWait.Wait()
	RemoveUnassignBlacklist()
	return nil
}

func RemoveUnassignBlacklist() {
	// Query the data to be deleted
	euu := models.QueryEulerUnassignUserAll(common.GetCurTime())
	if len(euu) > 0 {
		for _, eu := range euu {
			models.DelEulerUnassignBlack(eu.Id)
		}
	}
}

func MonthRelUnassign() error {
	// cur month
	_, endMonth := common.GetLastMonthDate()
	unassignAll := models.QueryEulerUnassignAll()
	issueCount := int8(beego.AppConfig.DefaultInt("claimed::issue_count", 3))
	if len(unassignAll) > 0 {
		for _, un := range unassignAll {
			eiu := models.QueryEulerIssueUserRecord(2, endMonth, un.UserId)
			eiuCount := int8(len(eiu))
			if eiuCount > 0 {
				if un.CountValue > eiuCount && un.UpdateTime >= endMonth {
					un.CountValue = un.CountValue - eiuCount
					un.UnassignTime = ""
					upErr := models.UpdateEulerUserUnassigned(&un, "CountValue", "UnassignTime")
					if upErr != nil {
						logs.Error(upErr)
					}
				} else {
					if un.CountValue < issueCount && un.UpdateTime < endMonth {
						delErr := models.DelEulerUserUnassigned(&un, "Id", "UserId")
						if delErr != nil {
							logs.Error(delErr)
						}
					}
				}
			}
		}
	}
	return nil
}

func ProcOvertimeClaimedTasks() error {
	// Process data that has not been reviewed by the instructor for the first time claim
	eiu := models.QueryInitialAuditClaimIssue(common.GetCurTime())
	if len(eiu) > 0 {
		for _, ei := range eiu {
			SynWait.Add(1)
			go ForciblyCancelClaim(ei, 1)
		}
	}
	uncompletedData := models.QueryUncompletedClaimIssue(common.GetCurTime())
	if len(uncompletedData) > 0 {
		for _, ei := range uncompletedData {
			SynWait.Add(1)
			go ForciblyCancelClaim(ei, 2)
		}
	}
	SynWait.Wait()
	return nil
}
