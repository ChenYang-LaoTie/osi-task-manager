package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"time"
)

type LoginUserToken struct {
	UerId          int       `orm:"pk;auto;column(user_id)"`
	UserName       string    `orm:"size(64);column(user_name)"`
	PassWord       string    `orm:"size(256);column(pass_word)"`
	AesKey         string    `orm:"size(512);colnum(aes_key)"`
	ExpirationTime string    `orm:"size(32);column(expiration_time)" description:"token的过期时间"`
	CreateTime     time.Time `orm:"auto_now;type(datetime);column(create_time)"`
	UpdateTime     time.Time `orm:"auto_now;type(datetime);column(update_time)"`
	DeleteTime     time.Time `orm:"auto_now;type(datetime);column(delete_time)"`
}

type StdUser struct {
	UserId     int64  `orm:"pk;auto;column(user_id)"`
	UserName   string `orm:"size(256);column(user_name);null"`
	GitId      int64  `orm:"column(git_id);unique" description:"git内部id"`
	GitUserId  string `orm:"size(512);column(git_login);unique"`
	EmailAddr  string `orm:"size(256);column(email_address)"`
	Status     int8   `orm:"default(1);column(status)" description:"1:正常用户;2:异常用户;3:用户已无法领取任务"`
	TestFlag   int8   `orm:"default(1);column(test_flag)" description:"1:未完成测试任务;2:已完成测试任务"`
	CreateTime string `orm:"size(32);column(create_time);"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type TutUser struct {
	UserId     int64  `orm:"pk;auto;column(user_id)"`
	UserName   string `orm:"size(256);column(user_name);null"`
	GitId      int64  `orm:"column(git_id);unique" description:"git内部id"`
	GitUserId  string `orm:"size(512);column(git_login);unique"`
	EmailAddr  string `orm:"size(256);column(email_address)"`
	Status     int8   `orm:"default(1);column(status)" description:"1:正常用户;2:异常用户;3:用户已无法领取任务"`
	CreateTime string `orm:"size(32);column(create_time);"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

// Individual claiming the issue task pool
type EulerBlackUser struct {
	Id         int64  `orm:"pk;auto;column(id)"`
	UserId     int64  `orm:"unique;column(user_id)"`
	UserName   string `orm:"size(256);column(user_name);null"`
	GitId      int64  `orm:"column(git_id);unique" description:"git内部id"`
	GitUserId  string `orm:"size(512);column(git_login);unique"`
	EmailAddr  string `orm:"size(256);column(email_address)"`
	Status     int8   `orm:"default(1);column(status)" description:"1:待释放任务；2:已释放任务"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type EulerUnassignUser struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"unique;column(user_id)"`
	GitId        int64  `orm:"column(git_id);unique" description:"git内部id"`
	GitUserId    string `orm:"size(512);column(git_login);unique"`
	CountValue   int8   `orm:"default(0);column(count_val)" description:"用户每个月领取任务取消次数"`
	UnassignTime string `orm:"size(32);column(unassign_time);null" description:"冻结结束时间"`
	ReleaseFlag  int8   `orm:"default(1);column(release_flag)" description:"1:主动释放; 2:被动释放"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

//TestOriginPr pr that already exist on Code Cloud
type TestOriginPr struct {
	OrId         int64  `orm:"pk;auto;column(or_id)"`
	PrId         int64  `orm:"column(Pr_id);unique" description:"pr id,gitee上唯一"`
	GitUrl       string `orm:"column(git_url);size(512)" description:"issue gitee 链接"`
	PrNumber     int64  `orm:"column(pr_num)" description:"pr 编号"`
	PrState      string `orm:"column(pr_state);size(50)" description:"pr 状态"`
	Title        string `orm:"column(issue_title);type(text);null" description:"issue 标题"`
	PrBody       string `orm:"column(pr_body);null;type(text)" description:"pr 主体"`
	PrLabel      string `orm:"size(512);column(pr_label)" description:"pr标签"`
	PrCreate     string `orm:"column(pr_create);size(256)" description:"pr 创建人"`
	PrUpdate     string `orm:"column(pr_update);size(256)" description:"pr 更新人"`
	PrAssignee   string `orm:"column(pr_assignee);size(256)" description:"pr 责任人,必填"`
	GitId        int64  `orm:"column(git_id)" description:"git内部id"`
	UserId       int64  `orm:"index;column(user_id)"`
	RepoPath     string `orm:"column(pr_repo);size(512)" description:"仓库空间地址"`
	RepoUrl      string `orm:"column(repo_url);type(text)" description:"仓库码云地址链接"`
	Owner        string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	TargetBranch string `orm:"column(target_branch);size(64)" description:"pr提交的分支"`
	Status       int8   `orm:"default(0);column(status)" description:"0:提交;1:合并;2:删除"`
	CreateTime   string `orm:"size(32);column(create_time);null" description:"issue创建时间"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	ClosedTime   string `orm:"size(32);column(closed_time);null"`
	MergedTime   string `orm:"size(32);column(merged_time);null"`
	GrabTime     string `orm:"size(32);column(grab_time)" description:"记录当前pr的时间"`
}

//GiteOriginIssue Issues that already exist on Code Cloud
type EulerOriginIssue struct {
	OrId           int64  `orm:"pk;auto;column(or_id)"`
	IssueId        int64  `orm:"column(issue_id);unique" description:"issue id,gitee上唯一"`
	GitUrl         string `orm:"column(git_url);size(512)" description:"issue gitee 链接"`
	IssueNumber    string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	IssueState     string `orm:"column(issue_state);size(50)" description:"issue 状态"`
	IssueType      string `orm:"column(issue_type);size(64)" description:"issue 类型"`
	Title          string `orm:"column(issue_title);type(text);null" description:"issue 标题"`
	IssueBody      string `orm:"column(issue_body);null;type(text)" description:"issue 主体"`
	IssueLabel     string `orm:"size(512);column(issue_label)" description:"issue标签"`
	IssueCreate    string `orm:"column(issue_create);size(256)" description:"issue issue创建人"`
	IssueAssignee  string `orm:"column(issue_assignee);size(256)" description:"issue 责任人,必填"`
	GitId          int64  `orm:"column(git_id)" description:"git内部id"`
	RepoPath       string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	RepoUrl        string `orm:"column(repo_url);type(text)" description:"仓库码云地址链接"`
	Owner          string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status         int8   `orm:"default(1);column(status)" description:"1:正常;2:已删除"`
	IssueStateName string `orm:"size(50);column(issue_state_name)" description:"issue 中文状态"`
	TaskScore      int    `orm:"default(0);column(task_score)" description:"任务分值"`
	CreateTime     string `orm:"size(32);column(create_time);null" description:"issue创建时间"`
	UpdateTime     string `orm:"size(32);column(update_time);null"`
	FinishedTime   string `orm:"size(32);column(finished_time);null"`
	DeleteTime     string `orm:"size(32);column(delete_time);null"`
	GrabTime       string `orm:"size(32);column(grab_time)" description:"记录当前issue抓取的时间"`
	RecordCount    int    `orm:"default(1);column(record_count)" description:"记录题目释放次数"`
	LabelFlag      int8   `orm:"default(0);column(label_flag)" description:"1:未打sig标签;2:已打sig标签"`
}

// sig label mapping
type SigLabelMapping struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	EulerLabel  string `orm:"column(euler_label);size(512);unique" description:"原来已有sig标签"`
	InternLabel string `orm:"column(intern_label);size(512)" description:"开源实习sig标签"`
	CreateTime  string `orm:"size(32);column(create_time)"`
}

// Individual claiming the issue task pool
type EulerIssueUser struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	UserId      int64  `orm:"column(user_id)"`
	OrId        int64  `orm:"unique;column(or_id)"`
	TutId       int64  `orm:"column(tut_id)" description:"导师用户id"`
	IssueNumber string `orm:"column(issue_num);size(50)" description:"issue 编号"`
	RepoPath    string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner       string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	SendEmail   int8   `orm:"default(1);column(send_email)" description:"1:未发送邮件; 2:已发送邮件"`
	Status      int8   `orm:"default(1);column(status)" description:"1:申请认领中;2:导师审核通过;3:任务已提交;4:已完成"`
	PointFlag   int8   `orm:"default(1);column(point_flag)" description:"1: 未获得积分;2:已获得积分"`
	AssignTime  string `orm:"size(32);column(assign_time);null" description:"赛题认领超时日期"`
	FreezeTime  string `orm:"size(32);column(freeze_time);null" description:"认领赛题，导师未确认到期时期"`
	CreateTime  string `orm:"size(32);column(create_time)"`
	UpdateTime  string `orm:"size(32);column(update_time);null"`
	DeleteTime  string `orm:"size(32);column(delete_time);null"`
}

type EulerUserClaimRecord struct {
	Id         int64  `orm:"pk;auto;column(id)"`
	UserId     int64  `orm:"column(user_id)"`
	OrId       int64  `orm:"column(or_id)"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type EulerIssueUserRecord struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	UserId      int64  `orm:"index;column(user_id)"`
	OrId        int64  `orm:"index;column(or_id)"`
	IssueNumber string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	RepoPath    string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner       string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status int8 `orm:"default(1);column(status)" description:"1:已认领;2:取消认领;3:完成提交;4:审核通过;
						5:认领超额; 6:重复认领; 7: 已被他人认领;8:提交他人任务; 9:重复完成提交; 
						10: 他人取消任务, 11: 任务完成,取消任务失败,
						12: 删除issue, 13: 黑名单用户认领失败, 14: 主动取消任务次数已达上线,不能再次认领任务, 
						15: 导师审核通过,16:导师审核不通过, 17:处理失败,18:导师未审核任务,直接提交任务, 
						19: 被动取消任务次数已达上线,不能再次认领任务, 20: 导师拒绝给学生分数,任务被释放,
						21: 超时系统主动取消认领"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type EulerIssueUserComplate struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id)"`
	OrId          int64  `orm:"unique;column(or_id)"`
	IssueNumber   string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	RepoPath      string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner         string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status        int8   `orm:"default(1);column(status)" description:"1:已完成"`
	IntegralValue int64  `orm:"column(integral_value)" description:"已获得多少积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
	UpdateTime    string `orm:"size(32);column(update_time);null"`
	DeleteTime    string `orm:"size(32);column(delete_time);null"`
}

type EulerUserIntegCount struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"unique;column(user_id)"`
	IntegralValue int64  `orm:"column(integral_value)" description:"用户获得总的积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
}

type EulerUserIntegDetail struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id)"`
	OrId          int64  `orm:"unique;column(or_id)"`
	IntegralValue int64  `orm:"column(integral_value)" description:"获得的积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
}

type EulerOriginPr struct {
	OrId         int64  `orm:"pk;auto;column(or_id)"`
	PrId         int64  `orm:"column(Pr_id);unique" description:"pr id,gitee上唯一"`
	GitUrl       string `orm:"column(git_url);size(512)" description:"issue gitee 链接"`
	PrNumber     int64  `orm:"column(pr_num)" description:"pr 编号"`
	IssueNumber  string `orm:"column(issue_num);size(50);null" description:"issue 编号"`
	PrState      string `orm:"column(pr_state);size(50)" description:"pr 状态"`
	Title        string `orm:"column(issue_title);type(text);null" description:"issue 标题"`
	PrBody       string `orm:"column(pr_body);null;type(text)" description:"pr 主体"`
	PrLabel      string `orm:"size(512);column(pr_label)" description:"pr标签"`
	PrCreate     string `orm:"column(pr_create);size(256)" description:"pr 创建人"`
	PrUpdate     string `orm:"column(pr_update);size(256)" description:"pr 更新人"`
	PrAssignee   string `orm:"column(pr_assignee);size(256)" description:"pr 责任人,必填"`
	GitId        int64  `orm:"column(git_id)" description:"git内部id"`
	RepoPath     string `orm:"column(pr_repo);size(512)" description:"仓库空间地址"`
	RepoUrl      string `orm:"column(repo_url);type(text)" description:"仓库码云地址链接"`
	Owner        string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	UserId       int64  `orm:"column(user_id)"`
	TargetBranch string `orm:"column(target_branch);size(64)" description:"pr提交的分支"`
	Status       int8   `orm:"default(0);column(status)" description:"0:提交;1:合并;2:删除"`
	CreateTime   string `orm:"size(32);column(create_time);null" description:"issue创建时间"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	ClosedTime   string `orm:"size(32);column(closed_time);null"`
	MergedTime   string `orm:"size(32);column(merged_time);null"`
	GrabTime     string `orm:"size(32);column(grab_time)" description:"记录当前pr的时间"`
}

type EmailList struct {
	Id        int64  `orm:"pk;auto"`
	EmailName string `orm:"size(512);column(email_name);null" description:"收件人名称"`
	EmailType int8   `orm:"column(email_type);default(1)" description:"1:openEuler;2:openGauss;3:都发"`
	SendType  int8   `orm:"column(send_type);default(1)" description:"1:收件人;2:抄送人"`
}

func CreateDb() bool {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("config init error:", err)
		return false
	}
	prefix := BConfig.String("mysql::dbprefix")
	InitdbType, _ := beego.AppConfig.Int("initdb")
	if InitdbType == 1 {
		orm.RegisterModelWithPrefix(prefix, new(TutUser), new(StdUser),
			new(TestOriginPr), new(LoginUserToken),
			new(EulerOriginIssue), new(EulerIssueUser),
			new(EulerIssueUserRecord), new(EulerIssueUserComplate),
			new(EulerUserIntegCount), new(EulerUserIntegDetail),
			new(EulerUnassignUser), new(EulerUserClaimRecord),
			new(EulerBlackUser), new(EulerOriginPr),
			new(EmailList), new(SigLabelMapping),
		)
		logs.Info("table create success!")
		errosyn := orm.RunSyncdb("default", false, true)
		if errosyn != nil {
			logs.Error(errosyn)
		}
	}
	return true
}
