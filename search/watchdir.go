package search

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type NotifyFile struct {
	watch     *fsnotify.Watcher
	logger    *zap.Logger
	indexFunc func(path string)
}

func NewNotifyFile(logger *zap.Logger, indexFunc func(path string)) *NotifyFile {
	w := new(NotifyFile)
	w.watch, _ = fsnotify.NewWatcher()
	w.indexFunc = indexFunc
	w.logger = logger
	return w
}

// 监控目录
func (this *NotifyFile) WatchDir(dir string) error {
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
			err = this.watch.Add(path)
			if err != nil {
				return err
			}
		} else {
			this.indexFunc(path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	go this.WatchEvent() //协程
	return nil
}

func (this *NotifyFile) WatchEvent() {
	for {
		select {
		case ev := <-this.watch.Events:
			{
				this.logger.Debug("watch Events", zap.String("Name", ev.Name), zap.String("Op", ev.Op))
				if ev.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("创建文件 : ", ev.Name)
					//获取新创建文件的信息，如果是目录，则加入监控中
					file, err := os.Stat(ev.Name)
					if err != nil {
						this.logger.Debug("watch Create Error", zap.String("Name", ev.Name), zap.Error(err))
						break
					}
					if file.IsDir() {
						this.watch.Add(ev.Name)
					} else {
						this.indexFunc(ev.Name)
					}
				}

				if ev.Op&fsnotify.Write == fsnotify.Write {
					//fmt.Println("写入文件 : ", ev.Name)
				}

				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Println("删除文件 : ", ev.Name)
					//如果删除文件是目录，则移除监控
					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						this.watch.Remove(ev.Name)
						fmt.Println("删除监控 : ", ev.Name)
					}
				}

				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					//如果重命名文件是目录，则移除监控 ,注意这里无法使用os.Stat来判断是否是目录了
					//因为重命名后，go已经无法找到原文件来获取信息了,所以简单粗爆直接remove
					fmt.Println("重命名文件 : ", ev.Name)
					this.watch.Remove(ev.Name)
				}
				if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					this.logger.Debug("watch Chmod", zap.String("Name", ev.Name))
				}
			}
		case err := <-this.watch.Errors:
			{
				this.logger.Debug("watch Errors", zap.Error(err))
				return
			}
		}
	}
}
