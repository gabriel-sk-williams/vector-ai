package model

type ContextHolder struct {
	Results []ContextLoaderScored `json:"results"`
	Query   string                `json:"query"`
}

func (cv *ContextHolder) AddLoader(val ContextLoaderScored) []ContextLoaderScored {
	cv.Results = append(cv.Results, val)
	return cv.Results
}

type ContextLoaderScored struct {
	DocumentId   string       `json:"documentId"`
	DocumentName string       `json:"documentName"`
	Context      *ChunkValues `json:"context"`
}

type ChunkValues struct {
	Values []ScoredChunk `json:"values"`
}

func (cv *ChunkValues) AddChunk(val ScoredChunk) []ScoredChunk {
	cv.Values = append(cv.Values, val)
	return cv.Values
}

type ScoredChunk struct {
	ID    string  `json:"id"`
	Value string  `json:"value"`
	Score float32 `json:"score"`
}
