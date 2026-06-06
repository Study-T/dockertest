# API 文档

> ns-tracking-go 云途 Webhook 服务

## 架构说明

```
┌─────────────────────────────────────────────────────────────┐
│                    云途 Webhook                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              对方的系统（Ruby）                               │
│              写入 Redis 队列                                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              Redis 队列 (queue:yun_express_webhook_track)     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              Queue Worker（BRPOP 监听）                       │
│              读取 → 标准化 → 写入 PostgreSQL → 更新 Redis     │
└─────────────────────────────────────────────────────────────┘
```

---

## POST /webhook

接收云途 Webhook 推送。

### 请求头

| Header | 必填 | 说明 |
|--------|------|------|
| `Content-Type` | 否 | `application/json`（可省略） |
| `X-Openapi-Request-Timestamp` | 是 | 毫秒时间戳 |
| `X-Openapi-Signature` | 是 | SHA256(timestamp + encryptKey + body) |

### 请求体

三种格式均可：

**格式 1：tisPushData（直接 body）**
```json
{
  "data": {
    "waybill_number": "YT2615000705163221",
    "tracking_number": "3SWLT0037525480",
    "package_status": "T",
    "track_events": [
      {
        "track_node_code": "ORDER_CREATION",
        "track_node_description": "Order created",
        "process_time": "2026-05-30T16:00:00",
        "process_utc_time": "2026-05-30T08:00:00"
      }
    ]
  },
  "data_code": "tisPushData"
}
```

**格式 2：标准信封**
```json
{
  "customerCode": "CNHC318962",
  "timestamp": "1717312800000",
  "body": { ... }
}
```

**格式 3：AES 加密**
```json
{
  "encrypt": "base64编码的密文"
}
```

### 响应

```json
{"status": "ok"}
```

### 错误码

| HTTP 状态码 | 说明 |
|-------------|------|
| 200 | 成功（含重复事件） |
| 400 | 请求体无效 |
| 401 | 签名验证失败 |
| 405 | 非 POST 方法 |
| 415 | Content-Type 不支持 |
| 500 | 内部错误 |

---

## GET /health

健康检查。

### 响应

```json
{"status": "ok"}
```

---

## GET /admin/tracking/:order_number

按订单号查询轨迹详情。

### 路径参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `order_number` | 是 | 运单号（waybill_number） |

### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "order_number": "YT2615000705163221",
      "package_status": "D",
      "track_Info": {
        "运单号": "YT2615000705163221",
        "tracking_number": "3SWLT0037525480",
        "customer_order_number": "BG260530152213025990",
        "product_code": "MUZXR",
        "product_name": "云途全球化妆品类专线挂号",
        "channel_code": "NLDHL",
        "check_in_time": "2025-06-02T07:15:23Z",
        "check_out_time": "2025-06-02T08:40:28Z",
        "pick_up_time": "0001-01-01T00:00:00Z",
        "customer_code": "CNHC318962",
        "origin_code": "CN",
        "destination_code": "NL",
        "postal_code": "",
        "actual_weight": 0.36,
        "interval_day": 3,
        "interval_work_day": 3,
        "last_mile_site": "https://www.postnl.nl/en/track-trace/3SWLT0037525480",
        "last_mile_name": "PostNL",
        "phone_number": "",
        "track_events": [
          {
            "process_time": "2025-05-30T16:00:00Z",
            "process_utc_time": "2025-05-30T16:00:00Z",
            "process_content": "已收到发货信息",
            "process_country": "",
            "process_province": "",
            "process_city": "",
            "process_location": "",
            "track_node_code": "ORDER_CREATION",
            "track_node_description": "已收到货物信息",
            "pod_url": ""
          }
        ],
        "pod_url": "https://example.com/pod/yt2615000705163221.jpg",
        "pod_urls": null,
        "IsSignature": false,
        "SignatureUrls": null,
        "EstimatedDeliveryToDateZone": "",
        "EstimatedDeliveryFromDateZone": ""
      }
    }
  ]
}
```

## POST /admin/batch-query

批量查询。

## GET /public/query

公开查询。

## POST /public/batch-query

公开批量查询。

---

## GET /tracking/:order_number

按订单号查询轨迹详情（公开接口）。

### 路径参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `order_number` | 是 | 运单号（waybill_number） |

### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "order_number": "YT2615000705163221",
      "package_status": "D",
      "track_Info": {
        "运单号": "YT2615000705163221",
        "tracking_number": "3SWLT0037525480",
        "customer_order_number": "BG260530152213025990",
        "product_code": "MUZXR",
        "product_name": "云途全球化妆品类专线挂号",
        "channel_code": "NLDHL",
        "check_in_time": "2025-06-02T07:15:23Z",
        "check_out_time": "2025-06-02T08:40:28Z",
        "pick_up_time": "0001-01-01T00:00:00Z",
        "customer_code": "CNHC318962",
        "origin_code": "CN",
        "destination_code": "NL",
        "postal_code": "",
        "actual_weight": 0.36,
        "interval_day": 3,
        "interval_work_day": 3,
        "last_mile_site": "https://www.postnl.nl/en/track-trace/3SWLT0037525480",
        "last_mile_name": "PostNL",
        "phone_number": "",
        "track_events": [
          {
            "process_time": "2025-05-30T16:00:00Z",
            "process_utc_time": "2025-05-30T16:00:00Z",
            "process_content": "已收到发货信息",
            "process_country": "",
            "process_province": "",
            "process_city": "",
            "process_location": "",
            "track_node_code": "ORDER_CREATION",
            "track_node_description": "已收到货物信息",
            "pod_url": ""
          }
        ],
        "pod_url": "https://example.com/pod/yt2615000705163221.jpg",
        "pod_urls": null,
        "IsSignature": false,
        "SignatureUrls": null,
        "EstimatedDeliveryToDateZone": "",
        "EstimatedDeliveryFromDateZone": ""
      }
    }
  ]
}
```
