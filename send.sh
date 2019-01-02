#!/bin/bash

for i in {1..12}
do
   MYSTR1="test out the server"
   printf "\n sending request:: $MYSTR1 \n"
   echo -n $MYSTR1 | nc localhost 7777
done

for i in {1..17}
do
   MYSTR2="something different"
   printf "\n sending request:: $MYSTR2 \n"
   echo -n $MYSTR2 | nc localhost 7777
done

for i in {1..7}
do
   MYSTR3="hhhhhhhhh aaaaaaaaa"
   printf "\n sending request:: $MYSTR3 \n"
   echo -n $MYSTR3 | nc localhost 7777
done

# The above produces a cluster of three.
