package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//HookPr gitee issue model
type HookPr struct {
	Id          int64
	Number      int64
	HtmlUrl     string `json:"html_url"` //Comment on the url on the code cloud
	Title       string
	Labels      []IssueLabel `json:"labels"`
	State       string       `json:"state"`
	Body        string       `json:"body"`
	CreateAt    time.Time    `json:"created_at"`
	UpdateAt    time.Time    `json:"updated_at"`
	ClosedAt    time.Time    `json:"closed_at"`
	MergedAt    time.Time    `json:"merged_at"`
	User        HookUser     `json:"user"`
	Assignee    string       `json:"assignee"`
	Assignees   []HookUser   `json:"assignees"`
	Tester      string       `json:"tester"`
	Testers     []HookUser   `json:"testers"`
	Base        PrBase       `json:"base"`
	Merged      bool
	Mergeable   bool
	MergeStatus string   `json:"merge_status"`
	UpdatedBy   HookUser `json:"updated_by"`
	Comments    int64
	commits     int64
}

//Repository gitee repository model
type PrBase struct {
	Label string         `json:"label"`
	Ref   string         `json:"ref"`
	User  HookUser       `json:"user"`
	Repo  HookRepository `json:"repo"`
}

type PrPayload struct {
	HookId       int64    `json:"hook_id"`   //  Hook id.
	HookUrl      string   `json:"hook_url"`  // route
	HookName     string   `json:"hook_name"` // issue_hooks。
	Password     string   `json:"password"`  // Hook code
	Action       string   //issue status
	PullRequest  HookPr   `json:"pull_request"`
	Sender       HookUser //The user information that triggered the hook.
	Iid          int64   //issue Logo
	Number       int64
	Title        string //issue title
	Description  string //issue description
	State        string //issue status
	Url          string //issue URL on code cloud
	Body         string
	MergeStatus  string `json:"merge_status"`
	Repository   HookRepository
	TargetBranch string `json:"target_branch"`
}

func QueryTestOriginPr(top *TestOriginPr, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(top, field...)
	return err
}

// insert data
func InsertTestOriginPr(top *TestOriginPr) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(top)
	return id, err
}

func UpdateTestOriginPr(top *TestOriginPr, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(top, fields...)
	return err
}

func QueryEulerOriginPr(top *EulerOriginPr, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(top, field...)
	return err
}

// insert data
func InsertEulerOriginPr(top *EulerOriginPr) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(top)
	return id, err
}

func UpdateEulerOriginPr(top *EulerOriginPr, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(top, fields...)
	return err
}
