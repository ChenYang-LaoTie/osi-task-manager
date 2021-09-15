package task

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"osi-task-manager/taskhandler"
)

// start task
func StartTask() {
	toolbox.StartTask()
}

func StopTask() {
	toolbox.StopTask()
}

//InitTask Timing task initialization
func InitTask() bool {
	// Download open source list information
	downloadflag, errxs := beego.AppConfig.Int("task::downloadflag")
	if downloadflag == 1 && errxs == nil {
		downloadexcel := beego.AppConfig.String("task::downloadexcel")
		DownloadExcelTask(downloadexcel)
	}
	// Calculate the integral data in the missed issue regularly
	calculatepointflag, errxs := beego.AppConfig.Int("task::calculatepointflag")
	if calculatepointflag == 1 && errxs == nil {
		calculatepoint := beego.AppConfig.String("task::calculatepoint")
		CalculateIssuePointsTask(calculatepoint)
	}

	// Get the original yaml data
	getIssueFlag, errxs := beego.AppConfig.Int("task::getissueflag")
	if getIssueFlag == 1 && errxs == nil {
		getIssue := beego.AppConfig.String("task::getissue")
		GetOriginIssueTask(getIssue)
	}

	// export points task
	exportPointFlag, errxs := beego.AppConfig.Int("task::exportwpointflag")
	if exportPointFlag == 1 && errxs == nil {
		exportWPoint := beego.AppConfig.String("task::exportwpoint")
		taskhandler.GetWeekPointsTask(exportWPoint)
	}

	exportMPointFlag, errxs := beego.AppConfig.Int("task::exportmpointflag")
	if exportMPointFlag == 1 && errxs == nil {
		exportMPoint := beego.AppConfig.String("task::exportmpoint")
		taskhandler.GetMonthPointsTask(exportMPoint)
	}

	// The classification is released, and the questions that the user has not dealt with
	// during the synchronization phase are redistributed to others
	reottaskflag, errxs := beego.AppConfig.Int("task::reottaskflag")
	if reottaskflag == 1 && errxs == nil {
		reottask := beego.AppConfig.String("task::reottask")
		RelOvertimeClaimedTasks(reottask)
	}

	// Release frozen user accounts and blacklisted accounts
	relblacklistflag, errxs := beego.AppConfig.Int("task::relblacklistflag")
	if relblacklistflag == 1 && errxs == nil {
		relblacklist := beego.AppConfig.String("task::relblacklist")
		EulerRelBlackTask(relblacklist)
	}

	// Export issue pr and number of comments
	exportissueprflag, errxs := beego.AppConfig.Int("task::exportissueprflag")
	if exportissueprflag == 1 && errxs == nil {
		exportissuepr := beego.AppConfig.String("task::exportissuepr")
		EulerIssueStatisticsTask(exportissuepr)
	}
	// Releasing the limited number of canceled tasks per month
	monthrelunassignflag, errxs := beego.AppConfig.Int("task::monthrelunassignflag")
	if monthrelunassignflag == 1 && errxs == nil {
		monthrelunassign := beego.AppConfig.String("task::monthrelunassign")
		MonthRelUnassignTask(monthrelunassign)
	}
	// Export points for the specified week
	specexportwpointflag, errxs := beego.AppConfig.Int("task::specexportwpointflag")
	if specexportwpointflag == 1 && errxs == nil {
		specexportwpoint := beego.AppConfig.String("task::specexportwpoint")
		taskhandler.GetSpecWeekPointsTask(specexportwpoint)
	}
	return true
}