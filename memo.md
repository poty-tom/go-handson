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