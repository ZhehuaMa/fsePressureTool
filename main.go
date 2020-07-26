package main

import (
    "flag"
    "fmt"
    "github.com/zhehuama/fse_go/tool"
    "os"
)

const (
    entityTask  = "entityTask"
    searchTask  = "searchTask"
    compareTask = "compareTask"
)

const (
    uuid = "uuid"
    num  = "num"
)

var (
    addr      string
    repo      string
    taskType  string
    idType    string
    qps       int
    threadNum int
    maxCount  int
)

func initFlags() {
    flag.StringVar(&addr, "addr", "127.0.0.1:30800", "FSE address, {IP}:{PORT}")
    flag.StringVar(&repo, "repo", "repo", "Repo name")
    flag.StringVar(&taskType, "type", entityTask,
        "Task type: 'entityTask' for adding features, 'searchTask' for searching features, 'compareTask' for comparing task")
    flag.StringVar(&idType, "id", uuid, "ID type: 'uuid' for UUID, 'num' for number sequence, starting from 0")
    flag.IntVar(&qps, "qps", 1, "QPS")
    flag.IntVar(&threadNum, "t", 1, "The number of threads")
    flag.IntVar(&maxCount, "max", 10, "Maximum times the task is executed")

    flag.Parse()

    if taskType != entityTask && taskType != searchTask && taskType != compareTask {
        fmt.Fprintln(os.Stderr, "Unknown task type", taskType)
        os.Exit(-1)
    }
    if idType != uuid && idType != num {
        fmt.Fprintln(os.Stderr, "Unknown ID type", idType)
        os.Exit(-1)
    }
}

func main() {
    initFlags()

    switch taskType {
    case entityTask:
    case searchTask:
    case compareTask:
    }
}
