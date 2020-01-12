# Running tests locally

Start up redis and minio. Either in a container or locally. 

e.g for minio:
```
$ docker run -e MINIO_ACCESS_KEY=minio -e MINIO_SECRET_KEY=insecure -p 9000:9000  minio/minio server start
```
