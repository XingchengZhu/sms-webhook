<<<<<<< HEAD
# sms-webhook

=======
# SMS Webhook

`SMS Webhook` 是一个处理 Prometheus Alertmanager 告警的服务。它接收来自 Alertmanager 的告警，并将告警信息转发到配置的短信接口。

## 功能

- 接收 Prometheus Alertmanager 的告警通知。
- 从告警中提取必要信息。
- 通过配置的短信 API 发送短信通知。

## 前提

- Webhook 接收消息体为字符串。
- SMS API 需要是特定支持的短信接口和参数。

## 环境变量

在运行应用之前，请确保设置以下环境变量：

- `SMS_API_URL`：短信 API 的 URL。默认值为 `https://default-sms-api-url.com/send`。
- `SMS_CODE`：短信接口的代码标识。默认值为 `ALERT_CODE`。
- `SMS_TARGET`：短信目标手机号码。默认值为 `15222222222`。
- `PORT`：服务监听的端口。默认值为 `8080`。
- `LOG_LEVEL`：日志等级。默认值为 `info`。可选值包括 `debug`, `info`, `warn`, `error`。

## 快速开始

### 构建镜像

使用以下命令来构建 Docker 镜像：

```bash
docker build -t your-dockerhub-username/sms-webhook:latest .
```

指定构建不同架构的镜像：

```bash
docker build --platform linux/amd64 -t your-dockerhub-username/sms-webhook:latest .
```

## 部署

### 运行容器

使用以下命令来运行 Docker 容器：

```bash
docker run -d -p 8080:8080 \
  -e SMS_API_URL="https://your-sms-api-url.com/send" \
  -e SMS_CODE="your-sms-code" \
  -e SMS_TARGET="your-sms-target" \
  -e PORT="8080" \
  -e LOG_LEVEL="info" \
  your-dockerhub-username/sms-webhook:latest
```

### Kubernetes 部署

使用 Kubernetes 部署应用时，请使用根目录`deploy.yaml`文件。

 1. 将配置文件配置文件`deploy.yaml`复制到集群`kubectl`可以执行的节点。
 2. 根据实际情况修改其中的参数。
 3. 使用以下命令将配置应用到 Kubernetes 集群：

```bash
kubectl apply -f deploy.yaml
```

  4.验证服务是否已正确部署：

```bash
  kubectl -n sms-webhook get pods
  kubectl -n sms-webhook get svc
```

  5.通过日志查看应用状态：

  ```bash
  kubectl -n sms-webhook logs <your-pod-name>
  ```

  6.配置Webhook：

  ```bash
  http://sms-webhook-service.sms-webhook.svc.cluster.local/webhook
  ```
>>>>>>> init
