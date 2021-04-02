package main

import (
	"flag"
	"fsePressureTool/fse"
	"github.com/golang/glog"
	"os"
	"strings"
)

const (
	entityTask  = "entityTask"
	searchTask  = "searchTask"
	compareTask = "compareTask"
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
	repos       []string
)

func initFlag() {
	flag.StringVar(&addr, "addr", "127.0.0.1:30800", "FSE address, {IP}:{PORT}")
	flag.StringVar(&repo, "repo", "repo", "Repo names, split by comma, ie. 'repo1,repo2'")
	flag.StringVar(&taskType, "type", entityTask,
		"Task type: 'entityTask' for adding features, 'searchTask' for searching features, 'compareTask' for comparing task")
	flag.StringVar(&idType, "id", fse.Uid, "ID type: 'uuid' for UUID, 'num' for number sequence, starting from 0")
	flag.Int64Var(&startTimeMs, "st", 0, "The start time in millisecond, only used in 'entityTask' type")
	flag.Int64Var(&endTimeMs, "et", 0, "The end time in millisecond, only used in 'entityTask' type")
	flag.IntVar(&topk, "topk", 3, "Top K")
	flag.IntVar(&featLen, "len", 384, "The length of feature")
	flag.IntVar(&LocNum, "loc", 1, "The number of locations, only used in 'entityTask' type")
	flag.Int64Var(&maxCount, "max", 10, "Maximum times the task is executed")

	flag.Parse()

	if taskType != entityTask && taskType != searchTask && taskType != compareTask {
		glog.Errorln("Unknown task type", taskType)
		os.Exit(-1)
	}
	if idType != fse.Uid && idType != fse.Num {
		glog.Errorln("Unknown ID type", idType)
		os.Exit(-1)
	}
	repos = strings.Split(repo, ",")
}

func main() {
	initFlag()

	switch taskType {
	case entityTask:
		if len(repos) != 1 {
			glog.Errorln("Entity task only support one repo name")
			os.Exit(-1)
		}
		option := fse.TimeLocationOption{
			StartTime:   startTimeMs,
			EndTime:     endTimeMs,
			LocationNum: LocNum,
		}
		task := fse.EntityTask{
			IPPort:        addr,
			RepoName:      repo,
			FeatureLength: featLen,
			IdType:        idType,
			Option:        option,
		}
		frame := fse.Frame{Task: task}
		frame.RunTask(maxCount)
	case searchTask:
		if len(repos) == 0 {
			glog.Errorln("Empty repos")
			os.Exit(-1)
		}
		task := fse.SearchTask{
			IPPort:        addr,
			MaxCandidates: topk,
			FeatureLength: featLen,
		}
		task.Repositories = repos
		frame := fse.Frame{Task: task}
		frame.RunTask(maxCount)
	case compareTask:
		task := fse.CompareTask{
			IPPort:        addr,
			FeatureLength: featLen,
		}
		frame := fse.Frame{Task: task}
		frame.RunTask(maxCount)
	}
}
