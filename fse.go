package fse

import (
    "bytes"
    "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "io"
    "math"
    "math/rand"
    "net/http"
    "os"
    "strconv"
    "time"
)

var transport = &http.Transport{
    MaxIdleConnsPerHost: 500,
    MaxConnsPerHost: 1000,
}

var client = &http.Client{
    Transport: transport,
}

type EntityData struct {
    Type string `json:"type"`
    Version string `json:"version"`
    Value string `json:"value"`
}

type IncludeItem struct {
    Data EntityData `json:"data"`
}

type SearchBody struct {
    Type string              `json:"type"`
    Include []IncludeItem    `json:"include"`
    IncludeThreshold float32 `json:"include_threshold"`
    Repositories []string `json:"repositories"`
    MaxCandidates int `json:"max_candidates"`
}

type ObjectItem struct {
    ID string `json:"id"`
    LocationID string `json:"location_id"`
    Time int64      `json:"time"`
    Data EntityData `json:"data"`
}

type CompareObject struct {
    ID   string     `json:"id"`
    Data EntityData `json:"data"`
}

type CompareBody struct {
    Type string `json:"type"`
    Threshold float32        `json:"threshold"`
    MObjects []CompareObject `json:"m_objects"`
    NObjects []CompareObject `json:"n_objects"`
}

func generateDefaultEntityData(value *string) *EntityData {
    entityData := &EntityData{
        Type: "feature",
        Version: "1.8.0.1",
        Value: *value,
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
        Data: *data,
        ID: id,
        LocationID: location,
        Time: time,
    }
    return objItem
}

func generateDefaultSearchBody(includeItem *IncludeItem) *SearchBody {
    searchBody := &SearchBody{
        Type: "face",
        Include: []IncludeItem{*includeItem},
        IncludeThreshold: 0,
        MaxCandidates: 3,
    }
    return searchBody
}

func GenerateRandomFeature(featureLength int) *[]float32 {
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
    return &feature
}

func generateCompareObject(id string, data *EntityData) *CompareObject {
    returnObjectItem := &CompareObject{
        ID: id,
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

func EncodeFeature(feature *[]float32) *string {
    buffer := new(bytes.Buffer)
    err := binary.Write(buffer, binary.LittleEndian, feature)
    if err != nil {
        fmt.Fprintf(os.Stderr, "binary.Write failed: %s\n", err.Error())
    }

    featureBytes := make([]byte, len(*feature) * 4)
    if n, err := buffer.Read(featureBytes); err != nil {
        fmt.Fprintf(os.Stderr, "buffer.Read failed: %s\n", err.Error())
        fmt.Fprintf(os.Stderr, "buffer.Read number: %d\n", n)
    }

    encodedString := base64.StdEncoding.EncodeToString(featureBytes)
    return &encodedString
}

func postAndCheck(requestBytes []byte, url string) int {
    resp, err := client.Post(url, "application/json", bytes.NewReader(requestBytes))
    if err != nil {
        fmt.Fprintf(os.Stderr, "client.Post failed: %s\n", err.Error())
        return -1
    }

    defer resp.Body.Close()

    responseBodyBytes := make([]byte, 1024 * 100)
    n, err := resp.Body.Read(responseBodyBytes)

    if resp.StatusCode / 100 != 2 {
        fmt.Fprintf(os.Stderr, "Status code is %d: %s\n", resp.StatusCode, string(responseBodyBytes))
        return -1
    }

    if err != nil && err != io.EOF {
        fmt.Fprintf(os.Stderr, "resp.Body.Read failed: %s\n", err.Error())
        return -1
    }

    if n <= 0 {
        fmt.Fprintf(os.Stderr, "Empty resp.Body\n")
        return -1
    }
    return 0
}

type Task interface {
    run(int64, int64) int
}

type SearchTask struct {
    IPPort string
    Repositories []string
    MaxCandidates int
    urlPrefix     string
}

func (t SearchTask) run(int64, int64) int {
    t.urlPrefix = "http://" + t.IPPort + "/x-api/v1/repositories/"
    feature := GenerateRandomFeature(384)
    encodedString := EncodeFeature(feature)

    entityData := generateDefaultEntityData(encodedString)
    includeItem := generateDefaultIncludeItem(entityData)
    searchBody := generateDefaultSearchBody(includeItem)
    searchBody.MaxCandidates = t.MaxCandidates

    for _, repo := range t.Repositories {
        searchBody.Repositories = append(searchBody.Repositories, repo)
    }

    numOfBytes, err := json.Marshal(*searchBody)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to marshal json: %s\n", err.Error())
    }

    return postAndCheck(numOfBytes, t.urlPrefix+ "search")
}

type CompareTask struct {
    IPPort    string
    urlPrefix string
}

func (t CompareTask) run(int64, int64) int {
    t.urlPrefix = "http://" + t.IPPort + "/x-api/v1/repositories/"
    feature1 := GenerateRandomFeature(384)
    encodedString1 := EncodeFeature(feature1)
    feature2 := GenerateRandomFeature(384)
    encodedString2 := EncodeFeature(feature2)

    entityData1 := generateDefaultEntityData(encodedString1)
    entityData2 := generateDefaultEntityData(encodedString2)

    mObjects := make([]CompareObject, 0, 1)
    nObjects := make([]CompareObject, 0, 1)
    mObjects = append(mObjects, *generateCompareObject("m-id-1", entityData1))
    nObjects = append(nObjects, *generateCompareObject("n-id-1", entityData2))
    compareBody := generateCompareBody(mObjects, nObjects)

    numOfBytes, err := json.Marshal(*compareBody)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to marshal json: %s\n", err.Error())
    }

    return postAndCheck(numOfBytes, t.urlPrefix+ "compare")
}

type TimeLocationOption struct {
    StartTime int64
    EndTime int64
    LocationNum int
}

type EntityTask struct {
    IPPort string
    RepoName string
    FeatureLength int
    Option    TimeLocationOption
    urlPrefix string
}

func setEntityTimeLocation(item *ObjectItem, option *TimeLocationOption, num, totalFeatureNum int64) {
    item.LocationID = strconv.Itoa(int(num) % option.LocationNum)

    start := time.Duration(option.StartTime) * time.Millisecond
    end := time.Duration(option.EndTime) * time.Millisecond
    timeRange := end - start
    if timeRange <= 0 {
        item.Time = 0
        return
    }

    timeStep := timeRange / time.Duration(totalFeatureNum) / time.Millisecond
    item.Time = num * int64(timeStep) + int64(timeStep) / 2
}

func (t EntityTask) run(num, totalFeatureNum int64) int {
    t.urlPrefix = "http://" + t.IPPort + "/x-api/v1/repositories/"
    feature := GenerateRandomFeature(t.FeatureLength)
    encodedString := EncodeFeature(feature)
    entityData := generateDefaultEntityData(encodedString)
    item := generateEntityItem(entityData, uuid.New().String(), "0", 0)
    setEntityTimeLocation(item, &t.Option, num, totalFeatureNum)
    numOfBytes, err := json.Marshal(*item)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to marshal json: %s\n", err.Error())
        return -1
    }

    return postAndCheck(numOfBytes, t.urlPrefix+ t.RepoName + "/entities")
}

type Frame struct {
    Task    Task
    startCh chan int64
    endCh      chan int
    latencyCh chan float64
    failureCh        chan int
    reportCh         <-chan time.Time
    endStatisticCh    chan int
    dropRequestsCh  chan int
    totalFeatureNum int64
}

func (frame *Frame)threadWrapper() {
    for {
        select {
        case id := <-frame.startCh:
            startTime := time.Now()
            if frame.Task.run(id, frame.totalFeatureNum) == 0 {
                frame.latencyCh <- time.Now().Sub(startTime).Seconds() * 1000
            } else {
                frame.failureCh <- 1
            }
        case <-frame.endCh:
            return
        }
    }
}

func (frame *Frame)getStatistics() {
    var successCount, failureCount, dropCount = 0, 0, 0
    var latency, maxLatency, minLatency, averageLatency float64 = 0, 0, 9999999999999999, 0
    currentTime := time.Now()
    printStatistics := func () {
            averageLatency /= float64(successCount)
            elapsedSec := time.Now().Sub(currentTime).Seconds()
            fmt.Printf("Last %.2f seconds: qps %.2f, avg_latency %.2fms, min_latency %.2fms, max_latency %.2fms, failure %d, drop %d\n",
                elapsedSec,
                        float64(successCount) /elapsedSec,
                averageLatency,
                minLatency,
                maxLatency,
                failureCount,
                dropCount)
            successCount, failureCount, dropCount = 0, 0, 0
            latency, maxLatency, minLatency, averageLatency = 0, 0, 9999999999999999, 0
            currentTime = time.Now()
    }
    for {
        select {
        case latency = <-frame.latencyCh:
            if maxLatency < latency {
                maxLatency = latency
            }
            if minLatency > latency {
                minLatency = latency
            }
            averageLatency += latency
            successCount += 1
        case <-frame.failureCh:
            failureCount += 1
        case <-frame.dropRequestsCh:
            dropCount += 1
        case <-frame.reportCh:
            printStatistics()
        case <-frame.endStatisticCh:
            printStatistics()
            return
        }
    }
}

func (frame *Frame)RunTask(qps, maxCount int64, threadNum int) {
    frame.startCh = make(chan int64)
    frame.endCh = make(chan int)
    frame.failureCh = make(chan int)
    frame.latencyCh = make(chan float64, threadNum)
    frame.endStatisticCh = make(chan int)
    frame.reportCh = time.NewTicker(time.Second * 10).C
    frame.dropRequestsCh = make(chan int, qps / 10)
    frame.totalFeatureNum = maxCount

    go frame.getStatistics()
    timeInterval := time.Second / time.Duration(qps)
    for i := 0; i < threadNum; i++ {
        go frame.threadWrapper()
    }

    var sum int64 = 0
    stopThread := false
    ticker := time.NewTicker(timeInterval)
    for {
        select {
        case <-ticker.C:
            select {
            case frame.startCh <- sum:
                sum += 1
                if sum >= maxCount {
                    fmt.Printf("sum: %d, break now\n", sum)
                    stopThread = true
                } else if sum % 10000 == 0 {
                    fmt.Printf("send %d requests\n", sum)
                }
            default:
                frame.dropRequestsCh <- 1
            }
        }
        if stopThread {
            break
        }
    }

    for i := 0; i < threadNum; i++ {
        frame.endCh <- 1
    }
    frame.endStatisticCh <- 1
    fmt.Println("All threads end")
}
