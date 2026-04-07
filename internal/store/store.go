package store

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

type DB struct{ db *sql.DB }
type Document struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int    `json:"size_bytes"`
	Folder    string `json:"folder"`
	Tags      string `json:"tags"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "archivist.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS documents(id TEXT PRIMARY KEY,title TEXT NOT NULL,filename TEXT DEFAULT '',mime_type TEXT DEFAULT '',size_bytes INTEGER DEFAULT 0,folder TEXT DEFAULT '/',tags TEXT DEFAULT '',notes TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(
	resource TEXT NOT NULL,
	record_id TEXT NOT NULL,
	data TEXT NOT NULL DEFAULT '{}',
	PRIMARY KEY(resource, record_id)
)`)
	return &DB{db: db}, nil
}
func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }
func (d *DB) Create(e *Document) error {
	e.ID = genID()
	e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO documents(id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at)VALUES(?,?,?,?,?,?,?,?,?)`, e.ID, e.Title, e.Filename, e.MimeType, e.SizeBytes, e.Folder, e.Tags, e.Notes, e.CreatedAt)
	return err
}
func (d *DB) Get(id string) *Document {
	var e Document
	if d.db.QueryRow(`SELECT id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at FROM documents WHERE id=?`, id).Scan(&e.ID, &e.Title, &e.Filename, &e.MimeType, &e.SizeBytes, &e.Folder, &e.Tags, &e.Notes, &e.CreatedAt) != nil {
		return nil
	}
	return &e
}
func (d *DB) List() []Document {
	rows, _ := d.db.Query(`SELECT id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at FROM documents ORDER BY created_at DESC`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Document
	for rows.Next() {
		var e Document
		rows.Scan(&e.ID, &e.Title, &e.Filename, &e.MimeType, &e.SizeBytes, &e.Folder, &e.Tags, &e.Notes, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}
func (d *DB) Update(e *Document) error {
	_, err := d.db.Exec(`UPDATE documents SET title=?,filename=?,mime_type=?,size_bytes=?,folder=?,tags=?,notes=? WHERE id=?`, e.Title, e.Filename, e.MimeType, e.SizeBytes, e.Folder, e.Tags, e.Notes, e.ID)
	return err
}
func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM documents WHERE id=?`, id)
	return err
}
func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&n)
	return n
}

func (d *DB) Search(q string, filters map[string]string) []Document {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (title LIKE ?)"
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id,title,filename,mime_type,size_bytes,folder,tags,notes,created_at FROM documents WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Document
	for rows.Next() {
		var e Document
		rows.Scan(&e.ID, &e.Title, &e.Filename, &e.MimeType, &e.SizeBytes, &e.Folder, &e.Tags, &e.Notes, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

func (d *DB) Stats() map[string]any {
	m := map[string]any{"total": d.Count()}
	return m
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
