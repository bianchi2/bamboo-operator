package rest

import (
	"encoding/json"
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	"io/ioutil"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
	"strings"
)

const (
	BambooApiUrl = "https://bamboo.kubedemo.ml/rest/api/latest"
)

var (
	setupLog = ctrl.Log.WithValues()
)

func GetOnlineAgents(path string, base64Creds string, idleOnly bool) (err error, number int64) {
	var result []map[string]interface{}
	client := &http.Client{}
	req, err := http.NewRequest("GET", BambooApiUrl + path, nil)
	if err != nil {
		return err, -1
	}
	req.Header.Add("Authorization", "Basic " + base64Creds)
	resp, err := client.Do(req)
	if err != nil {
		return err, -1
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &result)
		if err != nil {
			fmt.Println(err)
		}
		if idleOnly {
			for i := range result {
				agentState := result[i]["busy"]
				if agentState == false {
					number = number + 1
				}
			}
			return nil, number
		}
		number := len(result)
		defer resp.Body.Close()

		return nil, int64(number)
	}
}


func DeleteAgentsById(path string, agentIds []string, base64Creds string) (err error) {
	client := &http.Client{}
	for i := range agentIds {
		req, err := http.NewRequest("DELETE", BambooApiUrl + path + "/" + agentIds[i], nil)
		if err != nil {
			fmt.Println("Unable to delete agent. Error: %s", err)
			return err
		}
		req.Header.Add("Authorization", "Basic " + base64Creds)
		setupLog.Info("Deleting agent: " + agentIds[i])
		_, err = client.Do(req)
		if err != nil {
			setupLog.Error(err, "Failed to delete agent")
		}
	}
	return nil
}

func GetOnlineIdleAgents(path string, base64Creds string) (err error, number int64) {
	var result []map[string]interface{}
	client := &http.Client{}
	req, err := http.NewRequest("GET", BambooApiUrl+path, nil)
	if err != nil {

	}
	req.Header.Add("Authorization", "Basic " + base64Creds)
	resp, err := client.Do(req)
	if err != nil {
		return err, -1
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &result)
		if err != nil {
			fmt.Println(err)
		}
		for i := range result {
			agentState := result[i]["busy"]
			if agentState == false {
				number = number + 1
			}
		}
		defer resp.Body.Close()
		return nil, int64(number)
	}
}


func GetAgentStatus(path string, id string, base64Creds string) (err error, busy bool) {
	var result []map[string]interface{}
	client := &http.Client{}
	req, err := http.NewRequest("GET", BambooApiUrl+path, nil)
	if err != nil {

	}
	req.Header.Add("Authorization", "Basic " + base64Creds)
	resp, err := client.Do(req)
	if err != nil {
		return err, false
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return err, false
		}
		err = json.Unmarshal(b, &result)
		if err != nil {
			fmt.Println(err)
			return err, false
		}
		for i := range result {
			agentId := result[i]["id"]
			if strconv.FormatFloat(agentId.(float64), 'f', -1, 64) == id {
				busyStatus := result[i]["busy"]
				if busyStatus == true {
					return nil, true
				}
			}
		}
		defer resp.Body.Close()
		return nil, false
	}
}

func GetAgentIdByName(path string, names []string, bamboo *installv1alpha1.Bamboo, base64Creds string) (err error, agentIds []string ) {
	agentIds = []string{}
	var result []map[string]interface{}
	client := &http.Client{}
	req, err := http.NewRequest("GET", BambooApiUrl+path, nil)
	if err != nil {

	}
	req.Header.Add("Authorization", "Basic " + base64Creds)
	resp, err := client.Do(req)
	if err != nil {
		return err, agentIds
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return err, nil
		}
		err = json.Unmarshal(b, &result)
		if err != nil {
			return err, nil
			fmt.Println(err)
		}
		for i := range result {
			for n := range names {
				agentName := result[i]["name"]
				nameString := fmt.Sprintf("%v", agentName)
				contains := strings.Contains(nameString, bamboo.Name + "-agent-" + names[n])
				if contains {
					id := result[i]["id"]
					agentIds = append(agentIds, strconv.FormatFloat(id.(float64), 'f', -1, 64))


				}
			}
		}
		defer resp.Body.Close()
	}
	return nil, agentIds

}

type BuildQueue struct {
	Expand string `json:"expand"`
	Link   struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"link"`
	QueuedBuilds struct {
		Expand      string        `json:"expand"`
		Max_result  int64         `json:"max-result"`
		QueuedBuild []interface{} `json:"queuedBuild"`
		Size        int64         `json:"size"`
		Start_index int64         `json:"start-index"`
	} `json:"queuedBuilds"`
}

func GetQueueSize(path string, base64Creds string) (err error, queueSize int64) {
	var buildQueue BuildQueue
	client := &http.Client{}
	req, err := http.NewRequest("GET", BambooApiUrl+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", "Basic " + base64Creds)
	resp, err := client.Do(req)
	if err != nil {
		return err, -1
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return err, -1
		}
		err = json.Unmarshal(b, &buildQueue)
		if err != nil {
			fmt.Println(err)
			return err, -1
		}
		queueSize := buildQueue.QueuedBuilds.Size
		defer resp.Body.Close()
		return nil, queueSize
	}
}
