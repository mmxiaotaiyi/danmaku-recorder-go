package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	baseDir            = "danmaku"
	maxMessagesPerFile = 1000
	fileBufferSize     = 16 * 1024
)

type fileHandle struct {
	file  *os.File
	buf   *bufio.Writer
	count int
}

func (h *fileHandle) writeLine(data []byte) error {
	if _, err := h.buf.Write(data); err != nil {
		return err
	}
	return h.buf.WriteByte('\n')
}

func (h *fileHandle) close() error {
	err := h.buf.Flush()
	if err == nil {
		err = h.file.Sync()
	}
	if closeErr := h.file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}

func (ss *StorageService) initDirectory() error {
	dir := filepath.Join(baseDir, strconv.Itoa(ss.roomID), time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	ss.currentDir = dir
	return nil
}

func (ss *StorageService) openNew(cmd string) (*fileHandle, error) {
	fileName := fmt.Sprintf("%s_%d.jsonl", cmd, time.Now().Unix())
	file, err := os.OpenFile(filepath.Join(ss.currentDir, fileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}

	h := &fileHandle{
		file: file,
		buf:  bufio.NewWriterSize(file, fileBufferSize),
	}
	if old := ss.handles[cmd]; old != nil {
		if err := old.close(); err != nil {
			log.Errorf("storage close old file error: room=%d cmd=%s err=%v", ss.roomID, cmd, err)
		}
	}
	ss.handles[cmd] = h
	return h, nil
}

func (ss *StorageService) appendLine(cmd string, data []byte) {
	if ss.filesClosed {
		log.Errorf("storage write error: room=%d cmd=%s err=files already closed", ss.roomID, cmd)
		return
	}

	h := ss.handles[cmd]
	if h == nil || h.count >= maxMessagesPerFile {
		var err error
		h, err = ss.openNew(cmd)
		if err != nil {
			log.Errorf("storage write error: room=%d cmd=%s err=%v", ss.roomID, cmd, err)
			return
		}
	}

	if err := h.writeLine(data); err != nil {
		log.Errorf("storage write error: room=%d cmd=%s err=%v", ss.roomID, cmd, err)
		return
	}
	h.count++
}

func (ss *StorageService) flushFiles() {
	if ss.filesClosed {
		return
	}
	for cmd, h := range ss.handles {
		if err := h.buf.Flush(); err != nil {
			log.Errorf("storage flush error: room=%d cmd=%s err=%v", ss.roomID, cmd, err)
		}
	}
}

func (ss *StorageService) closeFiles() {
	if ss.filesClosed {
		return
	}
	ss.filesClosed = true
	for cmd, h := range ss.handles {
		if err := h.close(); err != nil {
			log.Errorf("storage close file error: room=%d cmd=%s err=%v", ss.roomID, cmd, err)
		}
	}
	ss.handles = make(map[string]*fileHandle)
}
