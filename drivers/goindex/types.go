package goindex

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type File struct {
	MimeType     string     `json:"mimeType"`
	Size         string     `json:"size"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	ModifiedTime *time.Time `json:"modifiedTime"`
}

type FolderResp struct {
	//ResCode    int    `json:"res_code"`
	//ResMessage string `json:"res_message"`
	Data struct {
		Files []File `json:"files"`
	} `json:"data"`
	NextPageToken string `json:"nextPageToken"`
}

type FolderRespFloat struct {
	//ResCode    int    `json:"res_code"`
	//ResMessage string `json:"res_message"`
	Data struct {
		Files []File `json:"files"`
	} `json:"data"`
	NextPageToken string `json:"nextPageToken"`
}

func fileToObj(f File) *model.ObjThumb {
	//convert Size(in string) to SizeInt(in int64)
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	file := &model.ObjThumb{
		Object: model.Object{
			Name:     f.Name,
			Modified: *f.ModifiedTime,
			Size:     size,
			IsFolder: f.MimeType == "application/vnd.google-apps.folder",
		},
		Thumbnail: model.Thumbnail{},
	}
	return file
}
