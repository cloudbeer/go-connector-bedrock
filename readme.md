# A go proxy for bedrock

Run:

```shell
docker run --rm -d --pull=always --name gobr  \
 -p 8081:8081 \
 -e AWS_REGION=us-west-2 \
 -e AWS_ACCESS_KEY_ID=Axxxxx \
 -e AWS_SECRET_ACCESS_KEY=w+xxxxxxxxxxxxxxxxxxxxx \
       cloudbeer/go-connector-for-bedrock
```
