package main

import (
	"flag"

	"github.com/zhehuama/fsePressureTool/fse"
)

var maxCount int64

func main() {
	flag.Int64Var(&maxCount, "max", 10, "Maximum times the task is executed")
	flag.Parse()

	task := (&fse.TaskFactory{}).CreateTask()
	frame := fse.Frame{Task: task}
	frame.RunTask(maxCount)
}
