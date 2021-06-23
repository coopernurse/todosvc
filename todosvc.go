package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/husobee/vestigo"
)

func main() {
	startHttp(initDb())
}

func initDb() *sql.DB {
	// connect
	dsn := os.Getenv("DB_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("ERROR opening connection to MySQL: %v", err)
	}

	// create tables
	sqls := []string{
		`create table if not exists todo (id int primary key auto_increment, note varchar(1024))`,
	}
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil {
			log.Fatalf("ERROR running db migration. sql=%s err=%v", sql, err)
		}
	}
	return db
}

func startHttp(db *sql.DB) {
	svc := &todoService{db: db}

	router := vestigo.NewRouter()
	router.Get("/", svc.Get)
	router.Post("/", svc.Create)
	router.Patch("/:id", svc.Update)
	router.Delete("/:id", svc.Delete)

	log.Printf("Starting todosvc on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func httpErr(w http.ResponseWriter, msg string, err error) {
	log.Printf("ERROR %s err=%v", msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

type todoService struct {
	db *sql.DB
}

func (s *todoService) Get(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("select id, note from todo order by id")
	if err != nil {
		httpErr(w, "500 - Unable to select from todo table", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var note string
		err = rows.Scan(&id, &note)
		if err != nil {
			log.Printf("ERROR Get Scan failed: %v", err)
			return
		}
		w.Write([]byte(fmt.Sprintf("%d,%s\n", id, note)))
	}
}

func (s *todoService) Create(w http.ResponseWriter, r *http.Request) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		httpErr(w, "500 - Create unable to read POST body", err)
		return
	}

	stmt, err := s.db.Prepare("insert into todo (note) values (?)")
	if err != nil {
		httpErr(w, "500 - Create unable to prepare sql", err)
		return
	}
	_, err = stmt.Exec(buf.String())
	if err != nil {
		httpErr(w, "500 - Create unable to exec sql", err)
		return
	}
	w.Write([]byte("Todo created"))
}

func (s *todoService) Update(w http.ResponseWriter, r *http.Request) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		httpErr(w, "500 - Update unable to read POST body", err)
		return
	}

	stmt, err := s.db.Prepare("update todo set note=? where id=?")
	if err != nil {
		httpErr(w, "500 - Update unable to prepare sql", err)
		return
	}
	id := vestigo.Param(r, "id")
	_, err = stmt.Exec(buf.String(), id)
	if err != nil {
		httpErr(w, "500 - Update unable to exec sql", err)
		return
	}
	w.Write([]byte("Todo updated"))
}

func (s *todoService) Delete(w http.ResponseWriter, r *http.Request) {
	stmt, err := s.db.Prepare("delete from todo where id=?")
	if err != nil {
		httpErr(w, "500 - Delete unable to prepare sql", err)
		return
	}
	id := vestigo.Param(r, "id")
	_, err = stmt.Exec(id)
	if err != nil {
		httpErr(w, "500 - Delete unable to exec sql", err)
		return
	}
	w.Write([]byte("Todo deleted"))
}
