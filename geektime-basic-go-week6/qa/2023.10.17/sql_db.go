package qa

import (
	"context"
	"database/sql"
)

func SqlDB() {
	db, err := sql.Open("mysql",
		"root:root@tcp(localhost:13316)/webook")
	if err != nil {
		panic(err)
	}
	// 增删改
	db.Exec("")
	db.ExecContext(context.Background(), "")
	// 查询
	//db.QueryRowContext()
}
