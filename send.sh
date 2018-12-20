#!/bin/bash

for i in {1..55}
do
   MYSTR1="test out the server"
   printf "\n sending request:: $MYSTR1"
   echo -n $MYSTR1 | nc localhost 7777
done


for i in {1..55}
do
   MYSTR2="something different"
   printf "\n sending request:: $MYSTR2"
   echo -n $MYSTR2 | nc localhost 7777
done

# Note: for this to work with our current dbscan code;
# len($MYSTR1) == len($MYSTR2)
# ie,
# len("test out the server")  == len("something different")