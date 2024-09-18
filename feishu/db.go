package feishu

import (
	"RoboAid/config"
	"RoboAid/core"
	"database/sql"
	"os"
	"time"
)

var fsDb *FeiShuDB

// 私有RSS订阅
type PrivateRss struct {
	ID         string
	SourceID   int64
	OpenID     string
	CreateDate string
}

// 推送设置
type SubscribeInfo struct {
	ID         string
	OpenId     string
	Subscribe  int
	UpdateTime time.Time
}

// 推送记录
type PushRecord struct {
	ID       string
	RecordID string
	OpenID   string
	PushDate string
}

type FeiShuDB struct {
	Path string
	Db   *sql.DB
}

func init() {
	dir := config.BotPath + config.BotConfig.DbPath
	_ = core.EnsureDirectoryExists(dir)
	dataSourceName := dir + "FeiShu.db"
	// 检查文件是否存在
	if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
		log.Debug("Database not found. Creating new database:" + dataSourceName)
		// 创建数据库
		db, err := sql.Open("sqlite3", dataSourceName)
		if err != nil {
			log.Error(err)
			return
		}
		// 创建表:私有RSS , 推送列表
		createTableQuery := `
        CREATE TABLE private_rss (
            id INTEGER PRIMARY KEY AUTOINCREMENT ,
            source_id INTEGER,
            open_id TEXT NOT NULL ,
            create_date TEXT
        );

		CREATE TABLE subscribe_info (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			open_id  TEXT NOT NULL,
			subscribe INTEGER,
			update_time TEXT
		);
	CREATE TABLE push_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			open_id  TEXT NOT NULL,
			record_id INTEGER,
			push_date TEXT
		);
        `
		_, err = db.Exec(createTableQuery)
		if err != nil {
			log.Error(err)
			return
		}
		fsDb = &FeiShuDB{
			Path: dataSourceName,
			Db:   db,
		}
	} else {
		// 打开已有的数据库
		db, err := sql.Open("sqlite3", dataSourceName)
		if err != nil {
			log.Error(err)
			return
		}
		fsDb = &FeiShuDB{
			Path: dataSourceName,
			Db:   db,
		}
	}
}

func (f *FeiShuDB) InsertSubscribeInfo(info *SubscribeInfo) error {
	query := `INSERT INTO subscribe_info (open_id,subscribe, update_time)VALUES (?,?,?)`
	_, err := f.Db.Exec(query, info.OpenId, info.Subscribe, info.UpdateTime.Format("2006-01-02"))
	if err != nil {
		log.Error(err)
	}
	return err
}

func (f *FeiShuDB) InsertPushRecord(openID string, recordID int64) error {
	query := `INSERT INTO push_record (open_id,record_id, push_date)VALUES (?,?,?)`
	_, err := f.Db.Exec(query, openID, recordID, time.Now().Format("2006-01-02"))
	if err != nil {
		log.Error(err)
	}
	return err
}

func (f *FeiShuDB) UpdateSubscribeInfo(openID string, sub int) error {
	query := `UPDATE subscribe_info SET subscribe = ?, update_time = ? WHERE open_id = ? `
	_, err := f.Db.Exec(query, sub, time.Now().Format("2006-01-02"), openID)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (f *FeiShuDB) GetSubscribeInfo(openID string) *SubscribeInfo {
	query := `SELECT open_id,subscribe FROM subscribe_info WHERE open_id = ?`

	sub := &SubscribeInfo{}
	err := f.Db.QueryRow(query, openID).Scan(&sub.OpenId, &sub.Subscribe)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Error(err)
		return nil
	}
	return sub
}

func (f *FeiShuDB) GetAllPrivateRss(id int64) ([]*PrivateRss, error) {
	// 查询语句
	query := `SELECT id, source_id, open_id, create_date FROM private_rss WHERE source_id = ?`

	// 执行查询
	rows, err := f.Db.Query(query, id)
	if err != nil {
		log.Error("Failed to retrieve private RSS records:", err)
		return nil, err
	}
	defer rows.Close()

	// 初始化结果切片
	var rssList []*PrivateRss

	// 遍历结果集
	for rows.Next() {
		var rss *PrivateRss
		err := rows.Scan(&rss.ID, &rss.SourceID, &rss.OpenID, &rss.CreateDate)
		if err != nil {
			log.Error("Failed to scan RSS record:", err)
			return nil, err
		}
		// 添加到结果切片
		rssList = append(rssList, rss)
	}

	// 检查 rows 是否有任何错误
	if err := rows.Err(); err != nil {
		log.Error("Error occurred during row iteration:", err)
		return nil, err
	}

	return rssList, nil
}

// 根据id
func (f *FeiShuDB) GetAllPrivateRssByUserID(openId string) ([]*core.RssSource, error) {
	// 查询语句
	query := `SELECT id, source_id, open_id, create_date FROM private_rss WHERE open_od = ?`

	// 执行查询
	rows, err := f.Db.Query(query, openId)
	if err != nil {
		log.Error("Failed to retrieve private RSS records:", err)
		return nil, err
	}
	defer rows.Close()
	// 初始化结果切片
	var rssList []*PrivateRss
	// 遍历结果集
	for rows.Next() {
		var rss *PrivateRss
		err := rows.Scan(&rss.ID, &rss.SourceID, &rss.OpenID, &rss.CreateDate)
		if err != nil {
			log.Error("Failed to scan RSS record:", err)
			return nil, err
		}
		// 添加到结果切片
		rssList = append(rssList, rss)
	}

	// 检查 rows 是否有任何错误
	if err := rows.Err(); err != nil {
		log.Error("Error occurred during row iteration:", err)
		return nil, err
	}
	var ressult []*core.RssSource
	for _, rss := range rssList {
		source := core.RssDb.GetRssSource(rss.SourceID)
		if source != nil {
			ressult = append(ressult, source)
		}
	}
	return ressult, nil
}
