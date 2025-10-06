# MinerU API 文档

## 单个文件解析

### 创建解析任务

#### 接口说明
适用于通过 API 创建解析任务的场景，用户须先申请 Token。

**注意：**
- 单个文件大小不能超过 200MB，文件页数不超出 600 页
- 每个账号每天享有 2000 页最高优先级解析额度，超过 2000 页的部分优先级降低
- 因网络限制，github、aws 等国外 URL 会请求超时
- 该接口不支持文件直接上传
- header头中需要包含 Authorization 字段，格式为 `Bearer + 空格 + Token`

#### Python 请求示例
```python
import requests

token = "官网申请的api token"
url = "https://mineru.net/api/v4/extract/task"
header = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}
data = {
    "url": "https://cdn-mineru.openxlab.org.cn/demo/example.pdf",
    "is_ocr": True,
    "enable_formula": False,
}
res = requests.post(url, headers=header, json=data)
print(res.status_code)
print(res.json())
print(res.json()["data"])
```

#### CURL 请求示例
```bash
curl --location --request POST 'https://mineru.net/api/v4/extract/task' \
--header 'Authorization: Bearer ***' \
--header 'Content-Type: application/json' \
--header 'Accept: */*' \
--data-raw '{
    "url": "https://cdn-mineru.openxlab.org.cn/demo/example.pdf",
    "is_ocr": true,
    "enable_formula": false
}'
```

#### 请求体参数说明

| 参数 | 类型 | 是否必选 | 示例 | 描述 |
|---|---|---|---|---|
| url | string | 是 | https://static.openxlab.org.cn/opendatalab/pdf/demo.pdf | 文件 URL，支持.pdf、.doc、.docx、.ppt、.pptx、.png、.jpg、.jpeg多种格式 |
| is_ocr | bool | 否 | false | 是否启动 ocr 功能，默认 false |
| enable_formula | bool | 否 | true | 是否开启公式识别，默认 true |
| enable_table | bool | 否 | true | 是否开启表格识别，默认 true |
| language | string | 否 | ch | 指定文档语言，默认 ch |
| data_id | string | 否 | abc** | 解析对象对应的数据 ID |
| callback | string | 否 | http://127.0.0.1/callback | 解析结果回调通知 URL |
| seed | string | 否 | abc** | 随机字符串，用于回调通知请求中的签名 |
| extra_formats | [string] | 否 | ["docx","html"] | 额外导出格式，支持docx、html、latex |
| page_ranges | string | 否 | 1-600 | 指定页码范围 |
| model_version | string | 否 | vlm | mineru模型版本，pipeline或vlm，默认pipeline |

#### 响应示例
```json
{
  "code": 0,
  "data": {
    "task_id": "a90e6ab6-44f3-4554-b4***"
  },
  "msg": "ok",
  "trace_id": "c876cd60b202f2396de1f9e39a1b0172"
}
```

### 获取任务结果

#### Python 请求示例
```python
import requests

token = "官网申请的api token"
url = f"https://mineru.net/api/v4/extract/task/{task_id}"
header = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}
res = requests.get(url, headers=header)
print(res.status_code)
print(res.json())
print(res.json()["data"])
```

#### CURL 请求示例
```bash
curl --location --request GET 'https://mineru.net/api/v4/extract/task/{task_id}' \
--header 'Authorization: Bearer *****' \
--header 'Accept: */*'
```

#### 响应示例
```json
{
  "code": 0,
  "data": {
    "task_id": "47726b6e-46ca-4bb9-******",
    "state": "done",
    "full_zip_url": "https://cdn-mineru.openxlab.org.cn/pdf/018e53ad-d4f1-475d-b380-36bf24db9914.zip",
    "err_msg": ""
  },
  "msg": "ok",
  "trace_id": "c876cd60b202f2396de1f9e39a1b0172"
}
```

## 批量文件解析

### 文件批量上传解析

#### Python 请求示例
```python
import requests

token = "官网申请的api token"
url = "https://mineru.net/api/v4/file-urls/batch"
header = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}
data = {
    "enable_formula": True,
    "language": "ch",
    "enable_table": True,
    "files": [
        {"name":"demo.pdf", "is_ocr": True, "data_id": "abcd"}
    ]
}
file_path = ["demo.pdf"]

try:
    response = requests.post(url, headers=header, json=data)
    if response.status_code == 200:
        result = response.json()
        print('response success. result:{}'.format(result))
        if result["code"] == 0:
            batch_id = result["data"]["batch_id"]
            urls = result["data"]["file_urls"]
            print('batch_id:{},urls:{}'.format(batch_id, urls))
            for i in range(0, len(urls)):
                with open(file_path[i], 'rb') as f:
                    res_upload = requests.put(urls[i], data=f)
                    if res_upload.status_code == 200:
                        print(f"{urls[i]} upload success")
                    else:
                        print(f"{urls[i]} upload failed")
        else:
            print('apply upload url failed,reason:{}'.format(result.msg))
    else:
        print('response not success. status:{} ,result:{}'.format(response.status_code, response))
except Exception as err:
    print(err)
```

#### CURL 请求示例
```bash
curl --location --request POST 'https://mineru.net/api/v4/file-urls/batch' \
--header 'Authorization: Bearer ***' \
--header 'Content-Type: application/json' \
--header 'Accept: */*' \
--data-raw '{
    "enable_formula": true,
    "language": "ch",
    "enable_table": true,
    "files": [
        {"name":"demo.pdf", "is_ocr": true, "data_id": "abcd"}
    ]
}'
```

#### CURL 文件上传示例
```bash
curl -X PUT -T /path/to/your/file.pdf 'https://****'
```

### URL 批量上传解析

#### Python 请求示例
```python
import requests

token = "官网申请的api token"
url = "https://mineru.net/api/v4/extract/task/batch"
header = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}
data = {
    "enable_formula": True,
    "language": "ch",
    "enable_table": True,
    "files": [
        {"url":"https://cdn-mineru.openxlab.org.cn/demo/example.pdf", "is_ocr": True, "data_id": "abcd"}
    ]
}

try:
    response = requests.post(url, headers=header, json=data)
    if response.status_code == 200:
        result = response.json()
        print('response success. result:{}'.format(result))
        if result["code"] == 0:
            batch_id = result["data"]["batch_id"]
            print('batch_id:{}'.format(batch_id))
        else:
            print('submit task failed,reason:{}'.format(result.msg))
    else:
        print('response not success. status:{} ,result:{}'.format(response.status_code, response))
except Exception as err:
    print(err)
```

#### CURL 请求示例
```bash
curl --location --request POST 'https://mineru.net/api/v4/extract/task/batch' \
--header 'Authorization: Bearer ***' \
--header 'Content-Type: application/json' \
--header 'Accept: */*' \
--data-raw '{
    "enable_formula": true,
    "language": "ch",
    "enable_table": true,
    "files": [
        {"url":"https://cdn-mineru.openxlab.org.cn/demo/example.pdf", "is_ocr": true, "data_id": "abcd"}
    ]
}'
```

### 批量获取任务结果

#### Python 请求示例
```python
import requests

token = "官网申请的api token"
url = f"https://mineru.net/api/v4/extract-results/batch/{batch_id}"
header = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}
res = requests.get(url, headers=header)
print(res.status_code)
print(res.json())
print(res.json()["data"])
```

#### CURL 请求示例
```bash
curl --location --request GET 'https://mineru.net/api/v4/extract-results/batch/{batch_id}' \
--header 'Authorization: Bearer *****' \
--header 'Accept: */*'
```

#### 响应示例
```json
{
  "code": 0,
  "data": {
    "batch_id": "2bb2f0ec-a336-4a0a-b61a-241afaf9cc87",
    "extract_result": [
      {
        "file_name": "example.pdf",
        "state": "done",
        "err_msg": "",
        "full_zip_url": "https://cdn-mineru.openxlab.org.cn/pdf/018e53ad-d4f1-475d-b380-36bf24db9914.zip"
      },
      {
        "file_name":"demo.pdf",
        "state": "running",
        "err_msg": "",
        "extract_progress": {
          "extracted_pages": 1,
          "total_pages": 2,
          "start_time": "2025-01-20 11:43:20"
        }
      }
    ]
  },
  "msg": "ok",
  "trace_id": "c876cd60b202f2396de1f9e39a1b0172"
}
```

## 常见错误码

| 错误码 | 说明 | 解决建议 |
|---|---|---|
| A0202 | Token 错误 | 检查 Token 是否正确，请检查是否有Bearer前缀或者更换新 Token |
| A0211 | Token 过期 | 更换新 Token |
| -500 | 传参错误 | 请确保参数类型及Content-Type正确 |
| -10001 | 服务异常 | 请稍后再试 |
| -10002 | 请求参数错误 | 检查请求参数格式 |
| -60001 | 生成上传 URL 失败 | 请稍后再试 |
| -60002 | 获取匹配的文件格式失败 | 检测文件类型失败，请求的文件名及链接中带有正确的后缀名 |
| -60003 | 文件读取失败 | 请检查文件是否损坏并重新上传 |
| -60004 | 空文件 | 请上传有效文件 |
| -60005 | 文件大小超出限制 | 检查文件大小，最大支持 200MB |
| -60006 | 文件页数超过限制 | 请拆分文件后重试 |
| -60007 | 模型服务暂时不可用 | 请稍后重试或联系技术支持 |
| -60008 | 文件读取超时 | 检查 URL 可访问 |
| -60009 | 任务提交队列已满 | 请稍后再试 |
| -60010 | 解析失败 | 请稍后再试 |
| -60011 | 获取有效文件失败 | 请确保文件已上传 |
| -60012 | 找不到任务 | 请确保task_id有效且未删除 |
| -60013 | 没有权限访问该任务 | 只能访问自己提交的任务 |
| -60014 | 删除运行中的任务 | 运行中的任务暂不支持删除 |
| -60015 | 文件转换失败 | 可以手动转为pdf再上传 |
| -60016 | 文件转换失败 | 文件转换为指定格式失败，可以尝试其他格式导出或重试 |
