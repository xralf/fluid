#!/usr/bin/env bash

url=http://localhost:50001/api/v1

example=${PWD}/../../examples/synthetic-slice-time-live

catalogId=$(curl ${url}/catalog/add \
--header "Content-Type: multipart/form-data" \
--request POST \
--silent \
-F "file=@${example}/catalog.json" | jq ".id")
echo $catalogId



# curl ${url}/catalog/add \
# --header "Content-Type: multipart/form-data" \
# --request POST \
# --silent \
# -F "file=@${example}/catalog.json" | jq "."










#query="from instance1.database1.schema1.table1 group by g2, g1 window session begin when c == "a" or c == "b" or c == "c" end when c == "x" or c == "y" or c == "z" inclusive expire after 5 seconds aggregate count(a) as aCount, avg(a) as aAvg, avg(b) as bAvg, sum(a) as aSum, first(c) as cFirst, last(c) as cLast, first(t2) as t2First, last(t2) as t2Last, first(rowid) as rowidFirst, last(rowid) as rowidLast, count(one) as oneCount, sum(one) as oneSum append t2First, t2Last, aCount, aSum, aAvg, bAvg, oneCount, oneSum, cFirst, cLast, rowidFirst, rowidLast where bAvg > 50 to instance1.database1.schema1.table2"
## Read query file and replace newlines by space.
# query=$(cat ../fluid/examples/session/query.fql)
# query=$(echo $query | sed 's/\n/ /g')
# query=$(echo $query | sed 's/"/\\"/g')
# echo ${query}

# echo "hello" > ~/trash/hello1.txt
# curl ${url}/catalog/upload \
# --header "Content-Type: multipart/form-data" \
# --request POST \
# -F "file=@/tmp/trash/hello1.txt"

# curl ${url}/health
# curl ${url}/job/list/10

# curl \
# --header "Content-Type: application/json" \
# --request POST \
# --data '{"query":"aaa"}' \
# ${url}/job/add | jq ".data.id"




# curl \
# --header "Content-Type: application/json" \
# --request POST \
# --data '{"query":"'"$query"'"}' \
# ${url}/query/add

# curl ${url}/catalog/list/10
# curl ${url}/job/list/10










# curl ${url}/query/add \
# --header "Content-Type: multipart/form-data" \
# --request POST \
# --silent \
# -F "file=@/tmp/git/xralf/fluid/examples/session/query.fql" | jq "."

# curl ${url}/catalog/add \
# --header "Content-Type: multipart/form-data" \
# --request POST \
# --silent \
# -F "file=@/tmp/git/xralf/fluid/examples/session/catalog.json" | jq "."

# curl ${url}/job/add \
# --header "Content-Type: application/json" \
# --request POST \
# --silent \
# --data '{"queryId":"5160e180-4b7d-4874-b7c6-6352b5d737b1", "catalogId":"029fa70b-5f7b-4be1-87e8-f28683c5d98f"}' | jq "."


# curl \
# --header "Content-Type: application/json" \
# --request POST \
# --data '{"id":"335cbb34-7fa7-4a44-a5d6-10a034907b18"}' \
# ${url}/query/delete

# curl \
# --header "Content-Type: application/json" \
# --request POST \
# --data '{"id":"335cbb34-7fa7-4a44-a5d6-10a034907b18"}' \
# ${url}/catalog/delete

#curl --silent ${url}/job/list/10 | jq "."
