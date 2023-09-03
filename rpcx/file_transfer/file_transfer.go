package file_transfer

import (
	"context"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"log/slog"
	"os"
	"time"
)

const (
	incidentFilename = "file.txt"
)

type Request struct {
	IncidentID int64
}

type Response struct {
	Filename string
	Filesize int64
}

type FileService struct {
	xClient client.XClient
}

func NewFileService(xClient client.XClient) *FileService {
	return &FileService{xClient: xClient}
}

func getFilenameByIncident(_ int64) (string, error) {
	return incidentFilename, nil
}

func (s *FileService) GetIncidentFile(_ context.Context, arg Request, reply *Response) error {
	filename, err := getFilenameByIncident(arg.IncidentID)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	stat, err := os.Stat(filename)
	if err != nil {
		slog.Warn("file does not exist", slog.String("filename", filename))
		return fmt.Errorf("file for incident %#v not found", filename)
	}

	err = s.xClient.SendFile(context.Background(), filename, 0, map[string]string{ // TODO: set rate limit?
		"my_name":         "client1",
		"send_start_time": time.Now().Format(time.StampMilli),
	})
	if err != nil {
		slog.Error(err.Error())
	} else {
		reply = &Response{
			Filename: filename,
			Filesize: stat.Size(),
		}
		slog.Info("send ok", slog.String("send_end_time", time.Now().Format(time.StampMilli)))
	}

	return err
}
