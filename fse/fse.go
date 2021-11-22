package fse

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/zhehuama/Ezio"
)

const (
	httpStr   = "http://"
	prefixUrl = "/x-api/v1/repositories/"
)

const (
	Uid = "uuid"
	Num = "num"
)

var KnnThreshold float64

func generateDefaultEntityData(value *string) *EntityData {
	entityData := &EntityData{
		Type:    "feature",
		Version: "2.7.3.0",
		Value:   *value,
	}
	return entityData
}

func generateDefaultIncludeItem(data *EntityData) *IncludeItem {
	includeItem := &IncludeItem{
		Data: *data,
	}
	return includeItem
}

func generateEntityItem(data *EntityData, id, location string, time int64) *ObjectItem {
	objItem := &ObjectItem{
		Data:       *data,
		ID:         id,
		LocationID: location,
		Time:       time,
	}
	return objItem
}

func generateDefaultSearchBody(includeItem *IncludeItem) *SearchBody {
	searchBody := &SearchBody{
		Type:             "face",
		Include:          []IncludeItem{*includeItem},
		IncludeThreshold: 0,
		MaxCandidates:    3,
		Options:          make(map[string]string),
	}
	return searchBody
}

func GenerateRandomFeature(featureLength int) []float32 {
	feature := make([]float32, featureLength)
	var sum float32 = 0
	rand.Seed(time.Now().UnixNano())
	for i := range feature {
		feature[i] = rand.Float32()
		sum += feature[i] * feature[i]
	}
	sum = float32(math.Sqrt(float64(sum)))
	for i := range feature {
		feature[i] /= sum
	}
	return feature
}

func generateCompareObject(id string, data *EntityData) *CompareObject {
	returnObjectItem := &CompareObject{
		ID:   id,
		Data: *data,
	}
	return returnObjectItem
}

func generateCompareBody(mObjects, nObjects []CompareObject) *CompareBody {
	retCompareBody := &CompareBody{
		Type:      "face",
		Threshold: 0,
		MObjects:  mObjects,
		NObjects:  nObjects,
	}
	return retCompareBody
}

func EncodeFeature(feature []float32) *string {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, feature)
	if err != nil {
		glog.Errorf("binary.Write failed: %s\n", err.Error())
	}

	featureBytes := make([]byte, len(feature)*4)
	if n, err := buffer.Read(featureBytes); err != nil {
		glog.Errorf("buffer.Read failed: %s\n", err.Error())
		glog.Errorf("buffer.Read number: %d\n", n)
	}

	encodedString := base64.StdEncoding.EncodeToString(featureBytes)
	return &encodedString
}

type task interface {
	run(int64)
}

func (t SearchTask) run(featureNum int64) {
	t.url = httpStr + t.IPPort + prefixUrl + "search"
	var i int64 = 0
	for ; i < featureNum; i++ {
		feature := GenerateRandomFeature(t.FeatureLength)
		encodedString := EncodeFeature(feature)

		entityData := generateDefaultEntityData(encodedString)
		includeItem := generateDefaultIncludeItem(entityData)
		searchBody := generateDefaultSearchBody(includeItem)
		searchBody.MaxCandidates = t.MaxCandidates
		searchBody.Options["knn_threshold"] = strconv.FormatFloat(KnnThreshold, 'f', 2, 32)

		searchBody.Repositories = append(searchBody.Repositories, t.Repositories...)

		body, err := json.Marshal(*searchBody)
		if err != nil {
			glog.Errorf("Fail to marshal json: %s\n", err.Error())
			continue
		}

		request, _ := http.NewRequest("POST", t.url, bytes.NewReader(body))
		httpTask := &Ezio.Task{Request: request}
		Ezio.Append(httpTask)
	}
}

func (t CompareTask) run(featureNum int64) {
	t.url = httpStr + t.IPPort + prefixUrl + "compare"
	var i int64 = 0
	for ; i < featureNum; i++ {
		feature1 := GenerateRandomFeature(t.FeatureLength)
		encodedString1 := EncodeFeature(feature1)
		feature2 := GenerateRandomFeature(t.FeatureLength)
		encodedString2 := EncodeFeature(feature2)

		entityData1 := generateDefaultEntityData(encodedString1)
		entityData2 := generateDefaultEntityData(encodedString2)

		mObjects := make([]CompareObject, 0, 1)
		nObjects := make([]CompareObject, 0, 1)
		mObjects = append(mObjects, *generateCompareObject("m-id-1", entityData1))
		nObjects = append(nObjects, *generateCompareObject("n-id-1", entityData2))
		compareBody := generateCompareBody(mObjects, nObjects)

		body, err := json.Marshal(*compareBody)
		if err != nil {
			glog.Errorf("Fail to marshal json: %s\n", err.Error())
			continue
		}

		request, _ := http.NewRequest("POST", t.url, bytes.NewReader(body))
		httpTask := &Ezio.Task{Request: request}
		Ezio.Append(httpTask)
	}
}

func setEntityTimeLocation(item *ObjectItem, option *TimeLocationOption, num, totalFeatureNum int64) {
	item.LocationID = strconv.Itoa(int(num % int64(option.LocationNum)))

	start := time.Duration(option.StartTime) * time.Millisecond
	end := time.Duration(option.EndTime) * time.Millisecond
	timeRange := end - start
	if timeRange <= 0 {
		item.Time = 0
		return
	}

	timeStep := timeRange / time.Duration(totalFeatureNum) / time.Millisecond
	item.Time = num*int64(timeStep) + int64(timeStep)/2 + option.StartTime
}

func (t EntityTask) run(featureNum int64) {
	t.url = httpStr + t.IPPort + prefixUrl + t.RepoName + "/entities"
	var i int64 = 0
	for ; i < featureNum; i++ {
		feature := GenerateRandomFeature(t.FeatureLength)
		encodedString := EncodeFeature(feature)
		entityData := generateDefaultEntityData(encodedString)
		var id string
		if t.IdType == Uid {
			id = uuid.New().String()
		} else {
			id = strconv.Itoa(int(i))
		}
		item := generateEntityItem(entityData, id, "0", 0)
		setEntityTimeLocation(item, &t.Option, i, featureNum)
		body, err := json.Marshal(*item)
		if err != nil {
			glog.Errorf("Fail to marshal json: %s\n", err.Error())
			continue
		}

		request, _ := http.NewRequest("POST", t.url, bytes.NewReader(body))
		httpTask := &Ezio.Task{Request: request}
		Ezio.Append(httpTask)
	}
}

func (frame *Frame) RunTask(maxCount int64) {
	Ezio.Run()
	frame.Task.run(maxCount)
	Ezio.WaitUntilFinish()
}
