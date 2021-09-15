package models

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"osi-task-manager/common"
)

type PointValue struct {
	Integration int64
}

func QueryOpenEulerIssueAll() (eoi []EulerOriginIssue) {
	o := orm.NewOrm()
	var num int64
	num, err := o.Raw("select * from osi_euler_origin_issue").QueryRows(&eoi)
	if num > 0 {
		logs.Info("QueryOpenEulerIssueAll, num: ", num)
	} else {
		logs.Error("QueryOpenEulerIssueAll, err: ", err)
	}
	return
}

func QueryOpenStdUserAll(flag int) (eu []StdUser) {
	o := orm.NewOrm()
	var num int64
	var err = error(nil)
	if flag == 1 {
		num, err = o.Raw("select * from osi_std_user where user_id in " +
			"(select user_id from osi_euler_user_integ_count order by integral_value desc)").QueryRows(&eu)
	} else {
		num, err = o.Raw("select * from osi_std_user where user_id not in " +
			"(select user_id from osi_euler_user_integ_count order by integral_value desc)").QueryRows(&eu)
	}
	if num > 0 {
		logs.Info("QueryOpenStdUserAll, num: ", num)
	} else {
		logs.Error("QueryOpenStdUserAll, err: ", err)
	}
	return
}

func QueryEulerUserIntegDetailValue(pv *PointValue, startTime, endTime string, userId int64) {
	o := orm.NewOrm()
	if len(startTime) > 1 {
		err := o.Raw("select sum(integral_value) as integration FROM osi_euler_user_integ_detail "+
			"where user_id = ? and create_time >= ? and create_time < ?", userId, startTime, endTime).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailValue, err: ", err)
		}
	} else {
		err := o.Raw("select sum(integral_value) as integration "+
			"FROM osi_euler_user_integ_detail where user_id = ?", userId).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailValue, err: ", err)
		}
	}
	return
}

func QueryEulerUserIntegDetailCount(pv *PointValue, startTime, endTime string, userId int64) {
	o := orm.NewOrm()
	if len(startTime) > 1 {
		err := o.Raw("select count(or_id) as integration FROM osi_euler_user_integ_detail "+
			"where user_id = ? and create_time >= ? and create_time < ? and integral_value > ?",
			userId, startTime, endTime, 0).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailCount, err: ", err)
		}
	} else {
		err := o.Raw("select count(or_id) as integration FROM osi_euler_user_integ_detail "+
			"where user_id = ? and integral_value > ?", userId, 0).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailCount, err: ", err)
		}
	}
	return
}

func (elt *EmailList) Read(field ...string) ([]EmailList, error) {
	o := orm.NewOrm()
	var el []EmailList
	var num int64
	num, err := o.Raw("select *"+
		" from osi_email_list where email_type = ?", elt.EmailType).QueryRows(&el)
	if err == nil && num > 0 {
		return el, nil
	}
	logs.Error("osi_email_list ,err: ", err)
	return el, err
}

func QueryUserPointsCount() (count int64) {
	sql := fmt.Sprintf(`SELECT COUNT(id) total FROM osi_euler_user_integ_count 
where integral_value > 0 order by id asc`)
	res := struct {
		Total int64
	}{}
	o := orm.NewOrm()
	err := o.Raw(sql).QueryRow(&res)
	if err != nil {
		logs.Error("QueryUserPointsCount, err: ", err)
		return 0
	}
	return res.Total
}

func QueryTotalPointsData(currentPage, pageSize int) (eu []EulerUserIntegCount) {
	startSize := (currentPage - 1) * pageSize
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM osi_euler_user_integ_count where integral_value > 0 "+
		"order by integral_value desc limit ? offset ?",
		pageSize, startSize).QueryRows(&eu)
	if err == nil && num > 0 {
		logs.Info("QueryTotalPointsData, search num: ", num)
	} else {
		logs.Error("QueryTotalPointsData, cur_time:",
			common.GetCurTime(), ",err: ", err)
	}
	return
}

func QueryUserPointsDetailCount(userId int64) (count int64) {
	sql := fmt.Sprintf(`SELECT COUNT(id) total FROM osi_euler_user_integ_detail 
where user_id = %d and integral_value > 0 order by id asc`, userId)
	res := struct {
		Total int64
	}{}
	o := orm.NewOrm()
	err := o.Raw(sql).QueryRow(&res)
	if err != nil {
		logs.Error("QueryUserPointsDetailCount, err: ", err)
		return 0
	}
	return res.Total
}

func QueryUserPointsDetail(currentPage, pageSize int, userId int64) (eu []EulerUserIntegDetail) {
	startSize := (currentPage - 1) * pageSize
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM osi_euler_user_integ_detail where user_id = %d and integral_value > 0 "+
		"order by id asc limit ? offset ?", userId,
		pageSize, startSize).QueryRows(&eu)
	if err == nil && num > 0 {
		logs.Info("QueryUserPointsDetail, search num: ", num)
	} else {
		logs.Error("QueryUserPointsDetail, cur_time:",
			common.GetCurTime(), ",err: ", err)
	}
	return
}
