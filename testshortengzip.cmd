curl -v -X POST --compressed -H "Accept-Encoding: gzip, deflate" -H "Content-Type: application/json" -H "Content-Encoding:gzip" --data-binary @shortenbody.json.gz http://localhost:8080/api/shorten