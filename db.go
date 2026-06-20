package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var beltOrder = map[string]int{
	"白": 1, "黄": 2, "橙": 3, "绿": 4, "蓝": 5, "红": 6, "黑": 7,
}

var validBelts = []string{"白", "黄", "橙", "绿", "蓝", "红", "黑"}
var validStatuses = []string{"在籍", "停训", "退出"}
var validTrainingTypes = []string{"基础", "对练", "体能", "套路"}
var validExamResults = []string{"通过", "未通过"}

type Student struct {
	ID      int64  `json:"id"`
	Nick    string `json:"nick"`
	Phone   string `json:"phone"`
	Belt    string `json:"belt"`
	Status  string `json:"status"`
}

type TrainingRecord struct {
	ID        int64  `json:"id"`
	StudentID int64  `json:"student_id"`
	Nick      string `json:"nick,omitempty"`
	Date      string `json:"date"`
	Type      string `json:"type"`
	Duration  int    `json:"duration"`
}

type ExamRecord struct {
	ID          int64  `json:"id"`
	StudentID   int64  `json:"student_id"`
	Nick        string `json:"nick,omitempty"`
	Date        string `json:"date"`
	TargetBelt  string `json:"target_belt"`
	Coach       string `json:"coach"`
	Result      string `json:"result"`
}

func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nick TEXT NOT NULL,
			phone TEXT NOT NULL,
			belt TEXT NOT NULL DEFAULT '白' CHECK (belt IN ('白','黄','橙','绿','蓝','红','黑')),
			status TEXT NOT NULL DEFAULT '在籍' CHECK (status IN ('在籍','停训','退出'))
		);
		CREATE TABLE IF NOT EXISTS training_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('基础','对练','体能','套路')),
			duration INTEGER NOT NULL CHECK (duration > 0),
			FOREIGN KEY (student_id) REFERENCES students(id)
		);
		CREATE TABLE IF NOT EXISTS exam_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			date TEXT NOT NULL,
			target_belt TEXT NOT NULL CHECK (target_belt IN ('白','黄','橙','绿','蓝','红','黑')),
			coach TEXT NOT NULL,
			result TEXT NOT NULL CHECK (result IN ('通过','未通过')),
			FOREIGN KEY (student_id) REFERENCES students(id)
		);
	`)
	return err
}

func containsStr(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

func validateBelt(b string) error {
	if !containsStr(validBelts, b) {
		return fmt.Errorf("无效腰带: %s", b)
	}
	return nil
}

func validateStatus(s string) error {
	if !containsStr(validStatuses, s) {
		return fmt.Errorf("无效状态: %s", s)
	}
	return nil
}

func validateTrainingType(t string) error {
	if !containsStr(validTrainingTypes, t) {
		return fmt.Errorf("无效训练类型: %s", t)
	}
	return nil
}

func validateExamResult(r string) error {
	if !containsStr(validExamResults, r) {
		return fmt.Errorf("无效考试结果: %s", r)
	}
	return nil
}

func GetAllStudents() ([]Student, error) {
	rows, err := db.Query("SELECT id, nick, phone, belt, status FROM students ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Nick, &s.Phone, &s.Belt, &s.Status); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func GetRedBeltAndAbove() ([]Student, error) {
	rows, err := db.Query(`SELECT id, nick, phone, belt, status FROM students 
		WHERE belt IN ('红','黑') AND status = '在籍' ORDER BY belt DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Nick, &s.Phone, &s.Belt, &s.Status); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func GetStudentByID(id int64) (*Student, error) {
	row := db.QueryRow("SELECT id, nick, phone, belt, status FROM students WHERE id = ?", id)
	var s Student
	if err := row.Scan(&s.ID, &s.Nick, &s.Phone, &s.Belt, &s.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("学员不存在")
		}
		return nil, err
	}
	return &s, nil
}

func AddStudent(s *Student) error {
	if s.Nick == "" || s.Phone == "" {
		return errors.New("昵称和手机不能为空")
	}
	if err := validateBelt(s.Belt); err != nil {
		return err
	}
	if err := validateStatus(s.Status); err != nil {
		return err
	}
	res, err := db.Exec("INSERT INTO students(nick, phone, belt, status) VALUES(?,?,?,?)",
		s.Nick, s.Phone, s.Belt, s.Status)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func UpdateStudent(s *Student) error {
	if s.Nick == "" || s.Phone == "" {
		return errors.New("昵称和手机不能为空")
	}
	if err := validateBelt(s.Belt); err != nil {
		return err
	}
	if err := validateStatus(s.Status); err != nil {
		return err
	}
	res, err := db.Exec("UPDATE students SET nick=?, phone=?, belt=?, status=? WHERE id=?",
		s.Nick, s.Phone, s.Belt, s.Status, s.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("学员不存在")
	}
	return nil
}

func DeleteStudent(id int64) error {
	res, err := db.Exec("DELETE FROM students WHERE id=?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("学员不存在")
	}
	return nil
}

func GetAllTrainingRecords() ([]TrainingRecord, error) {
	rows, err := db.Query(`SELECT tr.id, tr.student_id, tr.date, tr.type, tr.duration, s.nick 
		FROM training_records tr LEFT JOIN students s ON tr.student_id = s.id 
		ORDER BY tr.date DESC, tr.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TrainingRecord
	for rows.Next() {
		var r TrainingRecord
		var nick sql.NullString
		if err := rows.Scan(&r.ID, &r.StudentID, &r.Date, &r.Type, &r.Duration, &nick); err != nil {
			return nil, err
		}
		if nick.Valid {
			r.Nick = nick.String
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func AddTrainingRecord(r *TrainingRecord) error {
	if _, err := GetStudentByID(r.StudentID); err != nil {
		return err
	}
	if r.Date == "" {
		return errors.New("训练日期不能为空")
	}
	if _, err := time.Parse("2006-01-02", r.Date); err != nil {
		return errors.New("日期格式应为YYYY-MM-DD")
	}
	if err := validateTrainingType(r.Type); err != nil {
		return err
	}
	if r.Duration <= 0 {
		return errors.New("时长必须为正整数")
	}
	res, err := db.Exec("INSERT INTO training_records(student_id, date, type, duration) VALUES(?,?,?,?)",
		r.StudentID, r.Date, r.Type, r.Duration)
	if err != nil {
		return err
	}
	r.ID, err = res.LastInsertId()
	return err
}

func GetMonthlyTrainingDuration(studentID int64, year, month int) (int, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	row := db.QueryRow(`SELECT COALESCE(SUM(duration), 0) FROM training_records 
		WHERE student_id = ? AND date LIKE ?`, studentID, prefix+"%")
	var total int
	err := row.Scan(&total)
	return total, err
}

func GetMonthlyTrainingTypeCount(year, month int) (map[string]int, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := db.Query(`SELECT type, COUNT(*) FROM training_records 
		WHERE date LIKE ? GROUP BY type`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int)
	for _, t := range validTrainingTypes {
		result[t] = 0
	}
	for rows.Next() {
		var t string
		var c int
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		result[t] = c
	}
	return result, rows.Err()
}

func GetAllExamRecords() ([]ExamRecord, error) {
	rows, err := db.Query(`SELECT er.id, er.student_id, er.date, er.target_belt, er.coach, er.result, s.nick 
		FROM exam_records er LEFT JOIN students s ON er.student_id = s.id 
		ORDER BY er.date DESC, er.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ExamRecord
	for rows.Next() {
		var r ExamRecord
		var nick sql.NullString
		if err := rows.Scan(&r.ID, &r.StudentID, &r.Date, &r.TargetBelt, &r.Coach, &r.Result, &nick); err != nil {
			return nil, err
		}
		if nick.Valid {
			r.Nick = nick.String
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func AddExamRecord(r *ExamRecord) error {
	if err := validateExamResult(r.Result); err != nil {
		return err
	}
	if r.Date == "" {
		return errors.New("考试日期不能为空")
	}
	if _, err := time.Parse("2006-01-02", r.Date); err != nil {
		return errors.New("日期格式应为YYYY-MM-DD")
	}
	if err := validateBelt(r.TargetBelt); err != nil {
		return err
	}
	if r.Coach == "" {
		return errors.New("主考教练不能为空")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT id, nick, phone, belt, status FROM students WHERE id = ?", r.StudentID)
	var s Student
	if err := row.Scan(&s.ID, &s.Nick, &s.Phone, &s.Belt, &s.Status); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("学员不存在")
		}
		return err
	}

	currentLevel := beltOrder[s.Belt]
	targetLevel := beltOrder[r.TargetBelt]
	if targetLevel != currentLevel+1 {
		return fmt.Errorf("目标腰带必须比当前高一级，当前%s,不能直接考%s", s.Belt, r.TargetBelt)
	}

	res, err := tx.Exec("INSERT INTO exam_records(student_id, date, target_belt, coach, result) VALUES(?,?,?,?,?)",
		r.StudentID, r.Date, r.TargetBelt, r.Coach, r.Result)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()

	if r.Result == "通过" {
		_, err = tx.Exec("UPDATE students SET belt = ? WHERE id = ?", r.TargetBelt, r.StudentID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
