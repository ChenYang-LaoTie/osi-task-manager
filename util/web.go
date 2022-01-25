package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"osi-task-manager/common"
	"strconv"
	"time"
)

//HTTPPost post request
func HTTPPost(url string, requestBody string) (map[string]interface{}, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("HTTPPost, response Status:", resp.Status)
	logs.Info("HTTPPost, response Headers:", resp.Header)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("post failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("post successed!, body: ", string(body))
	var iss map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	logs.Info(iss)
	return iss, nil
}

//HTTPPatch patch request
func HTTPPatch(url string, requestBody string) (map[string]interface{}, error) {
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("PATCH request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("HTTPPatch, response Status:", resp.Status)
	logs.Info("HTTPPatch, response Headers:", resp.Header)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Patch request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("PATCH failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("PATCH successed!, body: ", string(body))
	var iss map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	logs.Info(iss)
	return iss, nil
}

//HTTPPut patch request
func HTTPPut(url string, requestBody string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("PATCH request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("HTTPPut, response Status:", resp.Status)
	logs.Info("HTTPPut, response Headers:", resp.Header)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Patch request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("PUT failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("PUT successed!, body: ", string(body))
	var iss []map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	return iss, nil
}

// http post label
func HTTPPostLabel(url string, requestBody string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("PATCH request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("HTTPPut, response Status:", resp.Status)
	logs.Info("HTTPPut, response Headers:", resp.Header)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Patch request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("PUT failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("PUT successed!, body: ", string(body))
	var iss []map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	return iss, nil
}

//HTTPGet get request
func HTTPGet(url string) ([]map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("HTTPGet, error: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("resp.Status: ", resp.Status, resp.Header)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error("resp.Body:", resp.Body, ",err: ", err)
		return nil, err
	}
	logs.Info("body: \n", string(body))
	var col []map[string]interface{}
	err = json.Unmarshal(body, &col)
	if err != nil {
		logs.Error("HTTPGet,err: ", err, "body: ", body)
	}
	return col, nil
}

//HTTPGitGet get request
func HTTPGitGet(url string) (col map[string]interface{}, err error) {
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("HTTPGitGet, error: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("resp.Status: ", resp.Status, resp.Header)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error("err: ", err)
		return nil, err
	}
	logs.Info("body: \n", string(body))
	err = json.Unmarshal(body, &col)
	if err != nil {
		logs.Error("HTTPGitGet,err: ", err)
		return col, err
	}
	return col, nil
}

//TimeStrToInt parse time string to unix nano
func TimeStrToInt(ts, layout string) int64 {
	if ts == "" {
		return 0
	}
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	t, err := time.ParseInLocation(layout, ts, time.Local)
	if err != nil {
		logs.Error(err)
		return 0
	}
	return t.Unix()
}
