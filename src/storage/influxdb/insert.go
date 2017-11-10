package influxdb

import (
	"fmt"
	"log"
	"time"

	clientPkg "github.com/influxdata/influxdb/client/v2"
)

func (conn *InfluxdbConnection) InsertVisits(visits []map[string]string) {
	log.Printf("Creating InfluxDB database %s...", conn.databaseName)
	command := fmt.Sprintf("CREATE DATABASE %s", conn.databaseName)
	_, err := conn.query(command)
	if err != nil {
		log.Fatalf("Error from %s: %s", command, err)
	}

	// Create a batch
	points, err := clientPkg.NewBatchPoints(clientPkg.BatchPointsConfig{
		Database:  conn.databaseName,
		Precision: "s",
	})
	if err != nil {
		log.Fatalf("Error from NewBatchPoints: %s", err)
	}

	for _, visit := range visits {
		timestamp := time.Unix(toInt("time", visit["time"]), 0)

		tags := map[string]string{
			"host": visit["host"],
		}
		fields := map[string]interface{}{
			"duration":           toFloat("duration", visit["duration"]),
			"response_size":      toInt("response_size", visit["response_size"]),
			"header_size":        toInt("header_size", visit["header_size"]),
			"trace":              visit["trace"],
			"server_region":      visit["server_region"],
			"protocol":           visit["protocol"],
			"property_name":      visit["property_name"],
			"status":             visit["status"], // don't insert as int
			"remote_addr":        visit["remote_addr"],
			"request_method":     visit["request_method"],
			"uri":                visit["uri"],
			"user_agent":         visit["user_agent"],
			"referer":            visit["referer"],
			"content_type":       visit["content_type"],
			"cache_status":       visit["cache_status"],
			"geo_continent":      visit["geo_continent"],
			"geo_continent_code": visit["geo_continent_code"],
			"geo_country":        visit["geo_country"],
			"geo_country_code":   visit["geo_country_code"],
		}
		point, err := clientPkg.NewPoint(
			conn.measurementName, tags, fields, timestamp)
		if err != nil {
			log.Fatalf("Error from NewPoint: %s", err)
		}
		points.AddPoint(point)
	}

	log.Printf("Inserting %d InfluxDB points...", len(points.Points()))
	if err := conn.client.Write(points); err != nil {
		log.Fatalf("Error from Write: %s", err)
	}
}