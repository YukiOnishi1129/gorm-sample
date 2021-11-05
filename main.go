package main

import (
	"Go/go-gorm/my"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

// db variable.
var dbDriver = "sqlite3"
var dbName = "data.sqlite3"

// session variable
var sesName = "ytboard-session"
var cs = sessions.NewCookieStore([]byte("secret-key-1234"))


// login check
func checkLogin(w http.ResponseWriter, rq *http.Request) *my.User {
	ses, _ := cs.Get(rq, sesName)
	if ses.Values["login"] == nil || !ses.Values["login"].(bool) {
		http.Redirect(w, rq, "/login", 302)
	}
	ac := ""
	if ses.Values["account"] != nil {
		ac = ses.Values["account"].(string)
	}

	var user my.User
	// DBアクセス
	db, _ := gorm.Open(dbDriver, dbName)
	// DBを解放
	defer db.Close()

	// First: 1つのレコードだけ取得する
	// First($モデル構造体)
	db.Where("account = ?", ac).First(&user)

	return &user
}

// Templates for no-template
func notemp() *template.Template {
	tmp,_ := template.New("index").Parse("NO PAGE.")
	return tmp
}

// get target Template.
func page(fname string) *template.Template {
	tmps,_ := template.ParseFiles("templates/"+fname+".html")
	return tmps
}

// top page handler
func index(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	db,_ := gorm.Open(dbDriver, dbName)
	defer db.Close()

	var pl []my.Post
	// Find: 引数にはポインタで配列を渡す
	// Find(&モデル配列)
	db.Where("group_id > 0").Order("created_at desc").Limit(10).Find(&pl)
	var gl []my.Group
	db.Where("created_at desc").Limit(10).Find(&gl)

	item := struct {
		Title string
		Message string
		Name string
		Account string
		Plist []my.Post
		Glist []my.Group
	} {
		Title: "Index",
		Message: "This is Top page.",
		Name: user.Name,
		Account: user.Account,
		Plist: pl,
		Glist: gl,
	}
	er := page("index").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// top page handler
func post(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	pid := rq.FormValue("pid")
	db,_ := gorm.Open(dbDriver, dbName)
	defer db.Close()

	if rq.Method == "POST" {
		msg := rq.PostFormValue("message")
		pId, _ := strconv.Atoi(pid)
		cmt := my.Comment {
			UserId: int(user.Model.ID),
			PostId: pId,
			Message: msg,
		}
		db.Create(&cmt)
	}

	var pst my.Post
	var cmts []my.CommentJoin

	db.Where("id = ?", pid).First(&pst)
	db.Table("comments").Select("comments.*, users.id, users.name").Joins("join users on users.id =comments.user_id").Where("comments.post_id = ?", pid).Order("created_at desc").Find(&cmts)

	item := struct {
		Title string
		Message string
		Name string
		Account string
		Post my.Post
		Clist []my.CommentJoin
	} {
		Title: "Post",
		Message: "Post id=" + pid,
		Name: user.Name,
		Account: user.Account,
		Post: pst,
		Clist: cmts,
	}

	er := page("post").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// home handler
func home(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	db,_ := gorm.Open(dbDriver, dbName)
	defer db.Close()

	if rq.Method == "POST" {
		switch rq.PostFormValue("form") {
		case "post":
			ad := rq.PostFormValue("address")
			ad = strings.TrimSpace(ad)
			if strings.HasPrefix(ad, "https://youtu.be/") {
				ad = strings.TrimPrefix(ad, "https://youtu.be/")
			}

			pt := my.Post{
				UserId:  int(user.Model.ID),
				Address: ad,
				Message: rq.PostFormValue("message"),
			}
			db.Create(&pt)
		case "group":
			gp := my.Group{
				UserId: int(user.Model.ID),
				Name: rq.PostFormValue("name"),
				Message: rq.PostFormValue("message"),
			}
			db.Create(&gp)
		}
	}

	var pts []my.Post
	var gps []my.Group

	db.Where("user_id=?", user.ID).Order("created_at desc").Limit(10).Find(&pts)
	db.Where("user_id=?", user.ID).Order("created_at desc").Limit(10).Find(&gps)

	item := struct{
		Title string
		Message string
		Name string
		Account string
		Plist []my.Post
		Glist []my.Group
	} {
		Title: "Home",
		Message: "User account=\"" + user.Account + "\".",
		Name: user.Name,
		Account: user.Account,
		Plist: pts,
		Glist: gps,
	}

	er := page("home").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}


// group handler
func group(w http.ResponseWriter, rq *http.Request) {
	user := checkLogin(w, rq)

	gid := rq.FormValue("gid")
	db,_ := gorm.Open(dbDriver, dbName)
	defer db.Close()

	if rq.Method == "POST" {
		ad := rq.PostFormValue("address")
		ad = strings.TrimSpace(ad)
		if strings.HasPrefix(ad, "https://youtu.be/") {
			ad = strings.TrimPrefix(ad, "https://youtu.be/")
		}
		gId ,_ := strconv.Atoi(gid)
		pt := my.Post{
			UserId: int(user.Model.ID),
			Address: ad,
			Message: rq.PostFormValue("message"),
			GroupId: gId,
		}
		db.Create(&pt)
	}

	var grp my.Group
	var pts []my.Post

	db.Where("id=?", gid).First(&grp)
	db.Where("created_at desc").Model(&grp).Related(&pts)

	itm := struct {
		Title string
		Message string
		Name string
		Account string
		Group my.Group
		Plist []my.Post
	} {
		Title: "Group",
		Message: "Group id=" + gid,
		Name: user.Name,
		Account: user.Account,
		Group: grp,
		Plist: pts,
	}
	er := page("group").Execute(w, itm)
	if er != nil {
		log.Fatal(er)
	}
}

// login handler
func login(w http.ResponseWriter, rq *http.Request) {
	item := struct {
		Title string
		Message string
		Account string
	} {
		Title: "Login",
		Message: "type your account & password",
		Account: "",
	}

	if rq.Method == "GET" {
		er := page("login").Execute(w, item)
		if er != nil {
			log.Fatal(er)
		}
		return
	}

	if rq.Method == "POST" {
		db,_ := gorm.Open(dbDriver, dbName)
		defer db.Close()

		usr := rq.PostFormValue("account")
		pass := rq.PostFormValue("pass")
		item.Account = usr

		// check account and password
		var re int
		var user my.User

		db.Where("account = ? and password = ?", usr, pass).Find(&user).Count(&re)

		if re <= 0 {
			item.Message = "Wrong account or password."
			page("login").Execute(w, item)
			return
		}

		// logined.
		ses,_ := cs.Get(rq, sesName)
		ses.Values["login"] = true
		ses.Values["account"] = usr
		ses.Values["name"] = user.Name
		ses.Save(rq, w)
		http.Redirect(w, rq, "/", 302)
	}

	er := page("login").Execute(w, item)
	if er != nil {
		log.Fatal(er)
	}
}

// logout handler
func logout(w http.ResponseWriter, rq *http.Request) {
	ses,_ := cs.Get(rq, sesName)
	ses.Values["login"] = nil
	ses.Values["account"] = nil
	ses.Save(rq, w)
	http.Redirect(w, rq, "/login", 302)
}

func main() {
	// マイグレーションを実行
	// my.Migrate()

	// index handling
	http.HandleFunc("/", func(w http.ResponseWriter, rq *http.Request) {
		index(w, rq)
	})

	// home handling
	http.HandleFunc("/home", func(w http.ResponseWriter, rq *http.Request) {
		home(w, rq)
	})
	
	// post handling
	http.HandleFunc("/post", func(w http.ResponseWriter, rq *http.Request) {
		post(w, rq)
	})

	// login handling
	http.HandleFunc("/login", func(w http.ResponseWriter, rq *http.Request) {
		login(w, rq)
	})

	// logout handling
	http.HandleFunc("/logout", func(w http.ResponseWriter, rq *http.Request) {
		logout(w, rq)
	})

	http.ListenAndServe("", nil)
}