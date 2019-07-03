Sensifai API Golang Client
====================


## Overview
This Golang client provides a wrapper around Sensifai [Image and Video recognition API](https://developer.sensifai.com).

## Installation
The API client is available on Github.
Using this API is pretty simple. First, you need to install it using `go get` like this:

```bash
go get github.com/sensifai/sensifai-sdk-go
```

Then, you can import it and use it as follows:

```go
import (
	"github.com/sensifai/sensifai-sdk-go"
)

func main() {
	// api is a struct of SensifaiAPI type
	api := sensifai.CreateSensifaiAPI("YOUR_APPLICATION_TOKEN")
}
```


### Sample Usage
The following example will set up the client and predict video or image attributes.
First of all, you need to import the library and define an instance as mentioned above.
You can get a free limited `token` from [Developer Panel](https://developer.sensifai.com) by creating an application.
Then, if you want to process Data by URL you can call `UploadByURL` like the below sample code.

```go
urlsList := []string{"https://url1.png", "http://url2.jpg"}
// result is a struct of UploadResponse type
result, err := api.UploadByURL( urlsList )
```

Also, if you want to process Data by File, you can call `UploadByFile` like the following sample code. 

```go
fileList := []string{"/home/test.jpg","/home/file.jpg"}
// result is a struct of UploadResponse type
result, err := api.UploadByFile( fileList )
```

In the end, to retrieve the result of a task, pass its taskID through `GetResult`.
Please don't forget to pass a single `TaskID`! this function won't work with a list of taskIDs.

```go
TaskID := 'XXXX-XXX-XXXX-XXXX'
result, err := api.GetResult( TaskID )
```

### Full Code

```go
import (
	"fmt"

	"github.com/sensifai/sensifai-sdk-go"
)

func main() {
	api := sensifai.CreateSensifaiAPI("YOUR_APPLICATION_TOKEN")
	paths := []string{
		"/home/test.jpg",
		"/home/test.png",
	}
	// links := []string{
	// 	"https://test.jpg",
	// 	"https://test.png",
	// }
	result, err := api.UploadByFile(paths)
	// result, err := api.UploadByURL(links)
	if err != nil {
		return panic(err)
	}
	fmt.Println(result.Succeed, result.Error, result.Result)
	// get result of first success link or file
	finalResult, err := api.GetResult(result.Succeed[0].TaskID)
	if err != nil {
		return panic(err)
	}
	// if your input is video, you can get ImageResultsResponse from
	// finalResult and get VideoResultsResponse if your file is video.
	// both of these are from json.RawMessage
	fmt.Println(finalResult.IsDone, finalResult.Error)
}
```
