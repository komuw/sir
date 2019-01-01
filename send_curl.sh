#!/bin/bash

# usage
# go run -race reverse.go -p localhost:7777 -r httpbin.org:443
# bash send_curl.sh

for i in {1..6}
do
   printf "\n sending request:: \n"
   curl -vL \
   -H "accept: application/json" \
   -H "Content-Type: application/json" \
   -H "Host: httpbin.org" \
   -d '{"name":"komu"}' \
   localhost:7777/post
done

for i in {1..3}
do
   printf "\n sending request:: \n"
   curl -vL \
   -H "accept: application/json" \
   -H "Content-Type: application/json" \
   -H "Host: httpbin.org" \
   -d '{"name":"juma"}' \
   localhost:7777/post
done

for i in {1..4}
do
   printf "\n sending request:: \n"
   curl -vL \
   -H "accept: application/json" \
   -H "Content-Type: application/json" \
   -H "Host: httpbin.org" \
   -d '{"name":"john"}' \
   localhost:7777/post
done

# Note: for this to work with our current dbscan code;
# the Content-Length of the 3 requests should be the same.
# thats why we have used;
# -d '{"name":"komu"}', -d '{"name":"juma"}', -d '{"name":"john"}'