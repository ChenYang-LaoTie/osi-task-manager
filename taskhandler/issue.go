package taskhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup
var issueLock sync.Mutex

//OrgInfo
type OrgInfo struct {
	ID           int32  `json:"id,omitempty"`
	Login        string `json:"login,omitempty"`
	URL          string `json:"url,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	ReposURL     string `json:"repos_url,omitempty"`
	EventsURL    string `json:"events_url,omitempty"`
	MembersURL   string `json:"members_url,omitempty"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	Enterprise   string `json:"enterprise,omitempty"`
	Members      int64  `json:"members,omitempty"`
	PublicRepos  int64  `json:"public_repos,omitempty"`
	PrivateRepos int64  `json:"private_repos,omitempty"`
}

//Branch Get all branches
type Branch struct {
	Name          string `json:"name,omitempty"`
	Protected     bool   `json:"protected,omitempty"`
	ProtectionURL string `json:"protection_url,omitempty"`
}

//PackageInfo package info model
type PackageInfo struct {
	Code string
	Msg  string
	Data Info
}

//Info cve info
type Info struct {
	Description string
}
type GaussIssueUserRecordTp struct {
	UserId      int64
	OrId        int64
	IssueNumber string
	RepoPath    string
	Owner       string
	Status      int8
	Type        int8
}

func GetOriginIssue(owner, eulerToken string) error {
	logs.Info("Synchronize gitee's issue start......")
	orgInfo, err := GetOrgInfo(eulerToken, owner)
	if err != nil {
		logs.Error("GetOrgInfo, owner: ", owner, ",err: ", err)
		return err
	}
	reposNum := orgInfo.PublicRepos + orgInfo.PrivateRepos
	if reposNum <= 0 {
		logs.Info(fmt.Sprintf("%v contain %v repository,grab issue finish!", owner, reposNum))
		return errors.New(fmt.Sprintf("%v contain %v repository,grab issue finish!", owner, reposNum))
	}
	pageSize := reposNum / int64(perPage)
	if reposNum%int64(perPage) > 0 {
		pageSize = pageSize + 1
	}
	var i int64
	for i = 1; i <= pageSize; i++ {
		go GetOrgRepos(eulerToken, owner, i)
	}
	wg.Wait()
	logs.Info("Synchronize gitee's issue  finish...")
	return nil
}

//GrabIssueByRepo grab issue by repository
func GrabIssueByRepo(accToken, owner, repo, state string) {
	page := 1
	for {
		list, err := GetIssueList(accToken, owner, repo, state, page)
		if err != nil {
			logs.Error("GetIssueList, repo: ", repo, ",err: ", err)
			break
		}
		issueLock.Lock()
		handleIssueList(list)
		issueLock.Unlock()
		if len(list) < perPage {
			break
		}
		page++

	}
}

func handleIssueList(list []models.HookIssue) {
	if len(list) == 0 {
		return
	}
	//var gil []models.GiteOriginIssue
	for _, v := range list {
		issueTitle := common.TrimString(v.Title)
		issueType := common.TrimString(v.IssueType)
		issueNumber := common.TrimString(v.Number)
		repoPath := common.TrimString(v.Repository.Path)
		owner := common.TrimString(v.Repository.NameSpace.Path)
		if issueType == CIssueType || strings.HasPrefix(issueTitle, CIssueType) {
			eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath, IssueId: v.Id, IssueNumber: issueNumber}
			eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
			if eiErr != nil {
				CreateIssueOrgData(v, &eoi, 1)
				eId, orErr := models.InsertEulerOriginIssue(&eoi)
				if orErr != nil {
					logs.Error("InsertEulerOriginIssue, id: ", eId, ",err: ", orErr)
					continue
				}
			} else {
				updateStr := CreateIssueOrgData(v, &eoi, 2)
				upErr := models.UpdateEulerOriginIssue(&eoi, updateStr...)
				if upErr != nil {
					logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
					continue
				}
			}
		}
	}
}

func CreateIssueOrgData(hi models.HookIssue, eoi *models.EulerOriginIssue, flag int) []string {
	updateStr := make([]string, 0)
	issueState := common.TrimString(hi.State)
	issueZhState := common.TrimString(hi.IssueState)
	eoi.IssueState = issueState
	updateStr = append(updateStr, "IssueState")
	eoi.IssueStateName = issueZhState
	updateStr = append(updateStr, "IssueStateName")
	eoi.GitUrl = hi.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	eoi.IssueCreate = hi.User.UserName
	updateStr = append(updateStr, "IssueCreate")
	eoi.RepoUrl = hi.Repository.Url
	//updateStr = append(updateStr, "RepoUrl")
	eoi.IssueNumber = common.TrimString(hi.Number)
	eoi.IssueId = hi.Id
	eoi.RepoPath = hi.Repository.Path
	eoi.Owner = hi.Repository.NameSpace.Path
	eoi.Status = 1
	updateStr = append(updateStr, "Status")
	if len(hi.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.CreateTime = hi.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.UpdateTime = hi.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.FinishedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		eoi.FinishedTime = hi.FinishedAt.String()
		updateStr = append(updateStr, "FinishedTime")
	}
	labelStr := ""
	if hi.Labels != nil && len(hi.Labels) > 0 {
		for _, la := range hi.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
	}
	eoi.IssueLabel = labelStr
	updateStr = append(updateStr, "IssueLabel")
	eoi.IssueType = common.TrimString(hi.IssueType)
	updateStr = append(updateStr, "IssueType")
	eoi.Title = hi.Title
	updateStr = append(updateStr, "Title")
	vb := strings.ReplaceAll(hi.Body, "：", "：")
	eoi.IssueBody = vb
	updateStr = append(updateStr, "IssueBody")
	el := TaskScoreRegZh.FindAllStringSubmatch(hi.Body, -1)
	if len(el) > 0 && len(el[0]) > 1 {
		value := RegexpDigit.FindAllStringSubmatch(common.TrimString(el[0][1]), -1)
		if len(value[0]) > 0 && len(value[0]) > 1 {
			eoi.TaskScore, _ = strconv.Atoi(value[0][1])
			updateStr = append(updateStr, "TaskScore")
		}
	} else {
		dd := TaskScoreRegEn.FindAllStringSubmatch(hi.Body, -1)
		if len(dd) > 0 && len(dd[0]) > 0 {
			value := RegexpDigit.FindAllStringSubmatch(common.TrimString(dd[0][1]), -1)
			eoi.TaskScore, _ = strconv.Atoi(value[0][1])
			updateStr = append(updateStr, "TaskScore")
		}
	}
	if flag == 1 {
		eoi.GrabTime = common.GetCurTime()
	}
	eoi.IssueAssignee = hi.Assignee.UserName
	updateStr = append(updateStr, "IssueAssignee")
	return updateStr
}

//GetOrgInfo get  organization information
func GetOrgInfo(accToken, org string) (OrgInfo, error) {
	oi := OrgInfo{}
	resp, err := http.Get(fmt.Sprintf(GiteOrgInfoURL, org, accToken))
	if err != nil {
		return oi, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return oi, err
	}
	err = json.Unmarshal(body, &oi)
	return oi, err
}

//GetOrgRepos get organization repository
func GetOrgRepos(accToken, org string, page int64) {
	wg.Add(1)
	defer wg.Done()
	resp, err := http.Get(fmt.Sprintf(GiteOrgReposURL, org, accToken, page, perPage))
	if err != nil {
		logs.Error("Get, GiteOrgReposURL: ", GiteOrgReposURL, ", org: ", GiteOrgReposURL, ",err: ", err)
		return
	}
	defer resp.Body.Close()
	var reps []models.Repository
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("ReadAll, GiteOrgReposURL: ", GiteOrgReposURL, ", org: ", GiteOrgReposURL, ",err: ", err)
		return
	}
	//logs.Info("GetOrgRepos, body: ", string(body))
	err = json.Unmarshal(body, &reps)
	if err != nil {
		logs.Error("Unmarshal, GiteOrgReposURL: ", GiteOrgReposURL, ", org: ", GiteOrgReposURL, ",err: ", err)
		return
	}
	for _, v := range reps {
		GrabIssueByRepo(accToken, org, v.Name, "all")
	}
}

//GetIssueList get the repository issue list
func GetIssueList(accToken, owner, repo, state string, page int) (issueList []models.HookIssue, err error) {
	giteUrl := fmt.Sprintf(GiteRepoIssuesURL, owner, repo, accToken, state, page, perPage)
	resp, err := http.Get(giteUrl)
	if err != nil {
		logs.Error("Get, GiteRepoIssuesURL: ", giteUrl, ", repo: ", repo, ", err: ", err)
		return issueList, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("ReadAll, GiteRepoIssuesURL: ", giteUrl, ", repo: ", repo, ", err: ", err)
		return issueList, err
	}
	logs.Info("-----------issue list: ", string(body))
	err = json.Unmarshal(body, &issueList)
	if err != nil {
		logs.Error("Unmarshal, GiteRepoIssuesURL: ", giteUrl, ", repo: ", repo, ", err: ", err)
	}
	//logs.Info("++++++++++issueList: ", issueList)
	return
}

func AddHookIssue(issueData *models.IssuePayload) {
	issueTitle := common.TrimString(issueData.Issue.Title)
	issueType := common.TrimString(issueData.Issue.TypeName)
	issueNumber := common.TrimString(issueData.Issue.Number)
	repoPath := common.TrimString(issueData.Repository.Path)
	owner := common.TrimString(issueData.Repository.NameSpace)
	if issueType == CIssueType || strings.HasPrefix(issueTitle, CIssueType) {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueData.Issue.Id, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eiErr != nil {
			CreateHookIssueOrgData(issueData, &eoi, 1)
			eId, orErr := models.InsertEulerOriginIssue(&eoi)
			if orErr != nil {
				logs.Error("InsertEulerOriginIssue, id: ", eId, ",err: ", orErr)
				return
			}
		} else {
			updateStr := CreateHookIssueOrgData(issueData, &eoi, 2)
			upErr := models.UpdateEulerOriginIssue(&eoi, updateStr...)
			if upErr != nil {
				logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
				return
			}
		}
		eulerToken := common.GetEnvToken(issueData.Repository.NameSpace)
		if len(eulerToken) < 1 {
			logs.Error("namespace error: ", issueData.Repository.NameSpace)
			return
		}
		osiTaskLabel := beego.AppConfig.String("osi_task")
		EditLabel(repoPath, issueNumber, osiTaskLabel, "", eulerToken, owner, eoi)
	}
}

func CreateHookIssueOrgData(hi *models.IssuePayload, eoi *models.EulerOriginIssue, flag int) []string {
	updateStr := make([]string, 0)
	issueState := common.TrimString(hi.State)
	issueZhState := common.TrimString(hi.Issue.StateName)
	eoi.IssueState = issueState
	updateStr = append(updateStr, "IssueState")
	eoi.IssueStateName = issueZhState
	updateStr = append(updateStr, "IssueStateName")
	eoi.GitUrl = hi.Issue.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	eoi.IssueCreate = hi.User.UserName
	updateStr = append(updateStr, "IssueCreate")
	eoi.RepoUrl = hi.Repository.Url
	updateStr = append(updateStr, "RepoUrl")
	eoi.IssueNumber = common.TrimString(hi.Issue.Number)
	eoi.IssueId = hi.Issue.Id
	eoi.RepoPath = hi.Repository.Path
	eoi.Owner = hi.Repository.NameSpace
	eoi.Status = 1
	updateStr = append(updateStr, "Status")
	if len(hi.Issue.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.CreateTime = hi.Issue.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.Issue.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.UpdateTime = hi.Issue.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.Issue.FinishedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		eoi.FinishedTime = hi.Issue.FinishedAt.String()
		updateStr = append(updateStr, "FinishedTime")
	}
	osiTaskLabel := beego.AppConfig.String("osi_task")
	labelStr := ""
	if hi.Issue.Labels != nil && len(hi.Issue.Labels) > 0 {
		labelFlag := false
		tmpLabelList := make([]string, 0)
		for _, la := range hi.Issue.Labels {
			tmpLabelList = append(tmpLabelList, la.Name)
			if la.Name == osiTaskLabel {
				labelFlag = true
			}
		}
		if !labelFlag {
			tmpLabelList = append(tmpLabelList, osiTaskLabel)
		}
		if len(tmpLabelList) > 0 {
			labelStr = strings.Join(tmpLabelList, ",")
		}
	} else {
		if len(osiTaskLabel) > 1 {
			labelStr = osiTaskLabel
		}
	}
	eoi.IssueLabel = labelStr
	updateStr = append(updateStr, "IssueLabel")
	eoi.IssueType = common.TrimString(hi.Issue.TypeName)
	updateStr = append(updateStr, "IssueType")
	eoi.Title = hi.Issue.Title
	updateStr = append(updateStr, "Title")
	vb := strings.ReplaceAll(hi.Issue.Body, "：", "：")
	eoi.IssueBody = vb
	updateStr = append(updateStr, "IssueBody")
	el := TaskScoreRegZh.FindAllStringSubmatch(hi.Issue.Body, -1)
	if len(el) > 0 && len(el[0]) > 1 {
		value := RegexpDigit.FindAllStringSubmatch(common.TrimString(el[0][1]), -1)
		if len(value[0]) > 0 && len(value[0]) > 1 {
			eoi.TaskScore, _ = strconv.Atoi(value[0][1])
			updateStr = append(updateStr, "TaskScore")
		}
	} else {
		dd := TaskScoreRegEn.FindAllStringSubmatch(hi.Issue.Body, -1)
		if len(dd) > 0 && len(dd[0]) > 0 {
			value := RegexpDigit.FindAllStringSubmatch(common.TrimString(dd[0][1]), -1)
			eoi.TaskScore, _ = strconv.Atoi(value[0][1])
			updateStr = append(updateStr, "TaskScore")
		}
	}
	if flag == 1 {
		eoi.GrabTime = common.GetCurTime()
	}
	eoi.IssueAssignee = hi.Assignee.UserName
	eoi.GitId = hi.Assignee.Id
	updateStr = append(updateStr, "IssueAssignee")
	updateStr = append(updateStr, "GitId")
	logs.Info("eoi===>", eoi)
	return updateStr
}

func DelHookIssue(issueData *models.IssuePayload) {
	issueTitle := common.TrimString(issueData.Issue.Title)
	issueType := common.TrimString(issueData.Issue.TypeName)
	issueNumber := common.TrimString(issueData.Issue.Number)
	repoPath := common.TrimString(issueData.Repository.Path)
	owner := common.TrimString(issueData.Repository.NameSpace)
	if issueType == CIssueType || strings.HasPrefix(issueTitle, CIssueType) {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueData.Issue.Id, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId == 0 {
			logs.Error("DelHookIssue, Data does not exist, eiErr: ", eiErr)
			return
		} else {
			userId, delErr := models.DeleteEulerOriginIssueAll(&eoi)
			if delErr != nil {
				logs.Error("DeleteEulerOriginIssueAll, Data deletion failed, delErr: ", delErr)
				return
			}
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: eoi.IssueNumber,
				RepoPath: eoi.RepoPath, Owner: owner, Status: 12}
			EulerIssueUserRecord(et)
		}
	}
}

// Get community issues
func GetGitOriginIssue() error {
	ownerStr := beego.AppConfig.String("repo::owner")
	if len(ownerStr) < 1 {
		logs.Error("No community issues can be obtained")
		return errors.New("Invalid value")
	}
	nameSpaceList := strings.Split(ownerStr, ",")
	for _, owner := range nameSpaceList {
		eulerToken := common.GetEnvToken(owner)
		if len(eulerToken) < 1 {
			logs.Error("namespace error: ", owner)
			continue
		}
		ownerSlice := strings.Split(ownerStr, ",")
		for _, owner := range ownerSlice {
			orErr := GetOriginIssue(owner, eulerToken)
			if orErr != nil {
				logs.Error("Failed to get issue, owner: ", owner)
				continue
			}
		}
	}
	return nil
}

func AutoAddLabelTask() error {
	osiTaskLabel := beego.AppConfig.String("osi_task")
	eulerIssue := models.QueryOpenEulerIssueAll(0)
	if len(eulerIssue) > 0 {
		for _, ei := range eulerIssue {
			eulerToken := common.GetEnvToken(ei.Owner)
			if len(eulerToken) < 1 {
				logs.Error("namespace error: ", ei.Owner)
				continue
			}
			_, allLabelList := QueryIssueLabels(eulerToken, ei.RepoPath, ei.IssueNumber, ei.Owner)
			if len(allLabelList) > 0 {
				repLabelList := make([]string, 0)
				labelFlag := false
				interFlag := false
				for _, lab := range allLabelList {
					if strings.Contains(lab, "sig/") {
						slm := models.SigLabelMapping{EulerLabel: lab}
						queryErr := models.QuerySigLabelMapping(&slm, "EulerLabel")
						if slm.Id < 1 {
							logs.Error("sig tag has no configuration information: ", lab, queryErr)
							repLabelList = append(repLabelList, lab)
						} else {
							if !labelFlag {
								repLabelList = append(repLabelList, slm.InternLabel)
								labelFlag = true
							} else {
								repLabelList = append(repLabelList, lab)
							}
						}
					} else {
						if lab == osiTaskLabel {
							interFlag = true
						}
						repLabelList = append(repLabelList, lab)
					}
				}
				if !interFlag {
					repLabelList = append(repLabelList, osiTaskLabel)
				}
				repLabelSlice := make([]string, 0)
				if len(repLabelList) > 0 {
					sort.Strings(repLabelList)
					repSlice := common.SliceRemoveDup(repLabelList)
					for _, rep := range repSlice {
						repLabelSlice = append(repLabelSlice, rep.(string))
					}
				}
				labels := ""
				if len(repLabelSlice) > 0 {
					labels = strings.Join(repLabelSlice, ",")
				}
				if len(labels) > 1 {
					if labelFlag {
						UpdateIssueLabels(eulerToken, ei.RepoPath, ei.IssueNumber, ei.Owner, labels)
						ei.LabelFlag = 2
						ei.IssueLabel = labels
						upErr := models.UpdateEulerOriginIssue(&ei, "IssueLabel", "LabelFlag")
						if upErr != nil {
							logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
						}
					} else {
						if !interFlag {
							UpdateIssueLabels(eulerToken, ei.RepoPath, ei.IssueNumber, ei.Owner, labels)
							ei.IssueLabel = labels
							upErr := models.UpdateEulerOriginIssue(&ei, "IssueLabel")
							if upErr != nil {
								logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
							}
						}
					}
				}
			} else {
				UpdateIssueLabels(eulerToken, ei.RepoPath, ei.IssueNumber, ei.Owner, osiTaskLabel)
				ei.IssueLabel = osiTaskLabel
				upErr := models.UpdateEulerOriginIssue(&ei, "IssueLabel")
				if upErr != nil {
					logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
				}
			}
		}
	}
	return nil
}
