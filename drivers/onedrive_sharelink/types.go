package onedrive_sharelink

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type FolderResp struct {
	//ResCode    int    `json:"res_code"`
	//ResMessage string `json:"res_message"`
	Data struct {
		Legacy struct {
			RenderListData struct {
				ListData struct {
					Items []Item `json:"Row"`
				} `json:"ListData"`
			} `json:"renderListDataAsStream"`
		} `json:"legacy"`
	} `json:"data"`
}

type Item struct {
	ObjType      string    `json:"FSObjType"`
	Name         string    `json:"FileLeafRef"`
	ModifiedTime time.Time `json:"Modified."`
	Size         string    `json:"File_x0020_Size"`
	Id           string    `json:"UniqueId"`
}

func fileToObj(f Item) *model.ObjThumb {
	//convert Size(in string) to SizeInt(in int64)
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	objtype, _ := strconv.Atoi(f.ObjType)
	file := &model.ObjThumb{
		Object: model.Object{
			Name:     f.Name,
			Modified: f.ModifiedTime,
			Size:     size,
			IsFolder: objtype == 1,
			ID:       f.Id,
		},
		Thumbnail: model.Thumbnail{},
	}
	return file
}

type GraphQLNEWRequest struct {
	ListData struct {
		NextHref string `json:"NextHref"`
		Row      []Item `json:"Row"`
	} `json:"ListData"`
}

type GraphQLRequest struct {
	Data struct {
		Legacy struct {
			RenderListDataAsStream struct {
				ListData struct {
					NextHref string `json:"NextHref"`
					Row      []Item `json:"Row"`
				} `json:"ListData"`
				ViewMetadata struct {
					ListViewXml string `json:"ListViewXml"`
				} `json:"ViewMetadata"`
			} `json:"renderListDataAsStream"`
		} `json:"legacy"`
	} `json:"data"`
}
