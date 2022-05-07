package routers

import (
	"github.com/astaxie/beego"
	"osi-task-manager/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/intern/admin/login", &controllers.InternAdminLoginControllers{})     //用户登录，获取用户token
	beego.Router("/intern/issue/hook/event", &controllers.InternHookEventControllers{}) //gitee评论回调
	beego.Router("/intern/student", &controllers.InternStudentControllers{})            //导入学生数据或删除
	beego.Router("/intern/tutor", &controllers.InternTutorControllers{})                //导入导师数据或删除
	beego.Router("/intern/points/lists", &controllers.InternPointsListControllers{})    //积分列表
	beego.Router("/intern/user/points", &controllers.InternUserPointsControllers{})     //单人积分
	beego.Router("/healthz/readiness", &controllers.HealthzReadController{})
	beego.Router("/healthz/liveness", &controllers.HealthzLiveController{})
}
