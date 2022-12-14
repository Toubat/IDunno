package sdfs

import (
	"fmt"
	"os"
)

type SDFSClientFS interface {
	ReadLocalFile(localFile string) ([]byte, error)
	WriteLocalFile(localFile string, data []byte) error
	GetFileSize(localFile string) (int64, error)
	ClearFiles()
}

// Read local file
func (c *SDFSClient) ReadLocalFile(localFile string) ([]byte, error) {
	size, err := c.GetFileSize(localFile)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(c.GetLocalFilePath(localFile))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, size)
	n, err := file.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:n], nil
}

func (c *SDFSClient) WriteLocalFile(localFile string, data []byte) error {
	// truncate, create, write, close
	file, err := os.OpenFile(c.GetLocalFilePath(localFile), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *SDFSClient) DeleteLocalFile(localFile string) error {
	return os.Remove(c.GetLocalFilePath(localFile))
}

// Get file size in bytes
func (c *SDFSClient) GetFileSize(localFile string) (int64, error) {
	file, err := os.Stat(c.GetLocalFilePath(localFile))
	if err != nil {
		return 0, err
	}

	return file.Size(), nil
}

// Remove files from local storage
func (c *SDFSClient) ClearFiles() {
	dir := c.GetLocalFilePath("")

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		os.Remove(c.GetLocalFilePath(file.Name()))
	}
}

func (c *SDFSClient) GetLocalFilePath(localFile string) string {
	return fmt.Sprintf("../data/%s/%s", c.SDFSServer.Ring.Address(), localFile)
}
