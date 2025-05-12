package core

import (
	"fmt"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/uclipboard/uclipboard/model"
)

var DB *sqlx.DB

var (
	clipboardTableName   = "clipboard_data"
	clipboardTableSchema = fmt.Sprintf(`CREATE TABLE if not exists %s (
		id integer primary key autoincrement,
		ts bigint not null,
		content text,
		hostname varchar(128) default 'unknown',
		content_type varchar(128) default 'text'
		);
	`, clipboardTableName)
	addFirstRecordToClipboardTable = fmt.Sprintf(`insert into %s
	(ts,content,hostname,content_type) select
	0,'uclipboard started!','uclipboard','text' where (select count(*) from %s) = 0
	`, clipboardTableName, clipboardTableName)
	insertReturnUpdatedClipboard = fmt.Sprintf(`insert into %s 
	(ts,content,hostname,content_type) values
	(:ts,:content,:hostname,:content_type)
	returning *
	`, clipboardTableName)
	getLatestClipboard = fmt.Sprintf(`select * from %s
	order by id desc limit `, clipboardTableName) + "%d"  //format limit N support

	fileMetadataTableName = "file_metadata"
	fileMetadataSchema    = fmt.Sprintf(`CREATE TABLE if not exists %s (
		id integer primary key autoincrement,
		created_ts bigint not null,
		expire_ts bigint not null,
		file_name varchar(128) not null,
		tmp_path varchar(256) not null
		);
	`, fileMetadataTableName)

	insertFileMetadata = fmt.Sprintf(`insert into %s
	(created_ts,expire_ts,file_name,tmp_path) values
	(:created_ts,:expire_ts,:file_name,:tmp_path)
	`, fileMetadataTableName)
	deleteFileMetadataById = fmt.Sprintf(`delete from %s
	where id = ?
	`, fileMetadataTableName)

	queryFileMetadataByExpireTs = fmt.Sprintf(`select * from %s
	where expire_ts < ?
	`, fileMetadataTableName)

	queryFileMetadataByIdOrName = fmt.Sprintf(`select * from %s
	where id = ? or file_name = ? order by created_ts desc limit 1 
	`, fileMetadataTableName)
	queryFileMetadataLatest = fmt.Sprintf(`select * from %s
	order by created_ts desc limit 1
	`, fileMetadataTableName)

	queryClipboardHistoryWithPage = fmt.Sprintf(`select * from %s
	order by id desc limit ? offset ?
	`, clipboardTableName)
	queryCountClipboardHistory = fmt.Sprintf(`select count(*) from %s
	`, clipboardTableName)

	deleteOldNClipboard = fmt.Sprintf(`delete from %s
	where id in (
		select id from %s
		order by id asc limit ?
	)`, clipboardTableName, clipboardTableName)
)
var dbLogger *logrus.Entry

func InitDB(c *model.UContext) {
	dbLogger = model.NewModuleLogger("DB") //InitDB is abosultely the first function to be called in server.Run
	DB = sqlx.MustConnect("sqlite3", c.Server.Store.DBPath)
	DB.MustExec(clipboardTableSchema)
	DB.MustExec(fileMetadataSchema)
	firsftRecordInsetResult := DB.MustExec(addFirstRecordToClipboardTable)
	N, err := firsftRecordInsetResult.RowsAffected()
	if err != nil {
		dbLogger.Fatalf("addFirstRecordToClipboardTable error: %v", err)
	}
	if N != 0 {
		dbLogger.Info("initialize clipboard table.")
	}
	dbLogger.Debug("DB init completed")
}

// This function is used to create a new record in the clipboard table
// and modify the content that is passed in
func AddClipboardRecord(c *model.Clipboard) error {
	dbLogger.Tracef("call AddclipboardRecord(%v)", c)
	rows, err := DB.NamedQuery(insertReturnUpdatedClipboard, c)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.StructScan(c)
	}
	return err
}

func QueryLatestClipboardRecord(N int) (clipboards []model.Clipboard, err error) {
	dbLogger.Tracef("call GetLatestClipboardRecord(%v)", N)
	err = DB.Select(&clipboards, fmt.Sprintf(getLatestClipboard, N))
	return
}

// find the latest record
func GetFileMetadataLatestRecord(d *model.FileMetadata) (err error) {
	dbLogger.Tracef("call GetFileMetadataLatestRecord(%v)", d)
	err = DB.Get(d, queryFileMetadataLatest)
	return
}

// find the latest record by id or name
func GetFileMetadataRecordByIdOrName(d *model.FileMetadata) (err error) {
	dbLogger.Tracef("call GetFileMetadataRecordByOrName(%v)", d)
	err = DB.Get(d, queryFileMetadataByIdOrName, d.Id, d.FileName)
	return
}

func AddFileMetadataRecord(d *model.FileMetadata) (fileId int64, err error) {
	dbLogger.Tracef("call AddFileMetadataRecord(%v)", d)
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
	dbLogger.Tracef("call DelFileMetadataRecordById(%v)", d)
	_, err = DB.Exec(deleteFileMetadataById, d.Id)
	if err != nil {
		return
	}
	return
}

func DelTmpFile(conf *model.UContext, d *model.FileMetadata) (err error) {
	dbLogger.Tracef("call DelTmpFile(%v)", d)
	err = os.Remove(path.Join(conf.Server.Store.TmpPath, d.TmpPath))
	if err != nil {
		return err
	}
	return nil
}

func QueryExpiredFiles(conf *model.UContext, t int64) (expiredFiles []model.FileMetadata, err error) {
	dbLogger.Tracef("call QueryExpiredFiles(%v)", t)
	err = DB.Select(&expiredFiles, queryFileMetadataByExpireTs, t)
	return
}

func QueryClipboardHistory(conf *model.UContext, page int) (clipboards []model.Clipboard, err error) {
	dbLogger.Tracef("call QueryClipboardHistory(%v)", page)
	err = DB.Select(&clipboards, queryClipboardHistoryWithPage,
		conf.Server.Api.HistoryPageSize,
		(page-1)*conf.Server.Api.HistoryPageSize)
	return
}

func CountClipboardHistory(conf *model.UContext) (count int, err error) {
	dbLogger.Tracef("call CountClipboardHistory()")
	err = DB.Get(&count, queryCountClipboardHistory)
	return
}

func DeleteOutdatedClipboard(conf *model.UContext) (err error) {
	dbLogger.Trace("call DeleteOutdatedClipboard()")
	var count int
	err = DB.Get(&count, queryCountClipboardHistory)
	if err != nil {
		return
	}

	if count > conf.Server.Store.MaxClipboardRecordNumber {
		dbLogger.Debugf("delete %d old clipboard records", count-conf.Server.Store.MaxClipboardRecordNumber)

		_, err = DB.Exec(deleteOldNClipboard, count-conf.Server.Store.MaxClipboardRecordNumber)
		if err != nil {
			return
		}
	}

	return
}
