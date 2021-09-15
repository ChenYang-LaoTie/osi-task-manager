package models

import "github.com/astaxie/beego/orm"

func QueryStdUser(su *StdUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(su, field...)
	return err
}

func InsertStdUser(su *StdUser) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(su)
	return id, err
}

func UpdateStdUser(su *StdUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(su, fields...)
	return err
}

func DelStdUser(su *StdUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Delete(su, fields...)
	return err
}

func QueryTutUser(tu *TutUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(tu, field...)
	return err
}

func InsertTutUser(tu *TutUser) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(tu)
	return id, err
}

func UpdateTutUser(tu *TutUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(tu, fields...)
	return err
}

func DelTutUser(tu *TutUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Delete(tu, fields...)
	return err
}

func QueryLoginUser(lut *LoginUserToken, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(lut, field...)
	return err
}