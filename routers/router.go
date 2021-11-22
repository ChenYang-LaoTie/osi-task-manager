package routers

import (
	"github.com/astaxie/beego"
	"osi-task-manager/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/intern/admin/login", &controllers.InternAdminLoginControllers{})
	beego.Router("/intern/issue/hook/event", &controllers.InternHookEventControllers{})
	beego.Router("/intern/student", &controllers.InternStudentControllers{})
	beego.Router("/intern/tutor", &controllers.InternTutorControllers{})
	beego.Router("/intern/points/lists", &controllers.InternPointsListControllers{})
	beego.Router("/intern/user/points", &controllers.InternUserPointsControllers{})
	beego.Router("/healthz/readiness", &controllers.HealthzReadController{})
	beego.Router("/healthz/liveness", &controllers.HealthzLiveController{})
}
