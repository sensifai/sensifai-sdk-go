package sensifai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

/*
SensifaiAPI : set token and base url to struct
*/
type SensifaiAPI struct {
	Token   string
	BaseURL *url.URL
}

/*
UploadResponse : bind graphql response to this struct
*/
type UploadResponse struct {
	Error   string `json:"error"`
	Result  bool   `json:"result"`
	Succeed []struct {
		File   string `json:"file"`
		TaskID string `json:"taskId"`
	} `json:"succeed"`
	CannotUpload []string `json:"cannotUpload"`
}

/*
ResultResponse : extract result from json
*/
type ResultResponse struct {
	Error  string `json:"errors"`
	IsDone bool   `json:"isDone"`
	*ImageResultsResponse
	*VideoResultsResponse
}

/*
ImageResultsResponse : extract result from json
*/
type ImageResultsResponse struct {
	ImageResults interface{} `json:"imageResults"`
}

/*
VideoResultsResponse : extract result from json
*/
type VideoResultsResponse struct {
	FPS          float32     `json:"fps"`
	Duration     float32     `json:"duration"`
	FramesCount  int         `json:"framesCount"`
	VideoResults interface{} `json:"videoResults"`
}

/*
CreateSensifaiAPI : for set your token and api url
*/
func CreateSensifaiAPI(Token string) *SensifaiAPI {
	apiString := "https://api.sensifai.com/api/"
	apiURL, err := url.Parse(apiString)
	if err != nil {
		return nil
	}
	return &SensifaiAPI{
		Token:   Token,
		BaseURL: apiURL,
	}
}

/*
UploadByFile : upload by file
*/
func (s *SensifaiAPI) UploadByFile(paths []string) (result UploadResponse, err error) {
	// initiate multipart
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// create array of null
	null := make([]*string, len(paths))

	// create payload
	variables := map[string]interface{}{
		"files": null,
		"token": s.Token,
	}
	operations := map[string]interface{}{
		"query":     "mutation( $token: String!, $files: [Upload!]! ){uploadByFile(token: $token, files: $files){result error succeed{file taskId} cannotUpload}}",
		"variables": variables,
	}
	operationsBytes, _ := json.Marshal(operations)
	payload := map[string]io.Reader{
		"operations": bytes.NewReader(operationsBytes),
	}
	fileMap := make(map[string][]string)

	// open files and append to map
	for index, path := range paths {
		indexString := strconv.Itoa(index)
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		payload[indexString] = file
		fileMap[indexString] = []string{"variables.files." + indexString}
	}

	// append map variable to payload
	bytesFileMap, _ := json.Marshal(fileMap)
	payload["map"] = bytes.NewReader(bytesFileMap)

	// create form file from payload
	for key, r := range payload {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			// Add file
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return result, err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return result, err
			}
		}
		if _, err := io.Copy(fw, r); err != nil {
			return result, err
		}
	}

	w.Close()

	// Create New POST Request
	req, err := http.NewRequest("POST", s.BaseURL.String(), &b)
	if err != nil {
		return result, err
	}

	// Set require heades
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Call Api and get response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Extract data from response
	body, _ := ioutil.ReadAll(resp.Body)
	responseMap := make(map[string]json.RawMessage)
	json.Unmarshal([]byte(body), &responseMap)
	json.Unmarshal([]byte(responseMap["data"]), &responseMap)
	json.Unmarshal([]byte(responseMap["uploadByFile"]), &result)
	if resp.Status != "200 OK" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

/*
UploadByURL : upload by url
*/
func (s *SensifaiAPI) UploadByURL(inputURLs []string) (result UploadResponse, err error) {
	// Create Payload
	variables := map[string]interface{}{
		"urls":  inputURLs,
		"token": s.Token,
	}
	payload := map[string]interface{}{
		"query":     "mutation( $token: String!, $urls: [String!]! ){uploadByUrl(token: $token, urls: $urls){result error succeed{file taskId} cannotUpload}}",
		"variables": variables,
	}
	payloadJSON, _ := json.Marshal(payload)

	// Create New POST Request
	req, err := http.NewRequest("POST", s.BaseURL.String(), bytes.NewBuffer(payloadJSON))
	if err != nil {
		return result, err
	}

	// Set require heades
	req.Header.Set("Content-Type", "application/json")

	// Call Api and get response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Extract data from response
	body, _ := ioutil.ReadAll(resp.Body)
	responseMap := make(map[string]json.RawMessage)
	json.Unmarshal([]byte(body), &responseMap)
	json.Unmarshal([]byte(responseMap["data"]), &responseMap)
	json.Unmarshal([]byte(responseMap["uploadByUrl"]), &result)
	if resp.Status != "200 OK" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

/*
GetResult : get result of a task
*/
func (s *SensifaiAPI) GetResult(taskID string) (result ResultResponse, err error) {
	variables := map[string]interface{}{
		"taskId": taskID,
	}
	payload := map[string]interface{}{
		"query":     "query( $taskId: String! ){apiResult( taskId: $taskId ){ ...on ImageResult{isDone errors imageResults{nsfwResult{type probability value}logoResult{description}landmarkResult{description}taggingResult{label probability}faceResult{detectedBoxesPercentage probability detectedFace label}}} ... on VideoResult{fps duration isDone framesCount errors videoResults{startSecond endSecond startFrame endFrame thumbnailPath taggingResult{label probability}actionResult{label probability}celebrityResult{name frequency} sportResult{label probability}nsfwResult{probability type value}}}}}",
		"variables": variables,
	}

	payloadJSON, _ := json.Marshal(payload)

	// Create New POST Request
	req, err := http.NewRequest("POST", s.BaseURL.String(), bytes.NewBuffer(payloadJSON))
	if err != nil {
		return result, err
	}

	// Set require heades
	req.Header.Set("Content-Type", "application/json")

	// Call Api and get response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Extract data from response
	body, _ := ioutil.ReadAll(resp.Body)
	responseMap := make(map[string]json.RawMessage)
	json.Unmarshal([]byte(body), &responseMap)
	json.Unmarshal([]byte(responseMap["data"]), &responseMap)
	json.Unmarshal([]byte(responseMap["apiResult"]), &result)
	if resp.Status != "200 OK" {
		return result, errors.New(result.Error)
	}
	return result, nil
}
