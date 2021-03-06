package controllers

import (
	r "github.com/cuu/softradius/routers"
	"github.com/cuu/softradius/models"
	"github.com/cuu/softradius/libs"
	"fmt"
	"strconv"
	//	"reflect"
	"github.com/astaxie/beego"

	//	sort "github.com/cuu/softradius/libs/sortutil"
	"encoding/json"
	re "gopkg.in/gorethink/gorethink.v3"
	rdb "github.com/cuu/softradius/database/shelf"
	
)

type DefController struct {
	BaseController
}

//每个controller 有这样的命名规则,保证不重复
var _def_ctl DefController

func init(){
	//在这儿独立处理映身关系
	// 事实上只有 方法是 get的 Route才应该是 Is_menu显示出来
	_ctl := &_def_ctl
	_cate := r.MenuPlugin
	
 	_ctl.routes = append( _ctl.routes,
		r.Route{Path:"/login",Name:"登录",Category:_cate,Is_menu :false, Order:1.2,Is_open:false, Methods:"*:Login"})


 	_ctl.routes = append( _ctl.routes,
		r.Route{Path:"/logout",Name:"登出",Category:_cate,Is_menu :false, Order:1.23,Is_open:false, Methods:"get:Logout"})

	
 	_ctl.routes = append( _ctl.routes,
		r.Route{Path:"/",Name:"主页",Category:_cate,Is_menu :false, Order:1.3,Is_open:true, Methods:"*:HomePage"})

 	_ctl.routes = append( _ctl.routes,
		r.Route{Path:"/quicksearch",Name:"快速搜索栏",Category:_cate,Is_menu :false, Order:1.4,Is_open:true, Methods:"*:QuickSearch"})
		
	_ctl.AddRoutes()
	
}

//把this中的routes 放到 routers.Permits 中,每个Controller写一遍
func (this *DefController) AddRoutes() {
	
	for i,v := range  this.routes {
		if v.Methods != "" {
			beego.Router(v.Path, this, v.Methods)
		}else{
			beego.Router(v.Path, this)
		}
		
		//Permits.routes is map,so no confict path key ! 
		r.Permits.Add_route(v.Path,&this.routes[i])
	}
	
}

//主要做Form的要关准备
func (this *DefController) GuuPrepare(){
	
	this.TplName = libs.GetTplName(this)
	
	this.Layout = "login.html"
	
	this.LayoutSections["Sidebar"] = ""
	this.LayoutSections["Header"]  = ""
	this.LayoutSections["Footer"]  = ""
	this.LayoutSections["ContentHeader"] = ""
	this.LayoutSections["HeadCss"] = ""

	this.PerPage = 100
}

//------------------------------------------------------------

func (this *DefController) HomePage(){
	this.Redirect("/node",302)
}


func (this *DefController) Logout(){
	this.Reset_cookie()
	this.Redirect("/",302)
	
	
}

func (this *DefController) Login()  {

	if this.Ctx.Input.IsPost() {
		this.Login_post()
		return
	}

	this.Render()
}


func (this *DefController) Login_post() {

	type RET struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
	}
        username := this.GetString("username")
	password := this.GetString("password")

	var ret RET

// 两种 用法 
//	rsp,err := rdb.DataBase().FilterGet("operators",re.Row.Field("Name").Eq(username).And(re.Row.Field("Pass").Eq(password)))
	rsp,err := rdb.DataBase().FilterGet("operators",map[string]string{"Name":username,"Pass":password})

	if err == nil  {
		var auser Operators
		
		err = rsp.One(&auser)
		if err == nil {
			fmt.Println("Login_post ", auser)
			this.SetCookie("username",auser.Name )
			this.SetCookie("opr_type",strconv.Itoa(auser.Type))
			this.SetCookie("login_time", libs.Get_currtime())
			this.SetCookie("login_ip",this.Get_clientip())

			if auser.Type > 0 {
				//先解绑 unbind_opr
				//再从数据库的 取得所有的 opr rule
				//bind_opr一回
				
			}
			ret.Code = 0
			ret.Msg = "ok"
                }else{
			this.Reset_cookie()
			ret.Code = 1
			ret.Msg ="用户名密码不符"
		}
		
	}else{

		this.Reset_cookie()
		ret.Code = 1
		ret.Msg  = "用户名密码不符"	
	}
       
        b,err:= json.Marshal(ret)
	if err == nil {
		this.Ctx.WriteString( string( b) )
	}else{
		this.Ctx.WriteString( "{code:1,msg:\"JSON ERROR\"}")
	}
	
}

func (this *DefController) QuickSearchMembers(qu string ,skip int )(int, []Members){
	var nods []Members

	rdb.DataBase().SearchSkipGetFunc(&nods,func(me re.Term)re.Term{
		return me.Field("Name").Match(qu)
	}, skip,this.PerPage)
	
	total := rdb.DataBase().SearchCount(&nods,"Name",qu)
	
	return total,nods
}


func (this *BaseController) MakePager(total int, page int){
	
	if page >= 0 {
		url := libs.RemoveSuffix(this.Ctx.Input.URI(),"/")
		pg   := models.NewPager(total,this.PerPage, page,url)	
		this.Data["Paginator"] = pg.Render()
	}
	
}

func (this *DefController) QuickSearch() {

	query := this.GetString("q")
	
	if query == "" {
		this.Abort("403")
	}
	
	nods := this.NodeList()
	pdus := this.ProductList()

	page := this.GetStringI("page_id")
//	if page <= 0 { page = 1}
	total,mbms := this.QuickSearchMembers(query, libs.Or(int(page-1),0).(int)*this.PerPage )
	this.MakePager(total,page)
	
	this.Data["MemberList"] = mbms
	this.Data["ProductMap"] = this.ToPairMapS(pdus,[]string{"Id","Name"})
	this.Data["NodeMap"]    = this.ToPairMapS(nods,[]string{"Id","Name"})
	this.Data["IsExpire"]  = libs.IsExpire
	
	this.ResetLayout()
	this.TplName ="bus_member_list.html"


	
	this.Render()
}
