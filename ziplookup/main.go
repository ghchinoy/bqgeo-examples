package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/compute/metadata"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/api/iterator"
)

var projectID string

type ZipCode struct {
	Zip       string `bigquery:"zip_code" json:"zip"`
	City      string `json:"city"`
	County    string `json:"county"`
	StateCode string `bigquery:"state_code" json:"state"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	projectID = getProjectID()
	if projectID == "" {
		log.Printf("requires PROJECT_ID")
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/ziplookup/{state}/{city}", ZipLookupHandler)
	http.Handle("/", r)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func ZipLookupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Printf("lookup for %s, %s", vars["city"], vars["state"])

	ctx := context.Background()
	codes, err := cityStateQuery(ctx, vars["city"], vars["state"])
	if err != nil {
		log.Printf("cityStateQuery error: %v", err)
		http.Error(w, "unable to query", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(codes)
}

// cityStateQuery queries the BigQuery public table for City, State and return Zip Codes
func cityStateQuery(ctx context.Context, city, state string) ([]ZipCode, error) {
	var zipcodes []ZipCode

	c, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return zipcodes, err
	}

	title := cases.Title(language.English)
	queryString := fmt.Sprintf(`SELECT zip_code, city, county, state_code 
		FROM `+"`bigquery-public-data.geo_us_boundaries.zip_codes`"+` 
		WHERE 
		city LIKE "%%%s%%" AND
		state_code = "%s"
		ORDER BY zip_code
		LIMIT 50`, title.String(city), strings.ToUpper(state))

	q := c.Query(queryString)
	it, err := q.Read(ctx)
	if err != nil {
		return zipcodes, err
	}
	for {
		var zipcode ZipCode
		err := it.Next(&zipcode)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return zipcodes, err
		}
		//log.Printf("%+v", zipcode)
		zipcodes = append(zipcodes, zipcode)
	}

	return zipcodes, nil
}

func getProjectID() string {
	c := metadata.NewClient(&http.Client{})
	p, err := c.ProjectID()
	if err != nil {
		return os.Getenv("PROJECT_ID")
	}
	return p
}
