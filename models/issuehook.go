package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"time"
)

//HookComment gitee comment model
type HookComment struct {
	User     *HookUser //Comment author information
	HtmlUrl  string    `json:"html_url"` //Comment on the url on the code cloud
	Id       int64
	Body     string    `json:"body"` //Comment content
	CreateAt time.Time `json:"created_at"`
	UpdateAt time.Time `json:"updated_at"`
}

//CommentPayload hook data triggered by a comment task operation
type CommentPayload struct {
	Action        string // Action comment
	HookName      string `json:"hook_name"`
	Password      string `json:"password"` //The password of the hook
	HookId        int64  `json:"hook_id"`  //Hook ID
	HookUrl       string `json:"hook_url"` //Hook routing
	Timestamp     string
	Sign          string       //The signature calculated by the hook based on the key
	Comment       *HookComment `json:"comment"` //Comment data
	Note          string       //Comment data of the commented target
	NoteableType  string       `json:"noteable_type"` //Target type being reviewed
	NoteableId    int64        `json:"noteable_id"`   //ID of the target being reviewed
	Title         string       //Target title being commented
	PerId         string       `json:"per_iid"`         //Target ID being reviewed
	ShortCommitId string       `json:"short_commit_id"` //The short sha in the commented commit submission
	Issue         *HookIssue   `json:"issue"`
	PullRequest   HookPr       `json:"pull_request"`
	Repository    HookRepository
}

// Delete hook related data
func DeleteEulerOriginIssueAll(eoi *EulerOriginIssue) (int64, error) {
	o := orm.NewOrm()
	errs := o.Begin()
	userId := int64(0)
	orId := eoi.OrId
	if errs == nil {
		if num, err := o.Delete(eoi, "OrId"); err == nil {
			logs.Info("delete osi_euler_origin_issue success, num: ", num, ",orId: ", orId)
		} else {
			logs.Error("delete osi_euler_origin_issue failed,", ",orId: ", orId, ", err: ", err)
			o.Rollback()
			return userId, err
		}
		var eu EulerIssueUser
		eu.OrId = orId
		if num, err := o.Delete(&eu, "OrId"); err == nil {
			logs.Info("delete osi_euler_issue_user success, num: ", num, ",orId: ", orId)
		} else {
			logs.Error("delete osi_euler_issue_user failed,", ",orId: ", orId, ", err: ", err)
			o.Rollback()
			return userId, err
		}
		var euc EulerIssueUserComplate
		euc.OrId = orId
		if num, err := o.Delete(&euc, "OrId"); err == nil {
			logs.Info("delete osi_euler_issue_user_complate success, num: ", num, ",orId: ", orId)
		} else {
			logs.Error("delete osi_euler_issue_user_complate failed,", ",orId: ", orId, ", err: ", err)
			o.Rollback()
			return userId, err
		}
		var eid EulerUserIntegDetail
		eid.OrId = orId
		eidErr := QueryEulerUserIntegDetail(&eid, "OrId")
		if eidErr == nil {
			userId = eid.UserId
			var eic EulerUserIntegCount
			eic.UserId = eid.UserId
			eicErr := QueryEulerUserIntegCount(&eic, "UserId")
			if eicErr == nil {
				if eic.IntegralValue >= eid.IntegralValue {
					eic.IntegralValue -= eid.IntegralValue
				} else {
					eic.IntegralValue = 0
				}
				ueicErr := UpdateEulerUserIntegCount(&eic, "IntegralValue", "UserId")
				if ueicErr != nil {
					logs.Error("UpdateEulerUserIntegCount, ueicErr: ", ueicErr)
					o.Rollback()
					return 0, ueicErr
				}
			}
			if num, err := o.Delete(&eid, "OrId", "UserId"); err == nil {
				logs.Info("delete osi_euler_user_integ_detail success, num: ", num, ",orId: ", orId)
			} else {
				logs.Error("delete osi_euler_user_integ_detail failed,", ",orId: ", orId, ", err: ", err)
				o.Rollback()
				return 0, err
			}
		}
		o.Commit()
	}
	return userId, nil
}

func RejectEulerOriginIssueAll(eoi *EulerOriginIssue) (int64, error) {
	o := orm.NewOrm()
	errs := o.Begin()
	userId := int64(0)
	orId := eoi.OrId
	if errs == nil {
		var eu EulerIssueUser
		eu.OrId = orId
		if num, err := o.Delete(&eu, "OrId"); err == nil {
			logs.Info("delete osi_euler_issue_user success, num: ", num, ",orId: ", orId)
		} else {
			logs.Error("delete osi_euler_issue_user failed,", ",orId: ", orId, ", err: ", err)
			o.Rollback()
			return userId, err
		}
		var euc EulerIssueUserComplate
		euc.OrId = orId
		if num, err := o.Delete(&euc, "OrId"); err == nil {
			logs.Info("delete osi_euler_issue_user_complate success, num: ", num, ",orId: ", orId)
		} else {
			logs.Error("delete osi_euler_issue_user_complate failed,", ",orId: ", orId, ", err: ", err)
			o.Rollback()
			return userId, err
		}
		var eid EulerUserIntegDetail
		eid.OrId = orId
		eidErr := QueryEulerUserIntegDetail(&eid, "OrId")
		if eidErr == nil {
			userId = eid.UserId
			var eic EulerUserIntegCount
			eic.UserId = eid.UserId
			eicErr := QueryEulerUserIntegCount(&eic, "UserId")
			if eicErr == nil {
				if eic.IntegralValue >= eid.IntegralValue {
					eic.IntegralValue -= eid.IntegralValue
				} else {
					eic.IntegralValue = 0
				}
				ueicErr := UpdateEulerUserIntegCount(&eic, "IntegralValue", "UserId")
				if ueicErr != nil {
					logs.Error("UpdateEulerUserIntegCount, ueicErr: ", ueicErr)
					o.Rollback()
					return 0, ueicErr
				}
			}
			if num, err := o.Delete(&eid, "OrId", "UserId"); err == nil {
				logs.Info("delete osi_euler_user_integ_detail success, num: ", num, ",orId: ", orId)
			} else {
				logs.Error("delete osi_euler_user_integ_detail failed,", ",orId: ", orId, ", err: ", err)
				o.Rollback()
				return 0, err
			}
		}
		o.Commit()
	}
	return userId, nil
}

func QueryEulerIssueUser(eiu *EulerIssueUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertEulerIssueUser(eiu *EulerIssueUser) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateEulerIssueUser(eiu *EulerIssueUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

func DeleteEulerIssueUser(eiu *EulerIssueUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Delete(eiu, fields...)
	return err
}

func QueryEulerBlackUser(eu *EulerBlackUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eu, field...)
	return err
}

func UpdateEulerBlackUser(eu *EulerBlackUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eu, fields...)
	return err
}

func QueryEulerIssueUnfinished(uerId int64, status int8) (eiu []EulerIssueUser) {
	o := orm.NewOrm()
	num := int64(0)
	err := error(nil)
	if status > 0 {
		num, err = o.Raw("select *"+
			" from osi_euler_issue_user where user_id = ? and status <= ? order by id desc",
			uerId, status).QueryRows(&eiu)
	} else {
		num, err = o.Raw("select *"+
			" from osi_euler_issue_user where user_id = ? order by id desc",
			uerId).QueryRows(&eiu)
	}

	if err != nil {
		logs.Info("QueryEulerIssueUnfinished, The current user has not claimed the task, err: ", err)
	} else {
		logs.Info("QueryEulerIssueUnfinished, The number of tasks claimed by the current user: ", num)
	}
	return
}

func InsertEulerIssueUserRecord(er *EulerIssueUserRecord) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(er)
	return id, err
}

func QueryEulerIssueUserRecordset(uerId, orId int64, status int8) (eiu []EulerIssueUserRecord) {
	o := orm.NewOrm()
	num, err := o.Raw("select *"+
		" from osi_euler_issue_user_record where user_id = ? and or_id = ? and status = ? order by id desc",
		uerId, orId, status).QueryRows(&eiu)
	if err != nil {
		logs.Info("QueryEulerIssueUserRecordset, err: ", err)
	} else {
		logs.Info("QueryEulerIssueUserRecordset, num: ", num)
	}
	return
}

func QueryEulerUserIntegDetail(eiu *EulerUserIntegDetail, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertEulerUserIntegDetail(eiu *EulerUserIntegDetail) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateEulerUserIntegDetail(eiu *EulerUserIntegDetail, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

func QueryEulerUserIntegCount(eiu *EulerUserIntegCount, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertEulerUserIntegCount(eiu *EulerUserIntegCount) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateEulerUserIntegCount(eiu *EulerUserIntegCount, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

func QueryEulerIssueUserComplate(eiu *EulerIssueUserComplate, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertEulerIssueUserComplate(eiu *EulerIssueUserComplate) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func AddEulerOrgIssueCount(orId int64) {
	o := orm.NewOrm()
	err := o.Raw("update osi_euler_origin_issue set record_count = record_count + 1 where or_id = ?", orId).QueryRow()
	logs.Info("AddEulerOrgIssueCount", err)
}
