package taskhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"osi-task-manager/util"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
)

const sheetPrName = "issue_pr_count"

func createStatistExcel(excelPath string, weekList [][]string) string {
	excelTitle := beego.AppConfig.String("excel_title")
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheetPrName)
	sheetTileMap := make(map[string]string)
	sheetTileMap["A1"] = "组件名称"
	sheetTileMap["B1"] = "组件对应的issue的url"
	sheetTileMap["C1"] = "组件对应的issue编号"
	for i, wl := range weekList {
		if len(excelTitle) >= i {
			inKey := string(excelTitle[i+3]) + "1"
			sheetTileMap[inKey] = wl[0] + "_" + wl[1] + "(issue数量, pr数量, 总和)"
		}
	}
	for k, v := range sheetTileMap {
		xlsx.SetCellValue(sheetPrName, k, v)
	}
	xlsx.SetActiveSheet(index)
	err := xlsx.SaveAs(excelPath)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return excelPath
}

func addIssueExcelData(ev []string) []interface{} {
	cveData := make([]interface{}, 0)
	for _, e := range ev {
		cveData = append(cveData, e)
	}
	return cveData
}

func ReadWriteIssueExcel(excelPath string, evlChan [][]string) error {
	file, openErr := excelize.OpenFile(excelPath)
	if openErr != nil {
		logs.Error("fail to open the file, excelPath: ", excelPath, ", openErr: ", openErr)
		return openErr
	}
	for _, ev := range evlChan {
		excelData := addIssueExcelData(ev)
		if len(excelData) > 0 {
			rows, sheetErr := file.GetRows(sheetPrName)
			if sheetErr != nil {
				logs.Error(sheetErr)
			}
			idx := len(rows) + 1
			axis := fmt.Sprintf("A%d", idx)
			setErr := file.SetSheetRow(sheetPrName, axis, &excelData)
			if setErr != nil {
				logs.Error("setErr: ", setErr)
			}
		}
	}
	fileErr := file.SaveAs(excelPath)
	if fileErr != nil {
		logs.Error("Failed to save file, ", fileErr)
	}
	return fileErr
}

func StatistIssueCommentCount(createdAt string, weekList [][]string, commentMap map[string]int) {
	ct := int64(0)
	if len(createdAt) > 0 {
		if len(createdAt) > 19 {
			ct = util.TimeStrToInt(createdAt[:19], "2006-01-02T15:04:05")
		} else {
			ct = util.TimeStrToInt(createdAt, "2006-01-02T15:04:05")
		}
	}
	for _, wl := range weekList {
		mapKey := wl[0] + "_" + wl[1]
		st := util.TimeStrToInt(wl[0], "2006-01-02 15:04:05")
		et := util.TimeStrToInt(wl[1], "2006-01-02 15:04:05")
		if st <= ct && ct <= et {
			commentMap[mapKey] += 1
		}
	}
}

func GetIssueComments(owner, repo, accessToken, issueNum string, page, perPage int, weekList [][]string, commentMap map[string]int) {
	localPage := page
	localPerPage := perPage

	for _, wl := range weekList {
		mapKey := wl[0] + "_" + wl[1]
		commentMap[mapKey] = 0
	}
	for {
		url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/issues/%v/comments?access_token=%v&page=%v&per_page=%v&order=asc",
			owner, repo, issueNum, accessToken, localPage, localPerPage)
		issueCommentData, err := util.HTTPGet(url)
		if err == nil && issueCommentData != nil && len(issueCommentData) > 0 {
			localPage += 1
			for _, value := range issueCommentData {
				if _, ok := value["id"]; !ok {
					logs.Error("issueCommentData, err: ", ok, "url: ", url)
					continue
				}
				if _, ok := value["created_at"]; !ok {
					logs.Error("created_at, err: ", ok, "url: ", url)
					continue
				}
				createdAt := value["created_at"].(string)
				StatistIssueCommentCount(createdAt, weekList, commentMap)
			}
		} else {
			break
		}
	}
}

func GetIssuePrComments(owner, repo, accessToken, prNumber string, page, perPage int, weekList [][]string, commentMap map[string]int) {
	localPage := page
	localPerPage := perPage

	for _, wl := range weekList {
		mapKey := wl[0] + "_" + wl[1]
		commentMap[mapKey] = 0
	}
	for {
		url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/pulls/%v/comments?access_token=%v&page=%v&per_page=%v&direction=asc",
			owner, repo, prNumber, accessToken, localPage, localPerPage)
		prCommentData, err := util.HTTPGet(url)
		if err == nil && prCommentData != nil && len(prCommentData) > 0 {
			localPage += 1
			for _, value := range prCommentData {
				if _, ok := value["id"]; !ok {
					logs.Error("prCommentData, err: ", ok, "url: ", url)
					continue
				}
				if _, ok := value["created_at"]; !ok {
					logs.Error("created_at, err: ", ok, "url: ", url)
					continue
				}
				createdAt := value["created_at"].(string)
				StatistIssueCommentCount(createdAt, weekList, commentMap)
			}
		} else {
			break
		}
	}
}

func getRepoIssueAllPR(token, owner, repo, issueNum string, weekList [][]string, commentMap map[string]int) {
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/issues/%v/pull_requests", owner, issueNum)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logs.Error("NewRequest, url: ", url, ",err: ", err)
		return
	}
	q := req.URL.Query()
	q.Add("access_token", token)
	q.Add("repo", repo)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.Error("DefaultClient, url: ", url, ",err: ", err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		issuePr := make([]map[string]interface{}, 0)
		read, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logs.Error("ReadAll, url: ", url, ",err: ", err)
			return
		}
		resp.Body.Close()
		err = json.Unmarshal(read, &issuePr)
		if err != nil {
			logs.Error("Unmarshal, url: ", url, ",err: ", err)
			return
		}
		for _, v := range issuePr {
			if _, ok := v["id"]; !ok {
				continue
			}
			prNumber := int64(v["number"].(float64))
			prNumberStr := strconv.FormatInt(prNumber, 10)
			GetIssuePrComments(owner, repo, token, prNumberStr, 1, 20, weekList, commentMap)
		}
	} else {
		resp.Body.Close()
	}
	return
}

func EulerIssueStatistics() error {
	fileName := ""
	dir := beego.AppConfig.String("path_file")
	eulerToken := beego.AppConfig.String("repo::git_token")
	GameStartTime := beego.AppConfig.String("game_start_time")
	curDate := common.GetCurDate()
	weekList := ChangeToWeek(GameStartTime[:10], curDate)
	logs.Info("weekList: ", weekList)
	// File storage directory
	CreateDir(dir)
	totalFileSlice := make([]string, 0)
	// Get user information
	eulerIssue := models.QueryOpenEulerIssueAll()
	if len(eulerIssue) > 0 {
		fileName = "osi_openEuler_issue_pr"
		fileName = fileName + "_" + GameStartTime[:10] + "_" + curDate + ".xlsx"
		fileName = filepath.Join(dir, fileName)
		fileExcelPath := createStatistExcel(fileName, weekList)
		if fileExcelPath == "" {
			logs.Error("Failed to create file")
			return errors.New("Failed to create file")
		}
		weekExcelValue := make([][]string, 0)
		totalCommentMap := make(map[string]int, 0)
		for _, ei := range eulerIssue {
			issueCommentMap := make(map[string]int, 0)
			prCommentMap := make(map[string]int, 0)
			excelList := []string{}
			GetIssueComments(ei.Owner, ei.RepoPath, eulerToken, ei.IssueNumber, 1, 20, weekList, issueCommentMap)
			getRepoIssueAllPR(eulerToken, ei.Owner, ei.RepoPath, ei.IssueNumber, weekList, prCommentMap)
			excelList = append(excelList, ei.RepoPath)
			excelList = append(excelList, ei.GitUrl)
			excelList = append(excelList, ei.IssueNumber)
			for _, wl := range weekList {
				mapKey := wl[0] + "_" + wl[1]
				excelList = append(excelList,
					strconv.Itoa(issueCommentMap[mapKey])+","+strconv.Itoa(prCommentMap[mapKey])+","+
						strconv.Itoa(issueCommentMap[mapKey]+prCommentMap[mapKey]))
				if _, ok := totalCommentMap[mapKey]; !ok {
					totalCommentMap[mapKey] = issueCommentMap[mapKey] + prCommentMap[mapKey]
				} else {
					totalCommentMap[mapKey] = totalCommentMap[mapKey] + issueCommentMap[mapKey] + prCommentMap[mapKey]
				}
			}
			weekExcelValue = append(weekExcelValue, excelList)
		}
		excelList := []string{}
		excelList = append(excelList, "issue + pr 总数")
		excelList = append(excelList, " ")
		excelList = append(excelList, " ")
		for _, wl := range weekList {
			mapKey := wl[0] + "_" + wl[1]
			excelList = append(excelList, strconv.Itoa(totalCommentMap[mapKey]))
		}
		weekExcelValue = append(weekExcelValue, excelList)
		issErr := ReadWriteIssueExcel(fileExcelPath, weekExcelValue)
		if issErr != nil {
			logs.Error("issErr: ", issErr)
			return issErr
		}

		totalFileSlice = append(totalFileSlice, fileName)
		SendIssueStatistExcel(fileName, GameStartTime[:10], curDate)
		DelFile(totalFileSlice)
	}
	return nil
}

func SendIssueStatistExcel(fileName, startTime, endTime string) {
	cBody := fmt.Sprintf("hi all: \r\n 附件中为excel文件: openEuler开源实习-统计赛题issue的评论和issue关联pr的评论的数量（%s~%s）, 请查收. \r\n", startTime, endTime)
	subject := fmt.Sprintf("openEuler开源实习-统计赛题issue的评论和issue关联pr的评论的数量", startTime, endTime)
	sendError := SendEmail(fileName, 1, cBody, subject)
	if sendError != nil {
		logs.Error("SendEmail, sendErr: ", sendError)
	}
	return
}
