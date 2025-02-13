package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kyma-project/runtime-watcher/listener/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const contentMapCapacity = 3

type UnmarshalError struct {
	Message       string
	HTTPErrorCode int
}

func UnmarshalSKREvent(req *http.Request) (*unstructured.Unstructured, *UnmarshalError) {
	pathVariables := strings.Split(req.URL.Path, "/")

	var contractVersion string
	_, err := fmt.Sscanf(pathVariables[1], "v%s", &contractVersion)

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, &UnmarshalError{"could not read contract version", http.StatusBadRequest}
	}

	if err != nil && errors.Is(err, io.EOF) || contractVersion == "" {
		return nil, &UnmarshalError{"contract version cannot be empty", http.StatusBadRequest}
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, &UnmarshalError{"could not read request body", http.StatusInternalServerError}
	}
	defer req.Body.Close()

	watcherEvent := &types.WatchEvent{}
	err = json.Unmarshal(body, watcherEvent)
	if err != nil {
		return nil, &UnmarshalError{"could not unmarshal watcher event", http.StatusInternalServerError}
	}

	genericEvtObject := &unstructured.Unstructured{}
	content := UnstructuredContent(watcherEvent)
	genericEvtObject.SetUnstructuredContent(content)
	genericEvtObject.SetName(watcherEvent.Owner.Name)
	genericEvtObject.SetNamespace(watcherEvent.Owner.Namespace)

	return genericEvtObject, nil
}

func UnstructuredContent(watcherEvt *types.WatchEvent) map[string]interface{} {
	content := make(map[string]interface{}, contentMapCapacity)
	content["owner"] = watcherEvt.Owner
	content["watched"] = watcherEvt.Watched
	content["watched-gvk"] = watcherEvt.WatchedGvk
	return content
}
