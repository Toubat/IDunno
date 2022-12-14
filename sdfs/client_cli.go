package sdfs

import (
	"fmt"
	"mp4/api"
	"mp4/utils"

	"os"
	"strings"
	"time"
)

type SDFSClientCLI interface {
	Put(localFile string, sdfsFile string) error
	Get(localFile string, sdfsFile string) error
	Delete(sdfsFile string) error
	List(sdfsFile string) error
	Store() error
	GetVersions(localFile string, sdfsFile string, versions int) error
	PutDir(localDir string, sdfsDir string) error
	ValidateDir(sdfsDir string) error
}

func (c *SDFSClient) Put(localFile string, sdfsFile string) error {
	data, err := c.ReadLocalFile(localFile)
	if err != nil {
		c.Printf("Error reading file %s\n", localFile)
		return err
	}

	writeId := api.WriteId{
		Ip:         c.SDFSServer.Ring.GetIp(),
		Port:       c.SDFSServer.Ring.GetPort(),
		CreateTime: api.CurrentTimestamp(),
	}
	task := SDFSPutTask{
		LocalFile: localFile,
		SDFSFile:  sdfsFile,
		Data:      data,
		WriteId:   &writeId,
	}

	now := time.Now()
	c.Println("Sending local file " + localFile + " to SDFS...")

	for {
		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		if res != nil {
			break
		}
	}

	c.Println("Successfully put file " + localFile + " to SDFS")
	c.CalculateTime(SDFSPutTask{SDFSFile: sdfsFile}, now)
	return nil
}

func (c *SDFSClient) Get(localFile string, sdfsFile string, version int32) error {
	task := SDFSGetTask{
		LocalFile: localFile,
		SDFSFile:  strings.Replace(sdfsFile, "/", ":", -1),
		Version:   version,
	}

	now := time.Now()
	c.Println("Receving SDFS file " + sdfsFile + "...")

	for {
		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res == nil {
			continue
		}

		if res.(SDFSGetTaskResult).GetStatus() == api.ResponseStatus_NOT_FOUND {
			c.Printf("File %s does not exist in SDFS\n", sdfsFile)
			return nil
		}

		if err := c.WriteLocalFile(localFile, res.(SDFSGetTaskResult).Data); err != nil {
			c.Printf("Error writing to file %s\n", localFile)
			return err
		}

		break
	}

	c.Println("Successfully get SDFS file " + sdfsFile + " to local " + localFile)
	c.CalculateTime(SDFSGetTask{SDFSFile: sdfsFile}, now)
	return nil
}

func (c *SDFSClient) Delete(sdfsFile string) error {
	task := SDFSDeleteTask{
		SDFSFile: sdfsFile,
	}

	now := time.Now()
	c.Println("Deleting SDFS file " + sdfsFile + "...")

	for {
		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res != nil {
			break
		}
	}

	c.CalculateTime(SDFSDeleteTask{SDFSFile: sdfsFile}, now)
	return nil
}

func (c *SDFSClient) List(sdfsFile string) error {
	now := time.Now()

	task := SDFSListTask{
		SDFSFile: sdfsFile,
	}

	for {
		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res == nil {
			continue
		}

		if res.(SDFSListTaskResults).GetStatus() == api.ResponseStatus_NOT_FOUND {
			c.Printf("File %s does not exist in SDFS\n", sdfsFile)
			return fmt.Errorf("file %s does not exist in SDFS", sdfsFile)
		}

		c.Printf("File %s is stored at: \n", sdfsFile)
		for _, r := range res.(SDFSListTaskResults).Results {
			c.Printf("	%s:%d\n", r.Ip, r.Port)
		}

		break
	}

	c.CalculateTime(SDFSListTask{SDFSFile: sdfsFile}, now)
	return nil
}

func (c *SDFSClient) Store() error {
	now := time.Now()

	for filename, versions := range *c.SDFSServer.FileTable {
		c.Printf("file-name: %s, num-versions: %d\n", filename, versions.Len())
	}

	c.CalculateTime(SDFSStoreTask{}, now)
	return nil
}

func (c *SDFSClient) GetVersions(localFile string, sdfsFile string, versions int) error {
	now := time.Now()
	queue := utils.NewQueue[SDFSGetTask]()

	for i := 1; i <= versions; i++ {
		queue.Push(SDFSGetTask{
			LocalFile: localFile,
			SDFSFile:  sdfsFile,
			Version:   int32(i),
		})
	}

	file, err := os.Create(c.GetLocalFilePath(localFile))
	if err != nil {
		c.Printf("Error creating file %s\n", localFile)
		return err
	}
	defer file.Close()

	for !queue.Empty() {
		task := queue.Top()

		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res == nil {
			continue
		}

		if res.(SDFSGetTaskResult).GetStatus() == api.ResponseStatus_NOT_FOUND {
			break
		}

		if _, err := file.Write(res.(SDFSGetTaskResult).Data); err != nil {
			c.Printf("Error writing to file %s\n", localFile)
			return err
		}
		if _, err := file.WriteString("\n"); err != nil {
			c.Printf("Error writing to file %s\n", localFile)
			return err
		}

		c.Printf("Successfully get version %d of file %s\n", task.Version, task.SDFSFile)
		queue.Pop()
	}

	c.Printf("Successfully get %d versions of file %s\n", versions, sdfsFile)
	c.CalculateTime(SDFSGetTask{SDFSFile: sdfsFile}, now)
	return nil
}

func (c *SDFSClient) PutDir(localDir string, sdfsDir string) error {
	now := time.Now()
	dir := c.GetLocalFilePath("") + localDir + "/"
	files, err := os.ReadDir(dir)
	if err != nil {
		c.Printf("Error reading directory %s\n", dir)
		return err
	}

	writeId := api.WriteId{
		Ip:         c.SDFSServer.Ring.GetIp(),
		Port:       c.SDFSServer.Ring.GetPort(),
		CreateTime: api.CurrentTimestamp(),
	}

	queue := utils.NewQueue[SDFSPutTask]()
	sdfsFiles := make([]string, 0)
	for _, file := range files {
		// remove evil file created by mac os system internally
		if file.Name() == ".DS_Store" {
			continue
		}

		queue.Push(SDFSPutTask{
			LocalFile: fmt.Sprintf("%s/%s", localDir, file.Name()),
			SDFSFile:  fmt.Sprintf("%s:%s", sdfsDir, file.Name()),
			WriteId:   &writeId,
		})
		sdfsFiles = append(sdfsFiles, fmt.Sprintf("%s:%s", sdfsDir, file.Name()))
	}

	i := 0
	for !queue.Empty() {
		task := queue.Top()
		data, err := c.ReadLocalFile(task.LocalFile)
		if err != nil {
			c.Printf("Error reading file %s\n", task.LocalFile)
			continue
		}

		res, err := c.ExecuteTask(SDFSPutTask{
			LocalFile: task.LocalFile,
			SDFSFile:  task.SDFSFile,
			WriteId:   task.WriteId,
			Data:      data,
		})
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		queue.Pop()
		i++

		if i%5 == 0 {
			// print percent round to 2 decimal places
			c.Printf("Progress: %.2f%%\r", float64(i)/float64(len(sdfsFiles))*100)
		}
	}

	dirFileTask := SDFSPutTask{
		LocalFile: localDir,
		SDFSFile:  sdfsDir,
		WriteId:   &writeId,
		Data:      []byte(strings.Join(sdfsFiles, "\n")),
	}

	for {
		res, err := c.ExecuteTask(dirFileTask)
		if err != nil {
			c.HandleTaskFailure(dirFileTask, err)
			return err
		}
		if res != nil {
			break
		}
	}

	c.Printf("Successfully put %s directory to SDFS\n", localDir)
	c.CalculateTime(SDFSPutTask{SDFSFile: sdfsDir}, now)
	return nil
}

func (c *SDFSClient) ValidateDir(sdfsDir string) error {
	localFile := utils.CreateTempFilename()

	err := c.Get(localFile, sdfsDir, LATEST_VERSION)
	if err != nil {
		c.Println("Error getting directory file " + sdfsDir + " from SDFS")
		return err
	}
	defer c.DeleteLocalFile(localFile)

	dirFile, err := os.ReadFile(c.GetLocalFilePath(localFile))
	if err != nil {
		c.Println("Error reading directory file " + localFile)
		return err
	}

	tasks := utils.NewQueue[SDFSListTask]()
	for _, sdfsFile := range strings.Split(string(dirFile), "\n") {
		tasks.Push(SDFSListTask{
			SDFSFile: sdfsFile,
		})
	}
	c.Println("Validating SDFS files in directory " + sdfsDir + "...")

	i := 0
	totalTasks := tasks.Len()
	for !tasks.Empty() {
		task := tasks.Top()
		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if res.(SDFSListTaskResults).GetStatus() == api.ResponseStatus_NOT_FOUND {
			c.Printf("File %s not found in SDFS\n", task.SDFSFile)
			continue
		}

		tasks.Pop()
		i++

		if i%5 == 0 {
			// print percent round to 2 decimal places
			c.Printf("Progress: %.2f%%\r", float64(i)/float64(totalTasks)*100)
		}
	}

	c.Printf("Successfully validated directory %s: all files have exactly 4 replicas\n", sdfsDir)
	return nil
}

func (c *SDFSClient) DeleteDir(sdfsDir string) error {
	localFile := utils.CreateTempFilename()

	err := c.Get(localFile, sdfsDir, LATEST_VERSION)
	if err != nil {
		c.Println("Error getting directory file " + sdfsDir + " from SDFS")
		return err
	}
	defer c.DeleteLocalFile(localFile)

	dirFile, err := os.ReadFile(c.GetLocalFilePath(localFile))
	if err != nil {
		c.Println("Error reading directory file " + localFile)
		return err
	}

	tasks := utils.NewQueue[SDFSDeleteTask]()
	for _, sdfsFile := range strings.Split(string(dirFile), "\n") {
		tasks.Push(SDFSDeleteTask{
			SDFSFile: sdfsFile,
		})
	}
	tasks.Push(SDFSDeleteTask{
		SDFSFile: sdfsDir,
	})

	c.Println("Deleting SDFS files in directory " + sdfsDir + "...")
	i := 0
	totalTasks := tasks.Len()
	for !tasks.Empty() {
		task := tasks.Top()
		res, err := c.ExecuteTask(task)
		if err != nil {
			c.HandleTaskFailure(task, err)
			return err
		}
		// res is nil if ring not converged
		if res == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		tasks.Pop()
		i++

		if i%5 == 0 {
			// print percent round to 2 decimal places
			c.Printf("Progress: %.2f%%\r", float64(i)/float64(totalTasks)*100)
		}
	}

	c.Printf("Successfully deleted directory %s\n", sdfsDir)
	return nil
}
