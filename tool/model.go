package tool

import "time"

type EntityData struct {
    Type    string `json:"type"`
    Version string `json:"version"`
    Value   string `json:"value"`
}

type IncludeItem struct {
    Data EntityData `json:"data"`
}

type SearchBody struct {
    Type             string        `json:"type"`
    Include          []IncludeItem `json:"include"`
    IncludeThreshold float32       `json:"include_threshold"`
    Repositories     []string      `json:"repositories"`
    MaxCandidates    int           `json:"max_candidates"`
}

type ObjectItem struct {
    ID         string     `json:"id"`
    LocationID string     `json:"location_id"`
    Time       int64      `json:"time"`
    Data       EntityData `json:"data"`
}

type CompareObject struct {
    ID   string     `json:"id"`
    Data EntityData `json:"data"`
}

type CompareBody struct {
    Type      string          `json:"type"`
    Threshold float32         `json:"threshold"`
    MObjects  []CompareObject `json:"m_objects"`
    NObjects  []CompareObject `json:"n_objects"`
}

type SearchTask struct {
    IPPort        string
    Repositories  []string
    MaxCandidates int
    FeatureLength int
    url           string
}

type CompareTask struct {
    IPPort        string
    FeatureLength int
    url           string
}

type TimeLocationOption struct {
    StartTime   int64
    EndTime     int64
    LocationNum int
}

type EntityTask struct {
    IPPort        string
    RepoName      string
    IdType        string
    FeatureLength int
    Option        TimeLocationOption
    url           string
}

type Frame struct {
    Task            task
    startCh         chan int64
    endCh           chan struct{}
    resultCh        chan *result
    reportCh        <-chan time.Time
    endStatisticCh  chan struct{}
    dropRequestsCh  chan struct{}
    totalFeatureNum int64
}
