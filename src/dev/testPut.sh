
#!/bin/bash

export uuid=`curl -H "Content-Type: application/json" -s http://localhost:8083/api/v1/namespace/pavedroad.io/usersLIST/ | jq -r '.UUID'`

echo "UUID for user test is :  $uuid"

curl -H "Content-Type: application/json" \
     -X PUT \
     -d @usersPutData.json \
     -v http://localhost:8083/api/v1/namespace/pavedroad.io/users/$uuid
