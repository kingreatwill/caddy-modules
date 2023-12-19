package search

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type NotifyFile struct {
	watch     *fsnotify.Watcher
	logger    *zap.Logger
	indexFunc func(path string, remove bool)
}

func NewNotifyFile(logger *zap.Logger, indexFunc func(path string, remove bool)) *NotifyFile {
	w := new(NotifyFile)
	w.watch, _ = fsnotify.NewWatcher()
	w.indexFunc = indexFunc
	w.logger = logger
	return w
}

// WatchDir 监控目录
func (nf *NotifyFile) WatchDir(dir string) error {
	//通过Walk来遍历目录下的所有子目录
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// handle possible path err, just in case...
			return err
		}
		//判断是否为目录，监控目录,目录下文件也在监控范围内，不需要加
		if d.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			err = nf.watch.Add(path)
			if err != nil {
				return err
			}
		} else {
			nf.indexFunc(path, false)
		}
		return nil
	})
	if err != nil {
		return err
	}
	go nf.WatchEvent() //协程
	return nil
}

func (nf *NotifyFile) WatchEvent() {
	for {
		select {
		case event, ok := <-nf.watch.Events:
			{
				if !ok {
					return
				}
				nf.logger.Debug("watch Events",
					zap.String("Event", event.String()),
				)
				if event.Has(fsnotify.Create) {
					file, err := os.Stat(event.Name)
					if err != nil {
						nf.logger.Debug("watch Create Stat Error",
							zap.String("Name", event.Name), zap.Error(err))
						break
					}
					if file.IsDir() {
						err = nf.watch.Add(event.Name)
						if err != nil {
							nf.logger.Debug("watch Create Watch Add Error",
								zap.String("Name", event.Name), zap.Error(err))
						}
					} else {
						nf.indexFunc(event.Name, false)
					}
				}
				if event.Has(fsnotify.Write) {
					// 修改文件 或者 目录中新增和删除文件
					nf.indexFunc(event.Name, false)
				}
				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					err := nf.watch.Remove(event.Name)
					if err != nil {
						nf.logger.Debug("watch Create Watch Remove/Rename Error",
							zap.String("Name", event.Name), zap.Error(err))
					}
					nf.indexFunc(event.Name, true)
				}
			}
		case err, ok := <-nf.watch.Errors:
			{
				if !ok {
					return
				}
				nf.logger.Debug("watch Errors", zap.Error(err))
				return
			}
		}
	}
}
