#!/bin/bash

for i in {1..50}
do
   MYSTR1="test out the server"
   printf "\n sending request:: $MYSTR1 \n"
   curl -vIkL -H "Host: google.com" localhost:7777
done

for i in {1..37}
do
   MYSTR2="something different"
   printf "\n sending request:: $MYSTR2 \n"
   curl -vIkL -H "Host: google.com" localhost:7777
done

for i in {1..11}
do
   MYSTR3="hhhhhhhhh aaaaaaaaa"
   printf "\n sending request:: $MYSTR3 \n"
   curl -vIkL -H "Host: google.com" localhost:7777
done

# The above produces a cluster of three.

# Note: for this to work with our current dbscan code;
# len($MYSTR1) == len($MYSTR2)
# ie,
# len("test out the server")  == len("something different")