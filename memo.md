ひさびさのdocker-compose.yaml
構文とか覚える気もないけどさすがに全くやらないと忘れるので定期的に書く場面に遭遇するのは割とありがたい

```
version: '3.3' //3.3なんて見たことないのだが。
services:
    mysql:
        image: mysql:5.7
        container_name: db-for-go
    command:
        - --character-set-server=utf8mb4
        - --collation-server=utf8mb4_unicode_ci
        - --sql-mode=ONLY_FULL_GROUP_BY,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION
    environment:
        MYSQL_ROOT_USER: root
        MYSQL_ROOT_PASSWORD: pass
        ...
    ports:
        - "3306:3306"

    volumes:
        - db-volume:/var/lib/mysql
volumes:
    db-volume:
```

createTable.sql
は別で作ってSQLを定義する

## GoからMySQLへ接続する話

```
dbConnection := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=True", dbUser, dbPassword, dbDatabase)

db, err := sql.Open("mysql", dbConnection) 
if err != nil {
    fmt.Println(err)
}

defer db.Close() // 終了タイミングの問題があるため
```

database/sqlパッケージのOpen関数を使う

```
func (db *DB) Ping()
```
ドライバを入れる必要がある。
go get github.com/go-sql-driver/mysql

## ドライバの必要性
DBと直接通信を行うための仕組みがdatabase/sqlパッケージ側にはないため、
実際にDBと通信を行うための通信レイヤの確立が必要

```
func main() {
    dbUser := "docker"
    ...
    ...
    dbConn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser....)

    db, err := sql.Open("mysql", dbConn)
    if err != nil {
        fmt.Println(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("connect to DB")
    }
}
```

↑の実装で動作的には問題ないけど、これだけだとDBとの通信はできないので、
別でgo getコマンドにてドライバをインストールしておく

```
go get github.com/go-sql-driver/mysql
```

標準パッケージ内にデータベース接続処理が実装されなかった理由は抽象化が目的
⇒DBの変更があった際にはそのままdb.Ping()が利用できる
実際に必要な変更は
- インストールするドライバを変更
- sql.Open関数に渡す接続先アドレス

```
func Open(driverName, dataSource string)(*DB, error)
```

Open関数に渡す引数はドライバによって異なるため利用するドライバ（間接的にはDBの種類）に応じてドキュメントを参照


## クエリを発行する
直接SQLを発行するところから
```
const sqlStr = `
    select * from articles;
`

rows, err := db.Query(sqlStr) // sql.Rows
if err != nil {
    fmt.Println(err)
    return
}

defer rows.Close() // Closeが必要なため事前にdeferで実行する

articleArray := make([]models.Article, 0) // model.Article型のスライスを作成
for rows.Next() { // イテレータと同じ動作。次の値か、終わりかを返す
    var article models.Article
    // fmt.Scanと同じ、引数に読み出し結果を格納したい変数のポインタを指定する
    err := rows.Scan(&article.Title, &article.Contents, &article.Username, ...)

    if err != nil {
        fmt.Println(err)
    } else {
        articleArray = append(articleArray, article)
    }
}

fmt.Printf("%+v\n", articleArray)
```

db.Queryによって返されるsql.Rows構造体
こいつも使い終わったらCloseしてあげる必要があるらしいのでdeferで遅延処理のCloseを挟んでおく
```
type Rows struct {

}
```

## null値を許容
```
rows, err := db.Query("select * from articles;") 
rows.Scan(,,,&article.created_at) // Created: null の場合
```
みたいなケースではエラーになる

○エラー回避案
- NULL出ない場合は通常通り&article.CreatedAtにぶち込む
- NULLだった場合は例外処理

NULLかどうかの判定はsql.NullXXX型を使って判断可能
```
type NullTime struct {
    Time time.Time
    Valid bool
}
```

```
var nt sql.NullTime
rows.Scan(&nt)

if nt.Valid {
    // rowsから読み取った日付型がnullだったので例外処理ないしは決められた処理ルートに進む
}
```

```
func main() {
    articleArray := make([]models.Article, 0)
    for rows.Next() {
        var article models.Article
        var createdTime sql.NullTime
        err := rows.Scan(&article.Id, ....&createdTime)

        if createdTime.Valid {
            article.CreatedAt = acreatedTime.Time
        }
    }
}
```

## クエリにバインド変数をぶっこむ話
ドライバによってはプレースホルダーが異なる。
go-sql-driver/mysqlの場合は?を利用する
```
articleID := 1
const sql = `
    select * from articles
    where article_id = ?;
    `
rows, err := db.Query(sql, articleID)
```
pqの場合は$1

## insert文の実行
参照更新それぞれで利用する関数が異なる
select ・・・ db.Query('select ...')
update/insert/delete ・・・ db.Exec('...')
```
func main() {
    dbUser...

    db, err := sql.Open("mysql", dbConnection)
    if err != nil {
        fmt.Println(err)
    }

    defer db.Close()

    article := models.Articles {
        Title: "insert test",
        ...
    }

    const sqlStr = `
        insert into articles (...) values (?, ?, ..., now());
    `
    // db.Execには参照ではなく値を渡す
    result, err := db.Exec(sqlStr, article.Title, ar....)
    if err != nil {
        fmt.Println(err)
        return
    }
}
```

### resultについて
```
type Result interface {
    LastInsertId() (int64, error) // 何番目にinsertされたか

    RowsAffected() (int64, error) // 何レコードが影響を受けたか
}
```

## トランザクションを利用した更新
```
func (db *DB) Begin() (*Tx, error)

type Tx struct {
    // sql.Tx型はsql.DBと同等の型っぽい。
    // トランザクション中ではこのTx型を介してSQLの発行をする
}
```
実際のコード例

```
func main() {
    tx, err := db.Begin()
    if err != nil {
        fmt.Println(err)
        return
    }

    article_id := 1
    const sqlStr = `
        select nice from articles
        where article_id = ?;
    `

    row := tx.QueryRow(sqlStr, article_id)
    if err := row.Exec(); err != nul {
        fmt.Println(err)
        tx.Rollback() // rollback transaction
        return
    }

    var nicenum int
    err = row.Scan(&nicenum)
    if err != nil {
        fmt.Println(err)
        tx.Rollback()
        return
    }

    const sqlUpd = `update articles set nice = ? where article_id = ?`
    _, err = tx.Exec(sqlUpd, nicenum+1, article_id)
    if err != nil {
        fmt.Println(err)
        tx.Rollback()
        return 
    }

    tx.Commit() // トランザクション内でエラーが起きなければcommit
}
```

# API側でDB操作を実装
ここから実践だけどメモしながら理解する
理解に問題がなければ実際のソースも修正していく

## 構成
```
/-handlers
/-models
    -models.go
    -data.go
/-repositories
    /-articles.go
    /-comments.go
```
事前にmysqlドライバをインストールしておく必要がある
```
go get github.com/go-sql-driver/mysql
```

# 4章
ユニットテストの実装

## テストコード
goではxx_test.goファイルがテストコードとして認識されるらしい

このことからも命名規則がとても大事だとわかる

### repositoryのサンプルテストコード
```
package repositories_test

import (
    "testing"

    "my-repository"
)

func TestSelectArticleDetail(t *testing.T) {
    // 期待する値
    expected := models.Article {

    }

    // リポジトリから取得された値
    got, err := repositories.SelectArticleDetail(適切な引数)
    if err != nil {
        t.Fatal(err)
    }

    if got != expected {
        t.Errorf("get %s but want %s\n", got, expected)
    }
}
```

- パッケージ名
![Go中級本に記載があった](./img/package_name_rule.png)

パッケージ名の命名規則では、
main以外のパッケージについてはディレクトリ名と同じパッケージ名をつける必要がある。
⇒repositoriesディレクトリ上では、package repositoriesとつけていたのはそういう意図

一方でテストコードについては[ディレクトリ名_test]というパッケージ名が例外的に許容されている


## testingパッケージ
標準パッケージであるtestingというものが用意されている。
こやつを使ってテストコードを実装する

## ユニットテスト関数の形
- 関数名:TestXxxx
- 引数:*testing.T
- 戻り値:なし

```
func TestSelectArticleDetail(t *testing.T) {
    expected := ...

    got, err := repositories.Xxxx
    if err != nil {
        t.Fatal(err)
    }

    if got != expected {
        t.Errorf("get ...)
    }
}
```

## t.Fatal系のメソッド

- Fatal
fmt.Printlnに近い
- Fatalf
fmt.Printfに近い


## test main関数
```
package main_test

import (

)

func setup() {
    // 前処理
}

func teardown() {
    // 後処理
}


func TestMain(m *testing.M) {
    setup()

    m.Run()

    teardown()
}

func TestA()

```


# サービス層やる

既存のハンドラの処理を整理
1. パスからIDを取得
2. IDの記事をDBから取得
3. 結果をレスポンスに書き込む

現時点で2ができてない

■復習
```
func ArticleListHandler(w http.ResponseWriter, req *http.Request) {
    queryMap := req.URL.Query()

    var page int

    log.Println(page)

    articleList := []models.Article{models.Article1, models.Article2}

    json.NewEncoder(w).Encode(articleList)
}
```


## サービス関数の定義
サービス層に必要な機能
- IDをもとにDBから記事の取得をする
- リクエスト内の記事情報を元にDBにレコードを追加する
- 指定記事にいいねする
etc.

サンプル

```
func GetArticleService(articleID int) (models.Article, error) {
    // TODO: sql.DB型を受け取って変数dbにぶっこむ
    article, err := repositories.SelectArticleDetail(db, articleID)
    if err != nil {
        return models.Article{}, err
    }

    commentList, err := repositories.SelectCommentList(db, articleID)
    if err != nil {
        return models.Article{}, err
    }
    
    article.CommentList = append(article.CommentList, commentList..)

    return article, nil
}
```

## helper.go

```
package services

var (
    dbUser = "docker"
    dbPassword = "docker"
    ...
)

func connectDB() (*sql.DB, error) {
    db, err := sql.Open("mysql", dbConn)
    if err != nil {
        return nil, err
    }

    return db, nil
}
```

MYSQL_ROOT_USER=root
MYSQL_ROOT_PASSWORD=password
MYSQL_DATABASE=sample
MYSQL_USER=mysql
MYSQL_PASSWORD=password

Getenv関数で環境変数を読むように処理を直す

```
import (
    ...
    "os"
)

var (
    dbUser = os.Getenv("MYSQL_USER")
    dbPassword = os.Getenv("MYSQL_PASSWORD)
    dbDatabase = os.Getenv("MYSQL_DATABASE")
    dbConn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser, dbPassword, dbDatabase)
)


```

一旦5章の内容でAPIを実装してみる。
dbConnectionを都度つど取得する形でやる感じか・。？

# 6章
アーキテクチャ改装
## サービス構造体の導入
```
package service

import "database/sql"

// API全体で利用するsql.DBへのポインタを持っておく
type MyAppService struct {
    db *sql.DB
}

func NewMyAppService(db *sql.DB) *MyAppService {
    return &MyAppService{db: db}
}
```

## 練習
上記サービス構造体を利用して既存のAPIを改変

1. サービス関数を修正 
```
// 修正前
func GetArticleListService(page int) ([]models.Article, error) {
	db, err := connectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	articleList, err := repositories.SelectArticleList(db, page)
	if err != nil {
		return nil, err
	}

	return articleList, nil
}


// 修正後
func (s *MyAppService) GetArticleListService(page int) ([]models.Article, error) {
	articleList, err := repositories.SelectArticleList(s.db, page)
	if err != nil {
		return nil, err
	}

	return articleList, nil
}

```

2. ハンドラ層の改変
```
func PostArticleHandler(w http.ResponseWriter, req *http.Request) {
    a := [service.MyAppService構造体を取得]
    article, err := s.PostArticleService(reqArticle)
}
```
でもこのやり方するとサービス層を改変したメリットがないので、同様にハンドラを改変してコントローラ層を導入する

```
// 以下のようにコントローラーパッケージを構築する
/--controllers
    /--controllers.go
```

```
type MyAppController struct {
    service *servies.MyAppService
}

func NewMyAppController(s *services.MyAppService) *MyAppController {
    return &MyAppController{service: s}
}
```

```
func (c *MyAppController) PostArticleHandler(w http.ResponseWriter, req *http.Request) {
    var reqArticle, models.Article
    if err := json.NewDecoder(req.Body).Decode(&reqArticle); err != nil {
        http.Error(...)
    }

    artcicle,err := c.service.PostArticlService(reqAricle)

}
```
コントローラ層もサービス層と同様に
コントローラ構造体を作り、構造体に対してのふるまいという形で定義する。



3. ハンドラ呼び出し側も修正
```
func main() {
    ser := services.NewMyAppService([sql.DB]型)
    con := controllers.NewMyAppController(ser)
    
    r := mux.NewRouter()

    r.HandleFunc("/article", con.PostArticleHandler).Methods(http.MethodPost)
    ...
}
```

4. ルータ層の作成
```
package routers

import (

)

func NewRouter(con *controllers.MyAppController) *mux.Router {
    // gorilla mux routerを作る
    r := mux.NewRouter()

    // コントローラないで定義されるハンドラをパスに紐づける
    r.HandleFunc("path", con.PostHandler).Methods(http.MethodPost)

    return r // 作成したrouterへのポインタを返す
}
```
この結果main自体がとてもすっきりする
```
fun main() {
    db, err := sql.Open("mysql", dbConn)

    if err != nil {
        log.Println("fail to connect DB")
        return 
    }

    ser := services.NewMyAppService(db)
    con := controllers.NewAppController(ser)

    r := routers.NewRouter(con)
}
```

## インターフェースによる抽象化・疎結合化
コントローラ層から見ると、サービスそうの実装はどうでもいい、
きめられた構造体を投げたら、構造体とエラーを返してほしいという取り決めくらいあればいい

⇒インタフェースを定義
```
/--controllers
    /--controllers.go
    /--services
        /--services.go
```

```
package services

import xxx

type MyAppServicer interface {
    PostArticleService(article models.Article)(models.Article, error)
    GetArticleListService(page int)([]models.Article, error)
    ...
}
```

昨日途切れてしまった続きからやる。
今日の内に6章ひととおり完了できれば良い感じか

インタフェースの導入によって、ハンドラが利用するサービスの付け替えが用意になる

## インターフェースの小型化による役割の分離
前項ではすべてのメソッドを持つものを一つのインタフェースとして定義していたけど、これだと抽象度が低い

```
type ArticleServicer interface {
    PostArticleService(article models.Article) (models.Article, error)
    GetArticleListService(page int) ([]models.Article, error)
    GetArticleService(articleID, int) (models.Article, error)
    PostNiceService(article models.Article) (models.Article, error)
}

type CommentSerivcer interface {
    PostCommentService(comment models.Comment) (models.Comment, error)
}
```

## コントローラの分離
```
/--controllers
    /--services
        /--services.go
    /--article_controller.go
    /--comment_controller.go
```

```
package controllers

type ArticleController struct {
    service services.ArticleServicer
}

func NewArticleController(s services.ArticleServicer) *ArticleController {
    return &ArticleController{service: s}
}

func (c *ArticleController) HelloHandler(w http.ResponseWriter, req *http.Request) {

}

func (c *ArticleController) PostArticleHandler(w http.ResponseWriter, req *http.Request) {

}

func (c *ArticleController) ArticleListHandler(w http.ResponseWriter, req *http.Request){

}

...


// comment用のコントローラ構造体も同様に作成
package controllers

import (

)

type CommentController struct {
    service services.CommentServicer
}

func NewCommentController(s services.CommentServicer) *CommentController {
    return &CommentController{service: s}
}

func (c *CommentController) PostCommentHandler(w http.ResponseWriter, req *http.Requst){

}

```

main関数側も改修
```
func main() {
    db, err := sql.Open("mysql", dbConn)
    if err != nil {
        log.Println("fail to connect DB")
        return 
    }

    ser := services.NewMyAppService(db)
    // service構造体から二つのコントローラを作成
    aCon := controllers.NewArticleController(ser)
    cCon := controllers.NewCommentController(ser)

    // 2つのコントローラ構造体から gorilla/muxのるーたを作成
    r := routers.NewRouter(aCon, cCon)
}
```


router側も改修

```
func NewRouter(aCon *controllers.ArticleController, cCon *controllers.CommentController) *mux.Router {
    r := mux.NewRouter()

    // article関連のハンドラを登録
    r.HandleFunc("/article", aCon.PostArticleHandler).Methods(http.MethodPost)
    r.HandleFunc("/article/list", aCon.ArticleListHandler).Methods(http.MethodGet)
    ...

    // comment関連のハンドラを登録
    r.HandleFunc("/comment", cCon.PostCommentHandler).Methods(http.MethodPost)

    return r
}
```

さらにrouterを改善
現在の実装だとコントローラ2つが直接指定されている
⇒今後もコントローラが増え続けた際に都度引数を増やしていくのは効率が悪いので、NewRouter関数内で必要になるコントローラ構造体をその場で作る

```
func NewRouter(db *sql.DB) *mux.Router {
    ser := services.NewMyAppService(db)
    aCon := controllers.NewArticleController(ser)
    cCon := controllers.NewCommentController(ser)

    r := mux.NewRouter()

    r.HandlerFunc(....)
    ...
}
```

結果的にmain関数内ももっとシンプルになる

```
func main() {
    db, err := sql.Open("mysql", dbConn)
    if err != nil {
        log.Println("fail to connect DB")
        return 
    }

    // connectionが取得できたら
    // routersでrouterを作成してあげるだけ。
    r := routers.NewRouter(db)
}
```