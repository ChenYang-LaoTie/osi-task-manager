package taskhandler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"net/http"
	"os"
	"osi-task-manager/common"
	"osi-task-manager/models"
	"path/filepath"
	"strconv"
)

type ExcelFileInfo struct {
	RemoteFileName string
	ExcelOwner     string
	ExcelRepo      string
	AccessToken    string
	LocalDir       string
}

func DownloadFile() error {
	dir := beego.AppConfig.DefaultString("path_file", "excel")
	stdExcelPath := beego.AppConfig.String("excel::std_excel_path")
	tutorFilePath := beego.AppConfig.String("excel::tutor_file_path")
	excelOwner := beego.AppConfig.String("excel::excel_owner")
	excelRepo := beego.AppConfig.String("excel::excel_repo")
	accessToken := os.Getenv("GITEE_SRC_TOKEN")
	common.CreateDir(dir)
	stdInfo := ExcelFileInfo{LocalDir: dir, RemoteFileName: stdExcelPath,
		ExcelOwner: excelOwner, ExcelRepo: excelRepo, AccessToken: accessToken}
	stdErr := DownloadExcel(stdInfo)
	if stdErr != nil {
		logs.Error("DownloadExcel, stdErr: ", stdErr)
	} else {
		stdFileName := filepath.Join(stdInfo.LocalDir, stdInfo.RemoteFileName)
		excelData := ExcelParse(stdFileName)
		if len(excelData) > 0 {
			ProcStdData(excelData, accessToken)
		}
	}
	tutInfo := ExcelFileInfo{LocalDir: dir, RemoteFileName: tutorFilePath,
		ExcelOwner: excelOwner, ExcelRepo: excelRepo, AccessToken: accessToken}
	tutErr := DownloadExcel(tutInfo)
	if tutErr != nil {
		logs.Error("tutErr: ", tutErr)
	} else {
		tutFileName := filepath.Join(tutInfo.LocalDir, tutInfo.RemoteFileName)
		excelData := ExcelParse(tutFileName)
		if len(excelData) > 0 {
			ProcTutData(excelData, accessToken)
		}
	}
	return nil
}

// Process the parsed data
func ProcStdData(excelData [][]string, accessToken string) {
	for _, ed := range excelData {
		su := models.StdUser{}
		gitUserId := ""
		emailAddr := ""
		statusType := 1
		if len(ed) <= 1 {
			logs.Error("data err :", ed)
			continue
		} else if len(ed) == 2 {
			gitUserId = ed[0]
			statusType, _ = strconv.Atoi(ed[1])
		} else {
			gitUserId = ed[0]
			emailAddr = ed[1]
			statusType, _ = strconv.Atoi(ed[2])
		}
		gui := GetGitId(accessToken, gitUserId)
		if gui.GitId == 0 {
			logs.Error("GetGitId, Failed to get gitid information: ", gui)
			continue
		}
		su.GitId = gui.GitId
		su.GitUserId = gitUserId
		switch statusType {
		case 3:
			delErr := models.DelStdUser(&su, "GitId")
			if delErr != nil {
				logs.Error("DelStdUser, delErr: ", delErr)
			}
		default:
			querErr := models.QueryStdUser(&su, "GitId")
			if su.UserId == 0 {
				su.EmailAddr = emailAddr
				su.GitId = gui.GitId
				su.GitUserId = gui.Login
				su.CreateTime = common.GetCurTime()
				su.Status = 1
				su.UserName = gui.Login
				inNum, inErr := models.InsertStdUser(&su)
				if inNum > 0 {
					logs.Info("InsertStdUser success, inNum:", inNum)
				} else {
					logs.Error("InsertStdUser failed, inNum: ", inNum, ", inErr", inErr, querErr)
				}
			} else {
				su.EmailAddr = emailAddr
				su.UpdateTime = common.GetCurTime()
				su.Status = 1
				su.GitUserId = gui.Login
				su.UserName = gui.Login
				upErr := models.UpdateStdUser(&su, "EmailAddr", "UpdateTime", "Status", "GitUserId")
				if upErr == nil {
					logs.Info("UpdateStdUser success, upErr:", upErr)
				} else {
					logs.Error("UpdateStdUser failed, upErr: ", upErr)
				}
			}
		}
	}
}

// Process the parsed data
func ProcTutData(excelData [][]string, accessToken string) {
	for _, ed := range excelData {
		tu := models.TutUser{}
		gitUserId := ""
		emailAddr := ""
		statusType := 1
		if len(ed) <= 1 {
			logs.Error("data err :", ed)
			continue
		} else if len(ed) == 2 {
			gitUserId = ed[0]
			statusType, _ = strconv.Atoi(ed[1])
		} else {
			gitUserId = ed[0]
			emailAddr = ed[1]
			statusType, _ = strconv.Atoi(ed[2])
		}
		tu.GitUserId = gitUserId
		gui := GetGitId(accessToken, gitUserId)
		if gui.GitId == 0 {
			logs.Error("GetGitId, Failed to get gitid information: ", gui)
			continue
		}
		tu.GitId = gui.GitId
		switch statusType {
		case 3:
			delErr := models.DelTutUser(&tu, "GitId")
			if delErr != nil {
				logs.Error("DelTutUser, delErr: ", delErr)
			}
		default:
			querErr := models.QueryTutUser(&tu, "GitId")
			if tu.UserId == 0 {
				tu.EmailAddr = emailAddr
				tu.GitId = gui.GitId
				tu.GitUserId = gui.Login
				tu.CreateTime = common.GetCurTime()
				tu.Status = 1
				tu.UserName = gui.Login
				inNum, inErr := models.InsertTutUser(&tu)
				if inNum > 0 {
					logs.Info("InsertTutUser success, inNum:", inNum)
				} else {
					logs.Error("InsertTutUser failed, inNum: ", inNum, ", inErr", inErr, querErr)
				}
			} else {
				tu.EmailAddr = emailAddr
				tu.UpdateTime = common.GetCurTime()
				tu.Status = 1
				tu.GitUserId = gui.Login
				tu.UserName = gui.Login
				upErr := models.UpdateTutUser(&tu, "EmailAddr", "UpdateTime", "Status", "GitUserId")
				if upErr == nil {
					logs.Info("UpdateTutUser success, upErr:", upErr)
				} else {
					logs.Error("UpdateTutUser failed, upErr: ", upErr)
				}
			}
		}
	}
}

func ExcelParse(fileName string) [][]string {
	xlFile, fErr := xlsx.OpenFile(fileName)
	if fErr != nil {
		logs.Error("fErr: ", fErr)
		return nil
	}
	length := len(xlFile.Sheets[0].Rows)
	resourceArr := make([][]string, length-1)
	for _, sheet := range xlFile.Sheets {
		for rowIndex, row := range sheet.Rows {
			if rowIndex == 0 {
				continue
			}
			cellList := make([]string, 0)
			for _, cell := range row.Cells {
				text := cell.String()
				cellList = append(cellList, text)
			}
			resourceArr = append(resourceArr, cellList)
		}
	}
	return resourceArr
}

// Download excel file
func DownloadExcel(efi ExcelFileInfo) error {
	localPath := filepath.Join(efi.LocalDir, efi.RemoteFileName)
	excePath := beego.AppConfig.String("excel::excel_path")
	remotePath := fmt.Sprintf("%v/%v", excePath, efi.RemoteFileName)
	gitUrl := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/contents/%v?access_token=%v",
		efi.ExcelOwner, efi.ExcelRepo, remotePath, efi.AccessToken)
	logs.Info("DownloadExcel, gitUrl: ", gitUrl)
	resp, err := http.Get(gitUrl)
	if err != nil {
		logs.Error("DownloadExcel, error: ", err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error(err)
		return err
	}
	//logs.Info("body: \n", string(body), "url: ", gitUrl)
	var contents map[string]interface{}
	err = json.Unmarshal(body, &contents)
	if err != nil {
		logs.Error(err)
		return err
	}
	if contents == nil || contents["type"] == nil {
		logs.Error("contents is nil or contents[type] is nil ", contents["type"])
		return errors.New("contents is nil")
	}
	f, ferr := os.Create(localPath)
	if ferr != nil {
		logs.Error(ferr)
		return ferr
	}
	fileType := contents["type"].(string)
	encoding := contents["encoding"].(string)
	content := contents["content"].(string)
	if fileType == "file" && encoding == "base64" {
		data, baseErr := base64.StdEncoding.DecodeString(content)
		if baseErr == nil {
			f.Write(data)
		}
	} else {
		f.WriteString(content)
	}
	return nil
}
