# sms-webhook

一个很小的 HTTP 服务，用来把告警（比如 K8S / Alertmanager / 其他系统的 Webhook）转发到一个或多个短信网关。

## 特性

- 一个入口 `/webhook`
- 支持在报文里写通道：`渠道: json,header`
- 支持同时配置多种短信接口：`json` / `form` / `header-json`
- 支持两种发送模式：
  - `pick`（默认）：只发一条，按注册顺序或叫 `default` 的那条发
  - `broadcast`：全部通道都发一遍
- 支持 K8S 部署，带一个假短信服务（echo-server）方便调试

## 环境变量

| 变量名              | 说明                                  | 默认值 |
|---------------------|---------------------------------------|--------|
| `PORT`              | http 端口                             | `8080` |
| `LOG_LEVEL`         | 日志级别                              | `info` |
| `SMS_API_URL`       | 老的单通道短信接口地址                 | -      |
| `SMS_CODE`          | 老的单通道里要带的 code                | `ALERT_CODE` |
| `SMS_TARGET`        | 默认的手机号                           | `15222222222` |
| `SMS_PROVIDER`      | 老的单通道类型：`json`/`form`/`header-json` | `json` |
| `SMS_SEND_MODE`     | `pick` or `broadcast`                 | `pick` |
| `SMS_PROVIDERS_JSON`| 多通道配置（JSON 数组）                | 空     |

### `SMS_PROVIDERS_JSON` 示例

```json
[
  {
    "name": "json",
    "kind": "json",
    "url": "http://fake-sms.sms-webhook.svc.cluster.local:9999/json",
    "code": "ALERT_CODE"
  },
  {
    "name": "header",
    "kind": "header-json",
    "url": "http://fake-sms.sms-webhook.svc.cluster.local:9999/header",
    "code": "ALERT_CODE",
    "api_key": "abc123",
    "header_key": "X-API-KEY"
  },
  {
    "name": "form",
    "kind": "form",
    "url": "http://fake-sms.sms-webhook.svc.cluster.local:9999/form",
    "code": "FORM_CODE",
    "phone_field": "target",
    "content_field": "content",
    "code_field": "code"
  }
]
