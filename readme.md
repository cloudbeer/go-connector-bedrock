# A go proxy for bedrock

当前只支持 chat, 固定了几个模型。

多模态和 tool 正在开发中...

Run:

```shell
docker run --rm -d --pull=always --name gobr  \
 -p 8081:8081 \
 -e AWS_REGION=us-west-2 \
 -e AWS_ACCESS_KEY_ID=Axxxxx \
 -e AWS_SECRET_ACCESS_KEY=w+xxxxxxxxxxxxxxxxxxxxx \
       cloudbeer/go-connector-for-bedrock
```
