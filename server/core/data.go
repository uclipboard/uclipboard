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
	addFirstRecordToClipboardTable = fmt.Sprintf(`insert into %s
	(ts,content,hostname,content_type) select
	0,'uclipboard started!','uclipboard','text' where (select count(*) from %s) = 0
	`, clipboard_table_name, clipboard_table_name)
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

	queryFileMetadataByIdOrName = fmt.Sprintf(`select * from %s
	where id = ? or file_name = ? order by created_ts desc limit 1 
	`, file_metadata_table_name)
	queryFileMetadataLatest = fmt.Sprintf(`select * from %s
	order by created_ts desc limit 1
	`, file_metadata_table_name)

	queryClipboardHistoryWithPage = fmt.Sprintf(`select * from %s
	order by id desc limit ? offset ?
	`, clipboard_table_name)
)
var logger *logrus.Entry

func InitDB(c *model.Conf) {
	logger = model.NewModuleLogger("DB") //InitDB is abosultely the first function to be called in server.Run
	DB = sqlx.MustConnect("sqlite3", c.Server.DBPath)
	DB.MustExec(clipboard_schema)
	DB.MustExec(file_metadata_schema)
	firsftRecordInsetResult := DB.MustExec(addFirstRecordToClipboardTable)
	N, err := firsftRecordInsetResult.RowsAffected()
	if err != nil {
		logger.Fatalf("addFirstRecordToClipboardTable error: %v", err)
	}
	if N != 0 {
		logger.Info("initialize clipboard table.")
	}
	logger.Debug("DB init completed")
}

func AddClipboardRecord(c *model.Clipboard) (err error) {
	logger.Tracef("call AddclipboardRecord(%v)", c)
	_, err = DB.NamedExec(insertClipboard, c)
	return
}

func QueryLatestClipboardRecord(N int) (clipboards []model.Clipboard, err error) {
	logger.Tracef("call GetLatestClipboardRecord(%v)", clipboards)
	err = DB.Select(&clipboards, fmt.Sprintf(getLatestClipboard, N))
	return
}

// find the latest record
func GetFileMetadataLatestRecord(d *model.FileMetadata) (err error) {
	logger.Tracef("call GetFileMetadataLatestRecord(%v)", d)
	err = DB.Get(d, queryFileMetadataLatest)
	return
}

// find the latest record by id or name
func GetFileMetadataRecordByOrName(d *model.FileMetadata) (err error) {
	logger.Tracef("call GetFileMetadataRecordByOrName(%v)", d)
	err = DB.Get(d, queryFileMetadataByIdOrName, d.Id, d.FileName)
	return
}

func AddFileMetadataRecord(d *model.FileMetadata) (fileId int64, err error) {
	logger.Tracef("call AddFileMetadataRecord(%v)", d)
	result, err := DB.NamedExec(insertFileMetadata, d)
	if err != nil {
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		return
	}
	return id, nil
}

func DelFileMetadataRecordById(d *model.FileMetadata) (err error) {
	logger.Tracef("call DelFileMetadataRecordById(%v)", d)
	_, err = DB.Exec(deleteFileMetadataById, d.Id)
	if err != nil {
		return
	}
	return
}

func DelTmpFile(conf *model.Conf, d *model.FileMetadata) (err error) {
	logger.Tracef("call DelTmpFile(%v)", d)
	err = os.Remove(path.Join(conf.Server.TmpPath, d.TmpPath))
	if err != nil {
		return err
	}
	return nil
}

func QueryExpiredFiles(conf *model.Conf, t int64) (expiredFiles []model.FileMetadata, err error) {
	logger.Tracef("call QueryExpiredFiles(%v)", t)
	err = DB.Select(&expiredFiles, queryFileMetadataByExpireTs, t)
	return
}

func QueryClipboardHistory(conf *model.Conf, page int) (clipboards []model.Clipboard, err error) {
	logger.Tracef("call QueryClipboardHistory(%v)", page)
	err = DB.Select(&clipboards, queryClipboardHistoryWithPage,
		conf.Server.ClipboardHistoryPageSize,
		(page-1)*conf.Server.ClipboardHistoryPageSize)
	return
}
