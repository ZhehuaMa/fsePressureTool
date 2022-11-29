package main

import (
	"github.com/zhehuama/fsePressureTool/fse"
)

func main() {
	task := (&fse.TaskFactory{}).CreateTask()
	frame := fse.Frame{Task: task}
	frame.RunTask(fse.MaxCount)
}
