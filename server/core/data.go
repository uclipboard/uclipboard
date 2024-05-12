package core

import (
	"fmt"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

var (
	clipboard_table_name = "clipboard_data"
	schema               = fmt.Sprintf(`CREATE TABLE if not exists %s (
		id integer primary key autoincrement,
		ts bigint not null,
		content text,
		hostname varchar(128) default 'unknown',
		content_type varchar(128) default 'text'
		);
	`, clipboard_table_name)

	insertClipboard = fmt.Sprintf(`insert into %s 
	(ts,content,hostname,content_type) values
	(:ts,:content,:hostname,:content_type)
	`, clipboard_table_name)
	getLatestClipboard = fmt.Sprintf(`select * from %s
	order by id desc limit 1`, clipboard_table_name)
)

func InitDB(c *model.Conf) {
	DB = sqlx.MustConnect("sqlite3", c.Server.DBPath)
	DB.MustExec(schema)

}

func AddClipboardRecord(c *model.Clipboard) (err error) {
	_, err = DB.NamedExec(insertClipboard, c)
	return
}

func GetLatestClipboardRecord(c *model.Clipboard) (err error) {
	err = DB.Get(c, getLatestClipboard)
	return
}
