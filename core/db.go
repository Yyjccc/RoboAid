package core

import (
	"RoboAid/config"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var RssDb *RssDB

type RssDB struct {
	Path string
	Db   *sql.DB
}

func init() {
	dir := config.BotPath + cfg.DbPath
	_ = EnsureDirectoryExists(dir)
	dataSourceName := dir + "Rss.db"
	// 检查文件是否存在
	if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
		log.Debug("Database not found. Creating new database:" + dataSourceName)
		// 创建数据库
		db, err := sql.Open("sqlite3", dataSourceName)
		if err != nil {
			log.Error(err)
			return
		}
		// 创建表
		// name , description ,link,collect_count,collect_date,update_time,creator
		createTableQuery := `
        CREATE TABLE rss_source (
            id INTEGER PRIMARY KEY AUTOINCREMENT ,
            name TEXT UNIQUE NOT NULL,
            description TEXT NOT NULL ,
            public INTEGER NOT NULL ,
            link TEXT  NOT NULL ,
            collect_count INTEGER ,
            collect_date TEXT ,
            update_time TEXT NOT NULL ,
            creator TEXT -- 创建者
        );

		 CREATE TABLE rss_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT ,
			source_id    INTEGER NOT NULL ,
			description TEXT ,
			title       TEXT NOT NULL ,
			link        TEXT ,
			publish_date TEXT NOT NULL ,
			author      TEXT
		 )
 `

		_, err = db.Exec(createTableQuery)
		if err != nil {
			log.Error(err)
			return
		}
		RssDb = &RssDB{
			Path: dataSourceName,
			Db:   db,
		}
		return
	} else {
		// 打开已有的数据库
		db, err := sql.Open("sqlite3", dataSourceName)
		if err != nil {
			log.Error(err)
			return
		}
		RssDb = &RssDB{
			Path: dataSourceName,
			Db:   db,
		}
	}
}

func (r *RssDB) InsertRssRecord(record *RssRecord) (int64, error) {
	// 插入数据的 SQL 语句
	query := `INSERT INTO rss_record (source_id , description ,link,title,publish_date,author)VALUES (?,?,?, ?, ?, ?)`
	result, err := r.Db.Exec(query, record.SourceID, record.Description, record.Link, record.Title, record.PublishDate, record.Author)
	if err != nil {
		log.Error(err)
	}
	return 0, err
	// 获取插入记录的自增ID
	id, err := result.LastInsertId()
	if err != nil {
		log.Error("Failed to retrieve the last insert ID:", err)
		return 0, err
	}
	return id, nil
}

// GetRecordBySourceID 根据RSS源ID 查询获取记录
func (r *RssDB) GetRecordBySourceID(sourceID int64) ([]*RssRecord, error) {
	// 查询指定 source_id 的记录
	query := `SELECT id, source_id, description, title, link, publish_date, author FROM rss_record WHERE source_id = ?`
	rows, err := r.Db.Query(query, sourceID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()
	var records []*RssRecord
	for rows.Next() {
		var record *RssRecord
		err := rows.Scan(&record.ID, &record.SourceID, &record.Description, &record.Title, &record.Link, &record.PublishDate, &record.Author)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		records = append(records, record)
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return records, nil
}

func (r *RssDB) InsertRssSource(source *RssSource) error {
	// 插入数据的 SQL 语句
	query := `INSERT INTO rss_source (name , description ,link,collect_count,collect_date,update_time,public,creator)VALUES (?,?,?, ?, ?, ?, ?)`
	_, err := r.Db.Exec(query, source.Name, source.Description, source.Link, source.CollectCount, source.CollectDate, source.UpdateTime, source.Public, source.Creator)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (r *RssDB) DeleteRssSource(name string) error {
	// 删除指定 name 的 RSS 资源
	query := `DELETE FROM rss_source WHERE name = ?`
	_, err := r.Db.Exec(query, name)
	if err != nil {
		log.Error(err)
	}
	return err
}

// GetAllRssSource 查询所有 RSS 资源
func (r *RssDB) GetAllRssSource() ([]*RssSource, error) {

	query := `SELECT id, name, description, link, collect_count, collect_date, update_time, creator FROM rss_source`
	rows, err := r.Db.Query(query)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()
	var rssSources []*RssSource
	for rows.Next() {
		var source RssSource
		err := rows.Scan(&source.ID, &source.Name, &source.Description, &source.Link, &source.CollectCount, &source.CollectDate, &source.UpdateTime, &source.Creator)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		rssSources = append(rssSources, &source)
	}

	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return rssSources, nil
}
