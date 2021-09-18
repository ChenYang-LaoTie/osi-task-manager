package taskhandler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"gopkg.in/gomail.v2"
	"io"
	"io/ioutil"
	"mime"
	"net/smtp"
	"os"
	"osi-task-manager/models"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const sheetName = "user_points_list"

type ExcelValue struct {
	CurIndex          int
	GitLogin          string
	GitEmail          string
	Points            int64
	IssueCount        int64
	CommitIssueCount  int64
	CommentIssueCount int64
	CommitPrCount     int64
	Decription        string
}

type Mail interface {
	Auth()
	Send(message Message) error
}

type SendMail struct {
	user     string
	password string
	host     string
	port     string
	auth     smtp.Auth
}

type Attachment struct {
	name        string
	contentType string
	withFile    bool
}

type Message struct {
	from        string
	to          []string
	cc          []string
	bcc         []string
	subject     string
	body        string
	contentType string
	attachment  Attachment
}

func createExcel(excelPath string) string {
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheetName)
	sheetTileMap := make(map[string]string)
	sheetTileMap["A1"] = "gitee用户账号"
	sheetTileMap["B1"] = "gitee邮箱"
	sheetTileMap["C1"] = "积分值"
	sheetTileMap["D1"] = "osi任务完成issue数量"
	sheetTileMap["E1"] = "个人提交issue数量"
	sheetTileMap["F1"] = "个人评论issue数量"
	sheetTileMap["G1"] = "个人提交pr数量"
	sheetTileMap["H1"] = "备注"
	for k, v := range sheetTileMap {
		xlsx.SetCellValue(sheetName, k, v)
	}
	xlsx.SetActiveSheet(index)
	err := xlsx.SaveAs(excelPath)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return excelPath
}

func ReadWriteEulerExcel(excelPath string, evlChan []ExcelValue) (error) {
	file, openErr := excelize.OpenFile(excelPath)
	if openErr != nil {
		logs.Error("fail to open the file, excelPath: ", excelPath, ", openErr: ", openErr)
		return openErr
	}
	for _, ev := range evlChan {
		if ev.IssueCount == 0 && ev.Points == 0 {
			logs.Error("The data does not meet the requirements and will not be exported, ev: ", ev)
			continue
		}
		excelData := addExcelData(ev)
		if len(excelData) > 0 {
			rows, sheetErr := file.GetRows(sheetName)
			if sheetErr != nil {
				logs.Error(sheetErr)
			}
			idx := len(rows) + 1
			axis := fmt.Sprintf("A%d", idx)
			setErr := file.SetSheetRow(sheetName, axis, &excelData)
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

func addExcelData(ev ExcelValue) []interface{} {
	cveData := make([]interface{}, 0)
	cveData = append(cveData, ev.GitLogin)
	cveData = append(cveData, ev.GitEmail)
	cveData = append(cveData, ev.Points)
	cveData = append(cveData, ev.IssueCount)
	cveData = append(cveData, ev.CommitIssueCount)
	cveData = append(cveData, ev.CommentIssueCount)
	cveData = append(cveData, ev.CommitPrCount)
	cveData = append(cveData, ev.Decription)
	return cveData
}

func ExcelData(excelPath string, evlChan []ExcelValue) {
	eulerErr := ReadWriteEulerExcel(excelPath, evlChan)
	if eulerErr != nil {
		logs.Error("ReadWriteEulerExcel, eulerErr: ", eulerErr)
	}
}

func SendEulerExcel(zipFileName, startTime, endTime string, flag int) {
	if flag == 1 {
		cBody := fmt.Sprintf("hi all: \r\n 附件中有两个excel文件分别为: openEuler开源实习-参赛者周(%s~%s)积分统计和总积分统计, 请查收. \r\n", startTime, endTime)
		subject := fmt.Sprintf("openEuler开源实习-参赛者周(%s~%s)积分和总积分统计", startTime, endTime)
		sendError := SendEmail(zipFileName, 1, cBody, subject)
		if sendError != nil {
			logs.Error("SendEmail, sendErr: ", sendError)
		}
	} else {
		cBody := fmt.Sprintf("hi all: \r\n 附件中有两个excel文件分别为: openEuler开源实习-参赛者月(%s~%s)积分统计和总积分统计, 请查收. \r\n", startTime, endTime)
		subject := fmt.Sprintf("openEuler开源实习-参赛者月(%s~%s)积分和总积分统计", startTime, endTime)
		sendError := SendEmail(zipFileName, 1, cBody, subject)
		if sendError != nil {
			logs.Error("SendEmail, sendErr: ", sendError)
		}
	}
	return
}

func DelFile(fileList []string) {
	if len(fileList) > 0 {
		for _, filex := range fileList {
			err := os.Remove(filex)
			if err != nil {
				logs.Error(err)
			}
		}
	}
}

func SendEmail(attchStr string, flag int, cBody, subject string) error {
	//var mail Mail
	SendTypeStr := ""
	if flag == 1 {
		SendTypeStr = beego.AppConfig.String("email::openeuler_send_type")
	}
	sendType := strings.Split(SendTypeStr, ",")
	toEmail := make([]string, 0)
	ccEmail := make([]string, 0)
	for _, st := range sendType {
		emailType, _ := strconv.Atoi(st)
		elt := models.EmailList{EmailType: int8(emailType)}
		el, eltErr := elt.Read("EmailType")
		if eltErr != nil {
			logs.Error("Failed to get mailing list, err: ", eltErr)
		} else {
			for _, em := range el {
				if em.SendType == 1 {
					toEmail = append(toEmail, em.EmailName)
				} else {
					ccEmail = append(ccEmail, em.EmailName)
				}
			}
		}
	}
	emailError := error(nil)
	if len(toEmail) > 0 {
		sendErr := SendCommonMail(toEmail, ccEmail, subject, cBody, attchStr)
		if sendErr == nil {
			logs.Info("Notify issue statistics that the email was sent successfully! attchStr: ", attchStr)
		} else {
			logs.Error("Notify issue statistics mail delivery failure! attchStr: ", attchStr)
		}
	}
	return emailError
}

func (mail *SendMail) Auth() {
	mail.auth = smtp.PlainAuth("", mail.user, mail.password, mail.host)
}

func (mail SendMail) Send(message Message) error {
	mail.Auth()
	buffer := bytes.NewBuffer(nil)
	boundary := "GoBoundary"
	Header := make(map[string]string)
	//Header["From"] = message.from
	Header["From"] = "osi-task-manager" + "<" + message.from + ">"
	Header["To"] = strings.Join(message.to, ";")
	Header["Cc"] = strings.Join(message.cc, ";")
	Header["Bcc"] = strings.Join(message.bcc, ";")
	Header["Subject"] = message.subject
	Header["Content-Type"] = "multipart/mixed;boundary=" + boundary
	Header["Mime-Version"] = "1.0"
	Header["Date"] = time.Now().String()
	mail.writeHeader(buffer, Header)
	body := "\r\n--" + boundary + "\r\n"
	body += "Content-Type:" + message.contentType + "\r\n"
	body += "\r\n" + message.body + "\r\n"
	buffer.WriteString(body)
	attachmentName := strings.Replace(message.attachment.name, "excel/", "", -1)
	attachmentName = strings.Replace(message.attachment.name, "excel", "", -1)
	if message.attachment.withFile {
		attachment := "\r\n--" + boundary + "\r\n"
		attachment += "Content-Transfer-Encoding:base64\r\n"
		attachment += "Content-Disposition:attachment\r\n"
		attachment += "Content-Type:" + message.attachment.contentType + ";name=\"" +
			attachmentName + "\"\r\n"
		buffer.WriteString(attachment)
		defer func() {
			if err := recover(); err != nil {
				logs.Error(err)
			}
		}()
		mail.writeFile(buffer, message.attachment.name)
	}
	buffer.WriteString("\r\n--" + boundary + "--")
	toSend := make([]string, 0)
	toSend = append(toSend, message.to...)
	toSend = append(toSend, message.cc...)
	header := smtp.SendMail(mail.host+":"+mail.port, mail.auth, message.from, toSend, buffer.Bytes())
	logs.Info("header: ", header)
	return nil
}

func (mail SendMail) writeHeader(buffer *bytes.Buffer, Header map[string]string) string {
	header := ""
	for key, value := range Header {
		header += key + ":" + value + "\r\n"
	}
	header += "\r\n"
	buffer.WriteString(header)
	return header
}

// read and write the file to buffer
func (mail SendMail) writeFile(buffer *bytes.Buffer, fileName string) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err.Error())
	}
	payload := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
	base64.StdEncoding.Encode(payload, file)
	buffer.WriteString("\r\n")
	for index, line := 0, len(payload); index < line; index++ {
		buffer.WriteByte(payload[index])
		if (index+1)%76 == 0 {
			buffer.WriteString("\r\n")
		}
	}
}

// srcFile could be a single file or a directory
// ZipFiles compresses one or many files into a single zip archive file.
func ZipFiles(filename string, files []string, oldform, newform string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	for _, file := range files {
		fisExist, _ := PathExists(file)
		if !fisExist {
			logs.Error("ZipFiles, not exist, file: ", file)
			continue
		}
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.Replace(file, oldform, newform, -1)
		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func SendCommonMail(mailTo, mailCc []string, subject string, body, attchName string) error {
	emailName := beego.AppConfig.String("email::email_name")
	emailPwd := beego.AppConfig.String("email::email_pwd")
	emailHost := beego.AppConfig.String("email::email_host")
	emailPwd = ""
	emailPort := beego.AppConfig.String("email::email_port")
	mailConn := map[string]string{
		"user": emailName,
		"pass": emailPwd,
		"host": emailHost,
		"port": emailPort,
	}
	port, _ := strconv.Atoi(mailConn["port"])
	m := gomail.NewMessage()
	m.SetHeader("From", "cve-manager"+"<"+mailConn["user"]+">")
	m.SetHeader("To", mailTo...)
	m.SetHeader("Cc", mailCc...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	//m.Attach(attchName)
	_, fileName := filepath.Split(attchName)
	m.Attach(attchName,
		gomail.Rename(fileName),
		gomail.SetHeader(map[string][]string{
			"Content-Disposition": []string{
				fmt.Sprintf(`attachment; filename="%s"`, mime.QEncoding.Encode("UTF-8", fileName)),
			},
		}),
	)
	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	err := d.DialAndSend(m)
	return err
}
