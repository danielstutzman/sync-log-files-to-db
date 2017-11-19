package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var discoveryJson []byte
var DATASETS_REGEXP = regexp.MustCompile("^/bigquery/v2/projects/(.*?)/datasets$")
var TABLES_REGEXP = regexp.MustCompile("^/bigquery/v2/projects/(.*?)/datasets/(.*?)/tables$")
var JOBS_REGEXP = regexp.MustCompile("^/bigquery/v2/projects/(.*?)/jobs$")
var QUERY_REGEXP = regexp.MustCompile("^/bigquery/v2/projects/(.*?)/queries/(.*?)$")
var INSERT_REGEXP = regexp.MustCompile("^/projects/(.*?)/datasets/(.*?)/tables/(.*?)/insertAll")

func serveDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Write(discoveryJson)
}

func serveDatasets(w http.ResponseWriter, r *http.Request, project string) {
	dataset := "belugacdn_logs"
	fmt.Fprintf(w, `{
		"kind": "bigquery#datasetList",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/qwnfLrlOKTXd94DjXLYMd9AnLA8\"",
		"datasets": [
		 {
			"kind": "bigquery#dataset",
			"id": "%s:%s",
			"datasetReference": {
			 "datasetId": "%s",
			 "projectId": "%s"
			}
		 }
		]
	 }`, project, dataset, dataset, project)
}

func serveTables(w http.ResponseWriter, r *http.Request, project, dataset string) {
	table := "visits"
	fmt.Fprintf(w, `{
		"kind": "bigquery#tableList",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/zZCSENSD7Bu0j7yv3iZTn_ilPBg\"",
		"tables": [
			{
				"kind": "bigquery#table",
				"id": "%s:%s.%s",
				"tableReference": {
					"projectId": "%s",
					"datasetId": "%s",
					"tableId": "%s"
				},
				"type": "TABLE",
				"creationTime": "1510171319097"
			}
		],
		"totalItems": 1
		}
	`, project, dataset, table, project, dataset, table)
}

func startJob(w http.ResponseWriter, r *http.Request, project string) {
	email := "a@b.com"
	dataset := "belugacdn_logs"
	table := "jobs"
	fmt.Fprintf(w, `{
		"kind": "bigquery#job",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/_oiKSu1NLem_L8Icwp_IYkfy3vg\"",
		"id": "%s:bqjob_r7c51234c0123569f_0000015fd1968828_1",
		"selfLink": "https://www.googleapis.com/bigquery/v2/projects/%s/jobs/bqjob_r7c51234c0123569f_0000015fd1968828_1",
		"jobReference": {
		 "projectId": "%s",
		 "jobId": "bqjob_r7c51234c0123569f_0000015fd1968828_1"
		},
		"configuration": {
		 "query": {
			"query": "select count(*) from %s.%s",
			"destinationTable": {
			 "projectId": "%s",
			 "datasetId": "_2cf7cfaa9c05dd2381014b72df999b53fd45fe85",
			 "tableId": "anon5fb7e0264db7f54e07e3df0833fbfcfd11d63e03"
			},
			"createDisposition": "CREATE_IF_NEEDED",
			"writeDisposition": "WRITE_TRUNCATE"
		 }
		},
		"status": {
		 "state": "DONE"
		},
		"statistics": {
		 "creationTime": "1511049825816",
		 "startTime": "1511049826072"
		},
		"user_email": "%s"
	 }`, project, project, project, dataset, table, project, email)
}

func serveQuery(w http.ResponseWriter, r *http.Request, project string) {
	fmt.Fprintf(w, `{
		"kind": "bigquery#getQueryResultsResponse",
		"etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/wLFL5h11OCxiWY3yDLqREwltkXs\"",
		"schema": {
			"fields": [
			{
				"name": "f0_",
				"type": "INTEGER",
				"mode": "NULLABLE"
			}
			]
		},
		"jobReference": {
			"projectId": "%s",
			"jobId": "bqjob_r6c744039b40f818e_0000015fd19a3130_1"
		},
		"totalRows": "1",
		"rows": [
			{
				"f": [
					{
						"v": "704"
					}
				]
			}
		],
		"totalBytesProcessed": "0",
		"jobComplete": true,
		"cacheHit": true
	}`, project)
}

func insertRows(w http.ResponseWriter, r *http.Request, project, dataset, table string) {
	decoder := json.NewDecoder(r.Body)
	var body map[string]interface{}
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	log.Println(body)

	// No errors implies success
	fmt.Fprintf(w, `{
		"kind": "bigquery#tableDataInsertAllResponse"
	}`)
}

func serve(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Incoming path: %s", path)

	if path == "/discovery/v1/apis/bigquery/v2/rest" {
		serveDiscovery(w, r)
		return
	}
	if match := DATASETS_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[1]
		serveDatasets(w, r, project)
		return
	}
	if match := TABLES_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[1]
		dataset := match[2]
		serveTables(w, r, project, dataset)
		return
	}
	if match := JOBS_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[1]
		startJob(w, r, project)
		return
	}
	if match := QUERY_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[1]
		serveQuery(w, r, project)
		return
	}

	if match := INSERT_REGEXP.FindStringSubmatch(path); match != nil {
		project := match[1]
		dataset := match[2]
		table := match[3]
		insertRows(w, r, project, dataset, table)
		return
	}

	log.Fatalf("Don't know how to serve path %s", r.URL.Path)
}

func main() {
	discoveryJsonPath := flag.String("discovery-json-path", "", "path to discovery.json")
	flag.Parse()

	if *discoveryJsonPath == "" {
		log.Fatalf("Please specify a -discovery-json-path")
	}

	var err error
	discoveryJson, err = ioutil.ReadFile(*discoveryJsonPath)
	if err != nil {
		panic(err)
	}
	discoveryJson = bytes.Replace(discoveryJson,
		[]byte("https://www.googleapis.com"),
		[]byte("http://localhost:9090"),
		-1)

	http.HandleFunc("/", serve)

	log.Printf("Listening on :9090...")
	err = http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
