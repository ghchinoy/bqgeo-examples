# ziplookup


An API that queries BigQuery public datasets for the zipcodes within a given city and state.


Deploy this then call it via URL pattern `/ziplookup/STATE/CITY` ex. `/ziplookup/CA/Mountain%20View`

And receive JSON such as

```
[
  {
    "zip": "92392",
    "city": "Hesperia city, Oak Hills CDP, Victorville city, Phelan CDP, Mountain View Acres CDP",
    "county": "San Bernardino County",
    "state": "CA"
  },
  {
    "zip": "94024",
    "city": "Loyola CDP, Mountain View city, Cupertino city, Los Altos city, Sunnyvale city, Los Altos Hills town",
    "county": "Santa Clara County",
    "state": "CA"
  }
]
```


The BigQuery query used is:


```
SELECT zip_code, city, county, state_code 
FROM `bigquery-public-data.geo_us_boundaries.zip_codes` 
WHERE 
  city LIKE "%Mountain View%" AND
  state_code = "CA"
ORDER BY zip_code
LIMIT 30
```


## Deploy as Cloud Run service

Deploying to Cloud Run with the Google Cloud SDK by typing `gcloud run deploy ziplookup --source .`.

Use service account to run the service vs the default account:

### Create a Service Account to use for Cloud Run

```
export PROJECT_ID=$(gcloud info --format='value(config.project)')
export SERVICE_ACCOUNT=ziplookup-sa@${PROJECT_ID}.iam.gserviceaccount.com
gcloud iam service-accounts create ziplookup-sa \
  --display-name "Ziplookup Cloud Run"
gcloud projects add-iam-policy-binding ${PROJECT_ID} --member \
  serviceAccount:${SERVICE_ACCOUNT} \
  --role=roles/bigquery.jobUser
```

### Clone this repo

```
git clone https://github.com/ghchinoy/bqgeo-examples.git
cd ziplookup
```

### Deploy with service account

```
gcloud run deploy ziplookup --source . \
  --region us-central1 --service-account ${SERVICE_ACCOUNT}
```





