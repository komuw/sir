#!/bin/bash

# usage
# go run -race reverse.go -p localhost:7777 -r httpbin.org:80
# bash send_curl.sh

for i in {1..22}
do
   printf "\n sending request:: \n"
   curl -vkL \
   -H "accept: application/json" \
   -H "Content-Type: application/json" \
   -H "Host: httpbin.org" \
   -d '{"name":"komu"}' \
   localhost:7777/post
done

for i in {1..17}
do
   printf "\n sending request:: \n"
   curl -vkL \
   -H "accept: application/json" \
   -H "Content-Type: application/json" \
   -H "Host: httpbin.org" \
   -d '{"name":"juma"}' \
   localhost:7777/post
done

for i in {1..11}
do
   printf "\n sending request:: \n"
   curl -vkL \
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