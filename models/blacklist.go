package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func QueryEulerBlackUserAll(status int8) (ebu []EulerBlackUser) {
	o := orm.NewOrm()
	if status > 0 {
		num, _ := o.Raw("select * from osi_euler_black_user where status = ?", status).QueryRows(&ebu)
		if num > 0 {
			logs.Info("QueryEulerBlackUserAll, num: ", num)
		}
	} else {
		num, _ := o.Raw("select * from osi_euler_black_user").QueryRows(&ebu)
		if num > 0 {
			logs.Info("QueryEulerBlackUserAll, num: ", num)
		}
	}
	return
}

func DelEulerUnassignBlack(id int64) {
	o := orm.NewOrm()
	err := o.Raw("delete from osi_euler_unassign_user where id = ?", id).QueryRow()
	logs.Info("DelEulerUnassignBlack", err)
}

func QueryEulerUnassignUserAll(afterDate string) (euu []EulerUnassignUser) {
	o := orm.NewOrm()
	num, _ := o.Raw("select * from osi_euler_unassign_user where "+
		"unassign_time < ? and unassign_time != ''", afterDate).QueryRows(&euu)
	if num > 0 {
		logs.Info("QueryEulerUnassignUserAll, num: ", num)
	}
	return
}

func QueryEulerIssueUserRecord(status int8, curTime string, userId int64) (eiu []EulerIssueUserRecord) {
	o := orm.NewOrm()
	num, _ := o.Raw("select * from osi_euler_issue_user_record where "+
		"status = ? and create_time < ? and user_id = ?", status, curTime, userId).QueryRows(&eiu)
	if num > 0 {
		logs.Info("QueryEulerIssueUserRecord, num: ", num)
	}
	return
}

func QueryEulerUnassignAll() (euu []EulerUnassignUser) {
	o := orm.NewOrm()
	num, _ := o.Raw("select * from osi_euler_unassign_user").QueryRows(&euu)
	if num > 0 {
		logs.Info("QueryEulerUnassignAll, num: ", num)
	}
	return
}

// Query all the data of the initial audit timeout
func QueryInitialAuditClaimIssue(afterDate string) (eiu []EulerIssueUser) {
	o := orm.NewOrm()
	num, euErr := o.Raw("select * from osi_euler_issue_user where "+
		"freeze_time < ? and status = ?", afterDate, 1).QueryRows(&eiu)
	if num > 0 {
		logs.Info("QueryInitialAuditClaimIssue, num: ", num)
	} else {
		logs.Error("QueryInitialAuditClaimIssue, euErr: ", euErr)
	}
	return
}

// Query the data of the student's uncompleted task after the instructor has passed the review
func QueryUncompletedClaimIssue(afterDate string) (eiu []EulerIssueUser) {
	o := orm.NewOrm()
	num, euErr := o.Raw("select * from osi_euler_issue_user where "+
		"assign_time < ? and status = ?", afterDate, 2).QueryRows(&eiu)
	if num > 0 {
		logs.Info("QueryUncompletedClaimIssue, num: ", num)
	} else {
		logs.Error("QueryUncompletedClaimIssue, euErr: ", euErr)
	}
	return
}
