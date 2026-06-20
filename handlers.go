package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func parseBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func getIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, nil
	}
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}

type beltAndMetaResponse struct {
	Belts        []string `json:"belts"`
	Statuses     []string `json:"statuses"`
	TrainingType []string `json:"training_types"`
	ExamResults  []string `json:"exam_results"`
}

func HandleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, beltAndMetaResponse{
		Belts:        validBelts,
		Statuses:     validStatuses,
		TrainingType: validTrainingTypes,
		ExamResults:  validExamResults,
	})
}

func HandleStudents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		students, err := GetAllStudents()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, students)
	case http.MethodPost:
		var s Student
		if err := parseBody(r, &s); err != nil {
			writeError(w, http.StatusBadRequest, "请求格式错误")
			return
		}
		if s.Belt == "" {
			s.Belt = "白"
		}
		if s.Status == "" {
			s.Status = "在籍"
		}
		if err := AddStudent(&s); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, s)
	default:
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

func HandleStudentByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, "无效ID")
		return
	}
	switch r.Method {
	case http.MethodGet:
		s, err := GetStudentByID(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, s)
	case http.MethodPut:
		var s Student
		if err := parseBody(r, &s); err != nil {
			writeError(w, http.StatusBadRequest, "请求格式错误")
			return
		}
		s.ID = id
		if err := UpdateStudent(&s); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, s)
	case http.MethodDelete:
		if err := DeleteStudent(id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})
	default:
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

func HandleRedBeltStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	students, err := GetRedBeltAndAbove()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, students)
}

func HandleTrainingRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		records, err := GetAllTrainingRecords()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, records)
	case http.MethodPost:
		var r2 TrainingRecord
		if err := parseBody(r, &r2); err != nil {
			writeError(w, http.StatusBadRequest, "请求格式错误")
			return
		}
		if err := AddTrainingRecord(&r2); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, r2)
	default:
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

func HandleMonthlyTrainingDuration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	q := r.URL.Query()
	idStr := q.Get("student_id")
	studentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效student_id")
		return
	}
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if y := q.Get("year"); y != "" {
		if yv, err := strconv.Atoi(y); err == nil {
			year = yv
		}
	}
	if m := q.Get("month"); m != "" {
		if mv, err := strconv.Atoi(m); err == nil && mv >= 1 && mv <= 12 {
			month = mv
		}
	}
	total, err := GetMonthlyTrainingDuration(studentID, year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{
		"student_id": studentID,
		"year":       year,
		"month":      month,
		"total_minutes": total,
	})
}

func HandleMonthlyTrainingTypeCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	q := r.URL.Query()
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if y := q.Get("year"); y != "" {
		if yv, err := strconv.Atoi(y); err == nil {
			year = yv
		}
	}
	if m := q.Get("month"); m != "" {
		if mv, err := strconv.Atoi(m); err == nil && mv >= 1 && mv <= 12 {
			month = mv
		}
	}
	counts, err := GetMonthlyTrainingTypeCount(year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{
		"year":   year,
		"month":  month,
		"counts": counts,
	})
}

func HandleExamRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		records, err := GetAllExamRecords()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, records)
	case http.MethodPost:
		var e ExamRecord
		if err := parseBody(r, &e); err != nil {
			writeError(w, http.StatusBadRequest, "请求格式错误")
			return
		}
		if err := AddExamRecord(&e); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, e)
	default:
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}
