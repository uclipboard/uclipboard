package core

import (
	"fmt"
	"os"
	"path"

	"github.com/dangjinghao/uclipboard/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var DB *sqlx.DB

var (
	clipboard_table_name = "clipboard_data"
	clipboard_schema     = fmt.Sprintf(`CREATE TABLE if not exists %s (
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

	file_metadata_table_name = "file_metadata"
	file_metadata_schema     = fmt.Sprintf(`CREATE TABLE if not exists %s (
		id integer primary key autoincrement,
		created_ts bigint not null,
		expire_ts bigint not null,
		file_name varchar(128) not null,
		tmp_path varchar(256) not null
		);
	`, file_metadata_table_name)

	insertFileMetadata = fmt.Sprintf(`insert into %s
	(created_ts,expire_ts,file_name,tmp_path) values
	(:created_ts,:expire_ts,:file_name,:tmp_path)
	`, file_metadata_table_name)
	deleteFileMetadataById = fmt.Sprintf(`delete from %s
	where id = ?
	`, file_metadata_table_name)

	queryFileMetadataByExpireTs = fmt.Sprintf(`select * from %s
	where expire_ts < ?
	`, file_metadata_table_name)
)
var logger *logrus.Entry

func InitDB(c *model.Conf) {
	logger = model.NewModuleLogger("InitDB")
	DB = sqlx.MustConnect("sqlite3", c.Server.DBPath)
	DB.MustExec(clipboard_schema)
	DB.MustExec(file_metadata_schema)

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

func AddFileMetadataRecord(d *model.FileMetadata) (err error) {
	logger.Tracef("in AddFileMetadataRecord: %v", d)
	_, err = DB.NamedExec(insertFileMetadata, d)
	return
}

func DelFileMetadataRecordById(d *model.FileMetadata) (err error) {
	logger.Tracef("DelFileMetadataRecordById: %v", d)
	_, err = DB.Exec(deleteFileMetadataById, d.Id)
	if err != nil {
		return
	}
	return
}

func DelTmpFile(conf *model.Conf, d *model.FileMetadata) (err error) {
	logger.Tracef("DelTmpFile: %v", d)
	err = os.Remove(path.Join(conf.Server.TmpPath, d.TmpPath))
	if err != nil {
		return err
	}
	return nil
}

func QueryExpiredFiles(conf *model.Conf, t int64) (expiredFiles []model.FileMetadata, err error) {
	err = DB.Select(&expiredFiles, queryFileMetadataByExpireTs, t)
	return
}
