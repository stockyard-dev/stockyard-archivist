package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Document struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	SizeBytes int `json:"size_bytes"`
	Folder string `json:"folder"`
	Tags string `json:"tags"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"archivist.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS documents(id TEXT PRIMARY KEY,title TEXT NOT NULL,filename TEXT DEFAULT '',mime_type TEXT DEFAULT '',size_bytes INTEGER DEFAULT 0,folder TEXT DEFAULT '/',tags TEXT DEFAULT '',notes TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Document)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO documents(id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at)VALUES(?,?,?,?,?,?,?,?,?)`,e.ID,e.Title,e.Filename,e.MimeType,e.SizeBytes,e.Folder,e.Tags,e.Notes,e.CreatedAt);return err}
func(d *DB)Get(id string)*Document{var e Document;if d.db.QueryRow(`SELECT id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at FROM documents WHERE id=?`,id).Scan(&e.ID,&e.Title,&e.Filename,&e.MimeType,&e.SizeBytes,&e.Folder,&e.Tags,&e.Notes,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Document{rows,_:=d.db.Query(`SELECT id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at FROM documents ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Document;for rows.Next(){var e Document;rows.Scan(&e.ID,&e.Title,&e.Filename,&e.MimeType,&e.SizeBytes,&e.Folder,&e.Tags,&e.Notes,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM documents WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&n);return n}
