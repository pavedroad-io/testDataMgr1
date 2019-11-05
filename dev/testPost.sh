
#!/bin/bash

curl -H "Content-Type: application/json" \
     -X POST \
     -d @users.json \
     -v http://localhost:8083/api/v1/namespace/pavedroad.io/users
