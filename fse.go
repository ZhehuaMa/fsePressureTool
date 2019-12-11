package fse

import (
    "os"
    "fmt"
    "encoding/json"
    "encoding/binary"
    "encoding/base64"
    "math/rand"
    "math"
    "time"
    "bytes"
    "net/http"
    "io"
    "github.com/google/uuid"
)

var transport = &http.Transport{
    MaxIdleConnsPerHost: 500,
    MaxConnsPerHost: 1000,
}

var client = &http.Client{
    Transport: transport,
}

type entityData struct {
    Type string `json:"type"`
    Version string `json:"version"`
    Value string `json:"value"`
}

type includeItem struct {
    Data entityData `json:"data"`
}

type searchBody struct {
    Type string `json:"type"`
    Include []includeItem `json:"include"`
    IncludeThreshold float32 `json:"include_threshold"`
    Repositories []string `json:"repositories"`
    MaxCandidates int `json:"max_candidates"`
}

type objectItem struct {
    ID string `json:"id"`
    LocationID string `json:"location_id"`
    Time int64 `json:"time"`
    Data entityData `json:"data"`
}

type compareBody struct {
    Type string `json:"type"`
    Threshold float32 `json:"threshold"`
    MObjects []objectItem `json:"m_objects"`
    NObjects []objectItem `json:"n_objects"`
}

func generateDefaultEntityData(value *string) *entityData {
    entity_data := &entityData{
        Type: "feature",
        Version: "1.8.0.1",
        Value: *value,
    }
    return entity_data
}

func generateDefaultIncludeItem(data *entityData) *includeItem {
    include_item := &includeItem{
        Data: *data,
    }
    return include_item
}

func generateEntityItem(data *entityData, id, location string, time int64) *objectItem {
    obj_item := &objectItem{
        Data: *data,
        ID: id,
        LocationID: location,
        Time: time,
    }
    return obj_item
}

func generateDefaultSearchBody(include_item *includeItem) *searchBody {
    search_body := &searchBody{
        Type: "face",
        Include: []includeItem{*include_item},
        IncludeThreshold: 0,
        MaxCandidates: 3,
    }
    return search_body
}

func GenerateRandomFeature(feature_length int) *[]float32 {
    feature := make([]float32, feature_length)
    var sum float32 = 0
    rand.Seed(time.Now().UnixNano())
    for i := range feature {
        feature[i] = rand.Float32()
        feature[i] = 0.01
        sum += feature[i] * feature[i]
    }
    sum = float32(math.Sqrt(float64(sum)))
    for i := range feature {
        feature[i] /= sum
    }
    return &feature
}

func generateObjectItem(id string, data *entityData) *objectItem {
    return_object_item := &objectItem{
        ID: id,
        Data: *data,
    }
    return return_object_item
}

func generateCompareBody(m_objects, n_objects []objectItem) *compareBody {
    ret_compare_body := &compareBody{
        Type: "face",
        Threshold: 0,
        MObjects: m_objects,
        NObjects: n_objects,
    }
    return ret_compare_body
}

func EncodeFeature(feature *[]float32) *string {
    buffer := new(bytes.Buffer)
    err := binary.Write(buffer, binary.LittleEndian, feature)
    if err != nil {
        fmt.Fprintf(os.Stderr, "binary.Write failed: %d\n", err.Error())
    }

    feature_bytes := make([]byte, len(*feature) * 4)
    if n, err := buffer.Read(feature_bytes); err != nil {
        fmt.Fprintf(os.Stderr, "buffer.Read failed: %s\n", err.Error())
        fmt.Fprintf(os.Stderr, "buffer.Read number: %d\n", n)
    }

    encoded_string := base64.StdEncoding.EncodeToString(feature_bytes)
    return &encoded_string
}

func postAndCheck(request_bytes []byte, url string) int {
    resp, err := client.Post(url, "application/json", bytes.NewReader(request_bytes))
    if err != nil {
        fmt.Fprintf(os.Stderr, "client.Post failed: %s\n", err.Error())
        return -1
    }

    defer resp.Body.Close()

    response_body_bytes := make([]byte, 1024 * 100)
    n, err := resp.Body.Read(response_body_bytes)

    if resp.StatusCode / 100 != 2 {
        fmt.Fprintf(os.Stderr, "Status code is %d: %s\n", resp.StatusCode, string(response_body_bytes))
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

type FSETask interface {
    Run() int
}

type SearchTask struct {
    UrlPrefix string
    Repositories []string
    MaxCandidates int
}

func (t SearchTask) Run() int {
    feature := GenerateRandomFeature(384)
    encoded_string := EncodeFeature(feature)

    entity_data := generateDefaultEntityData(encoded_string)
    include_item := generateDefaultIncludeItem(entity_data)
    search_body := generateDefaultSearchBody(include_item)
    search_body.MaxCandidates = t.MaxCandidates

    for _, repo := range t.Repositories {
        search_body.Repositories = append(search_body.Repositories, repo)
    }

    bytes, err := json.Marshal(*search_body)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to marshal json: %s\n", err.Error())
    }

    return postAndCheck(bytes, t.UrlPrefix + "search")
}

type CompareTask struct {
    UrlPrefix string
}

func (t CompareTask) Run() int {
    feature1 := GenerateRandomFeature(384)
    encoded_string1 := EncodeFeature(feature1)
    feature2 := GenerateRandomFeature(384)
    encoded_string2 := EncodeFeature(feature2)

    entity_data1 := generateDefaultEntityData(encoded_string1)
    entity_data2 := generateDefaultEntityData(encoded_string2)

    m_objects := make([]objectItem, 0, 1)
    n_objects := make([]objectItem, 0, 1)
    m_objects = append(m_objects, *generateObjectItem("m-di-1", entity_data1))
    n_objects = append(n_objects, *generateObjectItem("n-di-1", entity_data2))
    compare_body := generateCompareBody(m_objects, n_objects)

    bytes, err := json.Marshal(*compare_body)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to marshal json: %s\n", err.Error())
    }

    return postAndCheck(bytes, t.UrlPrefix + "compare")
}

type EntityTask struct {
    UrlPrefix string
    RepoName string
    FeatureLength int
}

func (t EntityTask) Run() int {
    feature := GenerateRandomFeature(t.FeatureLength)
    encoded_string := EncodeFeature(feature)
    entity_data := generateDefaultEntityData(encoded_string)
    item := generateEntityItem(entity_data, uuid.New().String(), "0", 0)
    bytes, err := json.Marshal(*item)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fail to marshal json: %s\n", err.Error())
        return -1
    }

    return postAndCheck(bytes, t.UrlPrefix + t.RepoName + "/entities")
}

type FSEFrame struct {
    Task FSETask
    start_ch chan int
    end_ch chan int
    latency_ch chan float64
    failure_ch chan int
    report_ch <-chan time.Time
    end_statistic_ch chan int
    drop_requests_ch chan int
}

func (frame *FSEFrame)threadWrapper() {
    for {
        select {
        case <-frame.start_ch:
            start_time := time.Now()
            if frame.Task.Run() == 0 {
                frame.latency_ch <- time.Now().Sub(start_time).Seconds() * 1000
            } else {
                frame.failure_ch <- 1
            }
        case <-frame.end_ch:
            return
        }
    }
}

func (frame *FSEFrame)getStatistics() {
    var success_count, failure_count, drop_count int = 0, 0, 0
    var latency, max_latency, min_latency, average_latency float64 = 0, 0, 9999999999999999, 0
    current_time := time.Now()
    print_statistics := func () {
            average_latency /= float64(success_count)
            elapsed_sec := time.Now().Sub(current_time).Seconds()
            fmt.Printf("Last %.2f seconds: qps %.2f, avg_latency %.2fms, min_latency %.2fms, max_latency %.2fms, failure %d, drop %d\n",
                        elapsed_sec,
                        float64(success_count) / elapsed_sec,
                        average_latency,
                        min_latency,
                        max_latency,
                        failure_count,
                        drop_count)
            success_count, failure_count, drop_count = 0, 0, 0
            latency, max_latency, min_latency, average_latency = 0, 0, 9999999999999999, 0
            current_time = time.Now()
    }
    for {
        select {
        case latency = <-frame.latency_ch:
            if max_latency < latency {
                max_latency = latency
            }
            if min_latency > latency {
                min_latency = latency
            }
            average_latency += latency
            success_count += 1
        case <-frame.failure_ch:
            failure_count += 1
        case <-frame.drop_requests_ch:
            drop_count += 1
        case <-frame.report_ch:
            print_statistics()
        case <-frame.end_statistic_ch:
            print_statistics()
            return
        }
    }
}

func (frame *FSEFrame)RunTask(qps, max_count int64, thread_num int) {
    frame.start_ch = make(chan int)
    frame.end_ch = make(chan int)
    frame.failure_ch = make(chan int)
    frame.latency_ch = make(chan float64, thread_num)
    frame.end_statistic_ch = make(chan int)
    frame.report_ch = time.NewTicker(time.Second * 10).C
    frame.drop_requests_ch = make(chan int, qps / 10)

    go frame.getStatistics()
    time_interval := time.Second / time.Duration(qps)
    for i := 0; i < thread_num; i++ {
        go frame.threadWrapper()
    }

    var sum int64 = 0
    stop_thread := false
    ticker := time.NewTicker(time_interval)
    for {
        select {
        case <-ticker.C:
            select {
            case frame.start_ch <- 1:
                sum += 1
                if sum >= max_count {
                    fmt.Printf("sum: %d, break now\n", sum)
                    stop_thread = true
                } else if sum % 10000 == 0 {
                    fmt.Printf("send %d requests\n", sum)
                }
            default:
                frame.drop_requests_ch <- 1
            }
        }
        if stop_thread {
            break
        }
    }

    for i := 0; i < thread_num; i++ {
        frame.end_ch <- 1
    }
    frame.end_statistic_ch <- 1
    fmt.Println("All threads end")
}
