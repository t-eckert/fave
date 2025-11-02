#!/bin/bash

curl -X POST http://localhost:8080/bookmarks \
	-H "Content-Type: application/json" \
	-d '{
		   "name": "Example Bookmark",
		   "url": "https://www.example.com",
		   "tags": ["example", "test"],
		   "description": "This is an example bookmark."
		 }'
curl -X GET http://localhost:8080/bookmarks | jq
curl -X GET http://localhost:8080/bookmarks/1 | jq
curl -X PUT http://localhost:8080/bookmarks/1 \
	-H "Content-Type: application/json" \
	-d '{
		   "name": "Updated Bookmark",
		   "url": "https://www.updated-example.com",
		   "tags": ["updated", "test"],
		   "description": "This is an updated bookmark."
		 }' | jq
curl -X GET http://localhost:8080/bookmarks | jq
curl -X GET http://localhost:8080/bookmarks/1 | jq
curl -X DELETE http://localhost:8080/bookmarks/1 | jq
