package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opesun/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	var i = 1
	db, err := sql.Open("mysql", "user:pwd@tcp(localhost:3306)/dbname?charset=utf8")
	checkErr(err)
	cookie := ""
	for {
		content := httpGet(i, cookie)
		parseHtml(content, db)
		i++
	}
	defer db.Close()
}

func httpGet(i int, cookie string) io.ReadCloser {
	url := "http://xxxxx.io/favorites?page=" + strconv.Itoa(i)
	fmt.Println(url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		print(err)
		os.Exit(1)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", cookie)

	resp, err := client.Do(req)
	checkErr(err)

	body, err := ioutil.ReadAll(resp.Body)

	//ioutil.ReadAll 在读完io.Reader后会将io.Reader清空,因此需要将其恢复
	buf := bytes.NewBuffer(body)
	resp.Body = ioutil.NopCloser(buf)

	flag := strings.Contains(string(body), "您还没有任何收藏")
	if flag {
		os.Exit(0)
	}

	return resp.Body
}

func parseHtml(content io.ReadCloser, db *sql.DB) {
	p, err := goquery.Parse(content)
	checkErr(err)
	favorites := p.Find(".post")
	for i := 0; i < favorites.Length(); i++ {
		d := favorites.Eq(i)
		title := d.Find(".title")
		link := title.Find("a")
		titleText := title.Text()
		linkText := link.Attr("href")
		titleText = strings.TrimSpace(titleText)
		fmt.Println(title.Text())
		fmt.Println(link.Attr("href"))
		save(titleText, linkText, db)
	}

}

func save(title, link string, db *sql.DB) {
	flag := query(db, title, link) //记录存在就不出理直接返回
	if flag {
		return
	}
	stmt, err := db.Prepare("INSERT toutiao SET title=?,link=?")
	checkErr(err)
	res, err := stmt.Exec(title, link)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println(id)
}

func query(db *sql.DB, title, link string) bool {
	var sql = "select id from toutiao where title = ? and link = ?"
	var id int
	err := db.QueryRow(sql, title, link).Scan(&id)
	if err != nil || id == 0 {
		return false
	}
	return true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
