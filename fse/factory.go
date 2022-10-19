package fse

import (
	"flag"
	"os"
	"strings"

	"github.com/golang/glog"
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
	topK        int
	featLen     int
	LocNum      int
	repos       []string
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:30800", "FSE address, {IP}:{PORT}")
	flag.StringVar(&repo, "repo", "repo", "Repo names, split by comma, ie. 'repo1,repo2'")
	flag.StringVar(&taskType, "type", entityTask, "Task type: 'entityTask' for adding features, 'searchTask' for searching features, 'compareTask' for comparing task")
	flag.StringVar(&idType, "id", Uid, "ID type: 'uuid' for UUID, 'num' for number sequence, starting from 0")
	flag.Int64Var(&startTimeMs, "st", 0, "The start time in millisecond, only used in 'entityTask' type")
	flag.Int64Var(&endTimeMs, "et", 0, "The end time in millisecond, only used in 'entityTask' type")
	flag.IntVar(&topK, "topk", 3, "Top K")
	flag.Float64Var(&KnnThreshold, "knnThreshold", 0.5, "KNN threshold")
	flag.IntVar(&featLen, "len", 384, "The length of feature")
	flag.IntVar(&LocNum, "loc", 1, "The number of locations, only used in 'entityTask' type")

	flag.Parse()

	if taskType != entityTask && taskType != searchTask && taskType != compareTask {
		glog.Errorln("Unknown task type", taskType)
		os.Exit(-1)
	}
	if idType != Uid && idType != Num {
		glog.Errorln("Unknown ID type", idType)
		os.Exit(-1)
	}
	repos = strings.Split(repo, ",")
}

type TaskFactory struct {}

type entityTaskFactory struct {}

type searchTaskFactory struct {}

type compareTaskFactory struct {}

func (f *TaskFactory) CreateTask() task {
	var task task
	switch taskType {
	case entityTask:
		return (&entityTaskFactory{}).createTask()
	case searchTask:
		return (&searchTaskFactory{}).createTask()
	case compareTask:
		return (&compareTaskFactory{}).createTask()
	}
	return task
}

func (f *entityTaskFactory) createTask() task {
	if len(repos) != 1 {
		glog.Errorln("Entity task only support one repo name")
		os.Exit(-1)
	}
	option := TimeLocationOption{
		StartTime:   startTimeMs,
		EndTime:     endTimeMs,
		LocationNum: LocNum,
	}
	return EntityTask{
		IPPort:        addr,
		RepoName:      repo,
		FeatureLength: featLen,
		IdType:        idType,
		Option:        option,
	}
}

func (f *searchTaskFactory) createTask() task {
	if len(repos) == 0 {
		glog.Errorln("Empty repos")
		os.Exit(-1)
	}
	return SearchTask{
		IPPort:        addr,
		MaxCandidates: topK,
		FeatureLength: featLen,
		Repositories:  repos,
	}
}

func (f *compareTaskFactory) createTask() task {
	return CompareTask{
		IPPort:        addr,
		FeatureLength: featLen,
	}
}