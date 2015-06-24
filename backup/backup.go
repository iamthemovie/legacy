package backup

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/rlmcpherson/s3gof3r"
)

// Backup ...
// All thie things
type Backup struct {
	FileSystemRoot    string
	S3StreamBucket    *s3gof3r.Bucket
	S3Path            string
	RemoveAfterUpload bool
	State             *State
	TaskQueue         chan *BackupUploadTask
	ResultsQueue      chan *BackupUploadTask
}

// State ...
type State struct {
	TotalUploadSize int64
	TotalFileCount  int32
	UploadedSize    int64
	UploadedCount   int32
}

// BackupUploadTask ...
type BackupUploadTask struct {
	FilePath string
	FileInfo *os.FileInfo
	Failed   bool
}

func (backup *Backup) Run() {
	backup.State = &State{
		TotalUploadSize: 0,
		TotalFileCount:  0,
		UploadedSize:    0,
		UploadedCount:   0,
	}

	log.Println("[backup] Computing upload size.")
	filepath.Walk(backup.FileSystemRoot, backup.ComputeBackupSize)
	log.Println(fmt.Sprintf("[backup] \n %s \n %d files (%d bytes)", backup.FileSystemRoot, backup.State.TotalFileCount, backup.State.TotalUploadSize))
	backup.TaskQueue = make(chan *BackupUploadTask, backup.State.TotalFileCount)
	backup.ResultsQueue = make(chan *BackupUploadTask, backup.State.TotalFileCount)

	// Selectable initial
	for i := 0; i < 3; i++ {
		go backup.ProcessUploads()
	}

	filepath.Walk(backup.FileSystemRoot, backup.AddToUploadQueue)
	close(backup.TaskQueue)
	for i := 1; i <= int(backup.State.TotalFileCount); i++ {
		result := <-backup.ResultsQueue
		if result.Failed {
			log.Println("Failed!!" + result.FilePath)
			continue
		}

		log.Println(
			fmt.Sprintf(
				"Uploaded %d/%d files. %d/%d bytes",
				atomic.LoadInt32(&backup.State.UploadedCount),
				backup.State.TotalFileCount,
				atomic.LoadInt64(&backup.State.UploadedSize),
				backup.State.TotalUploadSize))
	}

	close(backup.ResultsQueue)
}

func (backup *Backup) AddToUploadQueue(p string, f os.FileInfo, err error) error {
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	if f.IsDir() {
		return nil
	}

	log.Println(("Uploading " + f.Name() + " (" + strconv.FormatInt(f.Size(), 10) + " bytes)"))
	backup.TaskQueue <- &BackupUploadTask{
		FilePath: p,
		FileInfo: &f,
	}

	return nil
}

// ProcessUploads
func (backup *Backup) ProcessUploads() {
	for task := range backup.TaskQueue {
		state := backup.State
		if (*task.FileInfo).IsDir() {
			backup.ResultsQueue <- task
			continue
		}

		// Open ze file
		file, err := os.Open(task.FilePath)
		if err != nil {
			// Do something more ellaborate here.
			log.Println(err.Error())
			backup.ResultsQueue <- task
			continue
		}

		// Dry Run
		//	if false {
		s3FilePath := path.Join(backup.S3Path, (*task.FileInfo).Name())
		s3Writer, err := backup.S3StreamBucket.PutWriter(s3FilePath, nil, nil)
		if err != nil {
			// Do something more ellaborate here.
			log.Println(err.Error())
			backup.ResultsQueue <- task
			continue
		}

		io.Copy(s3Writer, file)
		s3Writer.Close()
		file.Close()

		atomic.AddInt64(&state.UploadedSize, (*task.FileInfo).Size())
		atomic.AddInt32(&state.UploadedCount, 1)

		if backup.RemoveAfterUpload {
			os.Remove(task.FilePath)
		}

		backup.ResultsQueue <- task
	}
}

func (backup *Backup) ComputeBackupSize(p string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}

	backup.State.TotalFileCount++
	backup.State.TotalUploadSize += f.Size()
	return nil
}
