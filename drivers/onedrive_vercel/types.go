package onedrive_vercel

import (
	"time"
)

type Obj struct {
	Id                   string        `json:"id"`
	Name                 string        `json:"name"`
	Size                 int64         `json:"size"`
	LastModifiedDateTime *time.Time    `json:"lastModifiedDateTime"`
	Folder               *FolderDetail `json:"folder"`
	File                 *FileDetail   `json:"file"`
}

type FolderDetail struct {
	ChildCount int `json:"childCount"`
}

type FileDetail struct {
	MimeType string `json:"mimeType"`
	Hashes   Hash   `json:"hashes"`
}

type Hash struct {
	QuickXorHash string `json:"quickXorHash"`
}

type FolderResp struct {
	//ResCode    int    `json:"res_code"`
	//ResMessage string `json:"res_message"`
	Folder struct {
		Context string `json:"@odata.context"`
		Value   []Obj  `json:"value"`
	} `json:"folder"`
	Next string `json:"next"`
}
