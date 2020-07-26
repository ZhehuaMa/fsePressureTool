package main

import (
    "flag"
    "fmt"
    "github.com/zhehuama/fsePressureTool/tool"
    "os"
    "strings"
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
    addr        string
    repo        string
    taskType    string
    idType      string
    startTimeMs int64
    endTimeMs   int64
    maxCount    int64
    topk        int
    featLen     int
    LocNum      int
    qps         int
    threadNum   int
    repos       []string
)

func initFlags() {
    flag.StringVar(&addr, "addr", "127.0.0.1:30800", "FSE address, {IP}:{PORT}")
    flag.StringVar(&repo, "repo", "repo", "Repo name")
    flag.StringVar(&taskType, "type", entityTask,
        "Task type: 'entityTask' for adding features, 'searchTask' for searching features, 'compareTask' for comparing task")
    flag.StringVar(&idType, "id", uuid, "ID type: 'uuid' for UUID, 'num' for number sequence, starting from 0")
    flag.Int64Var(&startTimeMs, "st", 0, "The start time in millisecond, only used in 'entityTask' type")
    flag.Int64Var(&endTimeMs, "et", 0, "The end time in millisecond, only used in 'entityTask' type")
    flag.IntVar(&topk, "topk", 3, "Top K")
    flag.IntVar(&featLen, "len", 384, "The length of feature")
    flag.IntVar(&LocNum, "loc", 1, "The number of locations, only used in 'entityTask' type")
    flag.IntVar(&qps, "qps", 1, "QPS")
    flag.IntVar(&threadNum, "t", 1, "The number of threads")
    flag.Int64Var(&maxCount, "max", 10, "Maximum times the task is executed")

    flag.Parse()

    if taskType != entityTask && taskType != searchTask && taskType != compareTask {
        fmt.Fprintln(os.Stderr, "Unknown task type", taskType)
        os.Exit(-1)
    }
    if idType != uuid && idType != num {
        fmt.Fprintln(os.Stderr, "Unknown ID type", idType)
        os.Exit(-1)
    }
    repos = strings.Split(repo, ",")
}

func main() {
    initFlags()

    switch taskType {
    case entityTask:
        if len(repos) != 1 {
            fmt.Fprintln(os.Stderr, "Entity task only support one repo name")
            os.Exit(-1)
        }
        option := tool.TimeLocationOption{
            StartTime:   startTimeMs,
            EndTime:     endTimeMs,
            LocationNum: LocNum,
        }
        task := tool.EntityTask{
            IPPort:        addr,
            RepoName:      repo,
            FeatureLength: featLen,
            Option:        option,
        }
        frame := tool.Frame{Task: task}
        frame.RunTask(qps, maxCount, threadNum)
    case searchTask:
        if len(repos) == 0 {
            fmt.Fprintln(os.Stderr, "Empty repos")
            os.Exit(-1)
        }
        task := tool.SearchTask{
            IPPort:        addr,
            MaxCandidates: topk,
        }
        task.Repositories = repos
        frame := tool.Frame{Task: task}
        frame.RunTask(qps, maxCount, threadNum)
    case compareTask:
        task := tool.CompareTask{IPPort: addr}
        frame := tool.Frame{Task: task}
        frame.RunTask(qps, maxCount, threadNum)
    }
}
