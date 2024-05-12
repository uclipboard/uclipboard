package core

import (
	"fmt"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
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
	order by id desc limit `, clipboard_table_name) + "%d"  //format limit N support
)
var logger *logrus.Entry

func InitDB(c *model.Conf) {
	logger = model.NewModuleLogger("InitDB")
	DB = sqlx.MustConnect("sqlite3", c.Server.DBPath)
	DB.MustExec(schema)

	logger.Debug("DB init completed")
}

func AddClipboardRecord(c *model.Clipboard) (err error) {
	logger.Tracef("AddclipboardRecord: %v", c)
	_, err = DB.NamedExec(insertClipboard, c)
	return
}

func GetLatestClipboardRecord(c *[]model.Clipboard, N int) (err error) {
	logger.Tracef("GetLatestClipboardRecord: %v", c)
	err = DB.Select(c, fmt.Sprintf(getLatestClipboard, N))
	return
}
