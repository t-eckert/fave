#!/bin/bash

go build .
./fave serve &

echo "List before adding:"
./fave list
echo "Adding a favorite..."
./fave add "Blog Post" "https://example.com/blog-post"
echo "List after adding:"
./fave list
echo "Getting the favorite with ID 1:"
./fave get 1
echo "Deleting the favorite with ID 1:"
./fave delete 1
echo "List after deleting:"
./fave list

kill %1
