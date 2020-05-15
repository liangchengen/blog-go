package main

import (
	md52 "crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//文章简要数据
type ArticleBriefly struct {
	Title  string `json:"title"`
	Time   int64  `json:"time"`
	Praise int32  `json:"praise"`
	View   int32  `json:"view"`
}

//获取文章数据文章数据
type GetArticle struct {
	Title  string `json:"title"`
	Data   string `json:"data"`
	Time   int64  `json:"time"`
	Praise int32  `json:"praise"`
	View   int32  `json:"view"`
}

//获取文章简要数据数组
type GetArticleBrieflyArray struct {
	Data [2]ArticleBriefly `json:"data"`
}
//错误
func checkErr(err error){
	if err != nil {
		fmt.Println(err)
	}
}
//网页列表请求
func articleList(w http.ResponseWriter, r *http.Request) {
	var (
		 id int
		 title string
		 time int
		 praise int
		 view int
		 response string
	     text map[string]interface{}
		 pageNumber int
	)
	//连接数据库
	db, err := sql.Open("mysql", "root:Liang!123@/web?charset=utf8")
	defer db.Close()
	checkErr(err)
	//查询总共有多少条数据
	row1,err := db.Query("select count(id) from article")
	for row1.Next(){
		row1.Scan(&pageNumber)
	}

	//查询数据
	x,_ := strconv.Atoi(r.URL.Query()["pages"][0])
	rows, err := db.Query("SELECT id,title,time,praise,view FROM article where id > "+ strconv.Itoa(x*12-12)+ " and id <= " + strconv.Itoa(x*12))
	checkErr(err)
	response = `{"page":{"number":`+ strconv.Itoa(pageNumber) +`},"data":[`
	for rows.Next() {
		err = rows.Scan(&id,&title,&time, &praise, &view)
		checkErr(err)
		response += "{\"id\":"+ strconv.Itoa(id) +",\"title\":\""+ title +"\",\"time\":"+ strconv.Itoa(time) +",\"praise\":"+ strconv.Itoa(praise) +",\"view\":"+ strconv.Itoa(view) + "},"

	}
	response = strings.TrimSuffix(response,",")
	response += `]}`
	err = json.Unmarshal([]byte(response), &text)
	checkErr(err)

	json, err := json.Marshal(text)
	checkErr(err)

	fmt.Fprintf(w, string(json))
}
//登录界面
func login(w http.ResponseWriter, r *http.Request) {
	var (
		cookie string
		response string
		text map[string]interface{}
	)

	//连接数据库
	db, err := sql.Open("mysql", "root:Liang!123@/web?charset=utf8")
	defer db.Close()

	checkErr(err)
	//查询cookie
	row1,err := db.Query("select cookie from user")
	checkErr(err)
	for row1.Next() {
		err = row1.Scan(&cookie)
		checkErr(err)
	}

	//验证cookie
	r.ParseForm()
	postData := r.PostFormValue("name") + r.PostFormValue("password") + "fgxnxnxiyiuvhj"

	md5 := md52.Sum([]byte(postData))
	postData= fmt.Sprintf("%x",string(md5[:]))



	if postData == cookie {
		response = `{"result":true,"text":"登录成功"}`
		//设置cookie
		http.SetCookie(w,&http.Cookie{
			Name:    "cookie",
			Value:   cookie,
			Expires: time.Now().Add(168 * time.Hour),
			HttpOnly: false,
			Path:     "/",
		})
	}else if postData != cookie {
		response = `{"result":false,"text":"账号或密码错误"}`
	}

	//返回消息
	err = json.Unmarshal([]byte(response), &text)
	checkErr(err)

	json, err := json.Marshal(text)
	checkErr(err)
	fmt.Fprintf(w, string(json))

}
//查询用户
func getUser(w http.ResponseWriter, r *http.Request){
	var(
		cookie string
		i int
		text map[string]interface{}
	)
	//连接数据库
	db, err := sql.Open("mysql", "root:Liang!123@/web?charset=utf8")
	defer db.Close()

	checkErr(err)
	//查询cookie
	x ,err :=r.Cookie("cookie")
	cookie = x.Value
	row1,err := db.Query(`select count(cookie) FROM user WHERE cookie="` + cookie +`"`)
	checkErr(err)
	for row1.Next() {
		err = row1.Scan(&i)
		checkErr(err)
	}

	//返回消息
	if i == 1 {
		err = json.Unmarshal([]byte(`{"signin":true}`), &text)
		checkErr(err)
	}else{
		err = json.Unmarshal([]byte(`{"signin":false}`), &text)
		checkErr(err)
	}

	json, err := json.Marshal(text)
	checkErr(err)
	fmt.Fprintf(w, string(json))


}
//新建文章
func addArticle(w http.ResponseWriter,r *http.Request){
	var (
		cookie string
		i int
		response string
		text map[string]interface{}
	)
	//连接数据库
	db, err := sql.Open("mysql", "root:Liang!123@/web?charset=utf8")
	defer db.Close()
	checkErr(err)
	//查询cookie
	x ,err :=r.Cookie("cookie")
	cookie = x.Value
	row1,err := db.Query(`select count(cookie) FROM user WHERE cookie="` + cookie +`"`)
	checkErr(err)
	for row1.Next() {
		err = row1.Scan(&i)
		checkErr(err)
	}

	//插入数据
	r.ParseForm()
	if i == 1{
		//添加文章到数据库
		stmt, err := db.Prepare("INSERT INTO article SET title=?,data=?,time=?,praise=?,view=?")
		defer db.Close()
		checkErr(err)
		_, err = stmt.Exec(r.PostFormValue("title"), r.PostFormValue("data"),strconv.FormatInt(time.Now().Unix()*1000,10),0,0)
		checkErr(err)
		//成功消息
		response = `{"result":true}`
	}else {
		//失败消息
		response = `{"result":false}`
	}

	//返回消息
	err = json.Unmarshal([]byte(response), &text)
	checkErr(err)
	json, err := json.Marshal(text)
	checkErr(err)
	fmt.Fprintf(w, string(json))
}
//文章详情
func article(w http.ResponseWriter,r *http.Request){
	var (
		title string
		data string
		view int
		response string
		text map[string]interface{}
	)
	//连接数据库
	db, err := sql.Open("mysql", "root:Liang!123@/web?charset=utf8")
	defer db.Close()
	checkErr(err)

	//查询文章
	r.ParseForm()
	row1,err := db.Query(`select title,data,view FROM article WHERE id=` + r.PostFormValue("id"))
	checkErr(err)
	for row1.Next() {
		err = row1.Scan(&title,&data,&view)
		checkErr(err)
	}

	//阅读量+1
	stmt, err := db.Prepare("update article set view=? where id=?")
	checkErr(err)
	_, err = stmt.Exec(view + 1, r.PostFormValue("id"))
	checkErr(err)

	//返回消息
	response = `{"title":"` + title +  `","data":"`+ data +`"}`
	err = json.Unmarshal([]byte(response), &text)
	checkErr(err)

	json, err := json.Marshal(text)
	checkErr(err)
	fmt.Fprintf(w, string(json))
}
//文章点赞
func setPraiseNumber(w http.ResponseWriter,r *http.Request){
	var (
		praise int
	)
	//连接数据库
	db, err := sql.Open("mysql", "root:Liang!123@/web?charset=utf8")
	defer db.Close()
	checkErr(err)

	//查询点赞量
	r.ParseForm()
	row1,err := db.Query(`select praise FROM article WHERE id=` + r.PostFormValue("id"))
	checkErr(err)
	for row1.Next() {
		err = row1.Scan(&praise)
		checkErr(err)
	}
	//点赞+1
	stmt, err := db.Prepare("update article set praise=? where id=?")
	checkErr(err)

	_, err = stmt.Exec(praise + 1, r.PostFormValue("id"))
	checkErr(err)
}

func main() {
	http.HandleFunc("/setPraiseNumber", setPraiseNumber)
	http.HandleFunc("/article", article)
	http.HandleFunc("/articlelist", articleList)
	http.HandleFunc("/addArticle", addArticle)
	http.HandleFunc("/getuser", getUser)
	http.HandleFunc("/login", login)

	http.ListenAndServe(":2929", nil)
}
