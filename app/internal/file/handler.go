package file

import (
	"encoding/json"
	"fmt"
	"github.com/Vladislav747/minio-project/internal/apperror"
	"github.com/Vladislav747/minio-project/pkg/logging"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	filesURL = "/api/files"
	fileURL  = "/api/files/:id"
)

type Handler struct {
	Logger      logging.Logger
	FileService Service
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, fileURL, apperror.Middleware(h.GetFile))
	router.HandlerFunc(http.MethodGet, filesURL, apperror.Middleware(h.GetFilesByNoteUUID))
	router.HandlerFunc(http.MethodPost, filesURL, apperror.Middleware(h.CreateFile))
	router.HandlerFunc(http.MethodDelete, fileURL, apperror.Middleware(h.DeleteFile))
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("GetFile")

	h.Logger.Debug("get note_uuid from URL")

	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}

	h.Logger.Debug("get field from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	fileId := params.ByName("id")

	f, err := h.FileService.GetFile(r.Context(), noteUUID, fileId)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Name))
	//Отдаем с таким же заголовком которые запрашивали
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	w.Write(f.Bytes)

	return nil
}

// Получить файлы по заметки по UUID
func (h *Handler) GetFilesByNoteUUID(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("GetFilesByNoteUUID")
	w.Header().Set("Content-Type", "form/json")

	h.Logger.Debug("get note_uuid from URL")

	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}

	files, err := h.FileService.GetFilesByNoteUUID(r.Context(), noteUUID)
	if err != nil {
		return err
	}

	//Все файлы заворачиваю тут их несколько
	fileBytes, err := json.Marshal(files)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)

	return nil
}

func (h *Handler) CreateFile(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("CreateFile")
	w.Header().Set("Content-Type", "form/json")

	// TODO maximum file size
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return err
	}
	h.Logger.Debug("decode create upload fileInfo dto")

	//Если не пришел заголовок с файлом
	files, ok := r.MultipartForm.File["file"]
	if !ok || len(files) == 0 {
		return apperror.BadRequestError("file required")
	}

	fileInfo := files[0]
	fileReader, err := fileInfo.Open()
	dto := CreateFileDTO{
		Name:   fileInfo.Filename,
		Size:   fileInfo.Size,
		Reader: fileReader,
	}

	err = h.FileService.Create(r.Context(), r.Form.Get("node_uuid"), dto)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) error {
	h.Logger.Info("DeleteFile")
	w.Header().Set("Content-Type", "application/json")

	h.Logger.Debug("get fileId from context")
	params := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	fileId := params.ByName("id")

	noteUUID := r.URL.Query().Get("note_uuid")
	if noteUUID == "" {
		return apperror.BadRequestError("note_uuid query parameter is required")
	}

	err := h.FileService.Delete(r.Context(), fileId, noteUUID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
