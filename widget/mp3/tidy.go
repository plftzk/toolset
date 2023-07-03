package mp3

import (
	"errors"
	"github.com/bogem/id3v2"
	"okapp/utils"
	"os"
	"path/filepath"
)

type TidyResult struct {
	Error      error
	FileTotal  int
	ErrorTotal int
	AudioTotal int
}

func reset(mp3Path string) (err error) {
	var mp3File *id3v2.Tag
	// 打开MP3文件
	mp3File, err = id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer mp3File.Close()
	for k, _ := range mp3File.AllFrames() {
		mp3File.DeleteFrames(k)
	}
	mp3File.DeleteAllFrames()
	comment := id3v2.CommentFrame{
		Encoding:    id3v2.Encoding{},
		Language:    "zho",
		Description: "",
		Text:        "",
	}
	mp3File.AddCommentFrame(comment)
	// 保存修改后的标签
	err = mp3File.Save()
	return
}

func BatchRemoveMp3Tag(audioDir string) (tidyResult TidyResult) {
	if ok, e := utils.IsDirectory(audioDir); !ok {
		if e != nil {
			tidyResult.Error = e
			return
		}
		tidyResult.Error = errors.New("当前路径不是文件夹")
		return
	}
	err := filepath.Walk(audioDir, func(fp string, info os.FileInfo, err error) error {
		tidyResult.FileTotal++
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(fp)
			if ext == ".mp3" {
				tidyResult.AudioTotal++
				e := reset(fp)
				if e != nil {
					tidyResult.ErrorTotal++
				}
			}
		}
		return nil
	})
	if err != nil {
		tidyResult.Error = err
	}
	return tidyResult
}
