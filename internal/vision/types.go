package vision

type RecognitionRequest struct {
	DeviceID      string `json:"device_id"`
	FrameID       string `json:"frame_id"`
	ImageBase64   string `json:"image_base64"`
	LastSceneHash string `json:"last_scene_hash,omitempty"`
	OfflineOK     bool   `json:"offline_ok,omitempty"`
}

type RecognitionResult struct {
	SceneHash string   `json:"scene_hash"`
	FromCache bool     `json:"from_cache"`
	Objects   []Object `json:"objects"`
}

type Object struct {
	Letter   string  `json:"letter"`
	Name     string  `json:"name"`
	Meaning  string  `json:"meaning"`
	Phonetic string  `json:"phonetic"`
	Sentence string  `json:"sentence"`
	Box      Box     `json:"box"`
	Score    float64 `json:"score"`
}

type Box struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Recognizer interface {
	Recognize(req RecognitionRequest) (RecognitionResult, error)
}
