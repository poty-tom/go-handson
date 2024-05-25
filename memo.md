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
    /-models.go
    /-data.go
/-repositories
    /-articles.go
    /-comments.go
```
事前にmysqlドライバをインストールしておく必要がある
```
go get github.com/go-sql-driver/mysql
```