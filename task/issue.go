package task

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"osi-task-manager/taskhandler"
)

// Download open source list information
func DownloadExcelTask(downloadexcel string) {
	downloadexcelTask := toolbox.NewTask("DownloadFile", downloadexcel, taskhandler.DownloadFile)
	toolbox.AddTask("DownloadFile", downloadexcelTask)
}

// Get the original issue data on gitee
func GetOriginIssueTask(getIssue string) {
	logs.Info("Get the original issue data and start...")
	issueTask := toolbox.NewTask("GetOriginIssue", getIssue, taskhandler.GetGitOriginIssue)
	toolbox.AddTask("GetOriginIssue", issueTask)
	logs.Info("End of obtaining the original data of the issue...")
}

// openEuler regularly releases tasks in the blacklist
func EulerRelBlackTask(relblacklist string) {
	logs.Info("releases tasks in the blacklist start...")
	blackTask := toolbox.NewTask("RelFrozenBlackTask", relblacklist, taskhandler.RelFrozenBlackTask)
	toolbox.AddTask("RelFrozenBlackTask", blackTask)
	logs.Info("End of releases tasks in the blacklist...")
}

// Automatically add sig tags
func AutoAddSigLabelTask(autaddsiglabel string) {
	logs.Info("Automatically add sig tags start...")
	labelTask := toolbox.NewTask("AutoAddLabelTask", autaddsiglabel, taskhandler.AutoAddLabelTask)
	toolbox.AddTask("AutoAddLabelTask", labelTask)
	logs.Info("End of Automatically add sig tags...")
}

// The classification is released, and the questions that the user has not dealt with
// during the synchronization phase are redistributed to others
func RelOvertimeClaimedTasks(reottask string) {
	logs.Info("Release the overtime claimed task to start...")
	overtimeTask := toolbox.NewTask("ProcOvertimeClaimedTasks", reottask, taskhandler.ProcOvertimeClaimedTasks)
	toolbox.AddTask("ProcOvertimeClaimedTasks", overtimeTask)
	logs.Info("Release overtime claimed task ends...")
}

func EulerIssueStatisticsTask(exportissuepr string) {
	logs.Info("Issue statistics service started...")
	statisticsTask := toolbox.NewTask("EulerIssueStatistics", exportissuepr, taskhandler.EulerIssueStatistics)
	toolbox.AddTask("EulerIssueStatistics", statisticsTask)
	logs.Info("Issue statistics service end...")
}

func MonthRelUnassignTask(monthrelunassign string) {
	logs.Info("Release the limit on the number of monthly cancellations start...")
	monthrelunassignTask := toolbox.NewTask("MonthRelUnassign", monthrelunassign, taskhandler.MonthRelUnassign)
	toolbox.AddTask("MonthRelUnassign", monthrelunassignTask)
	logs.Info("Release the limit on the number of monthly cancellations end...")
}

func CalculateIssuePointsTask(calculatepoint string) {
	logs.Info("Calculate the integral data in the missed issue regularly and start...")
	pointsTask := toolbox.NewTask("CalculateIssuePoints", calculatepoint, taskhandler.CalculateIssuePoints)
	toolbox.AddTask("CalculateIssuePoints", pointsTask)
	logs.Info("Periodic calculation of the integral data in the missed issue of the statistics ends...")
}
