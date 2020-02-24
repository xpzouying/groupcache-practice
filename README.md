# groupcache-practice
The practice for groupcache

Create the project use [groupcache](https://github.com/golang/groupcache)

Environments:

- based on custom slow database for test
- multiple cache nodes for HA


TODO list:

- [x] write custom database for test
- [x] write cache server to provide API for caller outside
- [ ] write outside tester

**groupcache version**

- Date: 2020-02-24
- Commit ID: 8c9f03a8e57eb486e42badaed3fb287da51807ba


**Run step:**

1. run database server

```bash
cd database && go run main.go
```

2. insert data into database

```bash
# insert one entry:
# key: name, value: zouying
curl -H "Content-Type: application/json" -X POST -d '{"key": "name", "value": "zouying"}' http://localhost:9000/set

# check insert
curl -H "Content-Type: application/json" -X POST -d '{"key": "name"}' http://localhost:9000/get
```

3. run frontend (include cache). Two node in cache cluster.

```bash
cd frontend

# run the first node
go run ./main.go -addr=":8001" -port ":18001"

# run the second node
go run ./main.go -addr=":8002" -port ":18002"
```

4. get the value, the first try will get from database
```bash
# try the first node api
curl -H "Content-Type: application/json" -X POST -d '{"key": "name"}' http://localhost:18001/get

# try the second node api
curl -H "Content-Type: application/json" -X POST -d '{"key": "name"}' http://localhost:18002/get
```