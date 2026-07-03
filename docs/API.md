# API

## GET /healthz

健康检查。

响应：

```json
{
  "status": "ok"
}
```

## GET /

眼镜 HUD Demo。

页面会调用 `/api/vision/recognize`，并把返回的 `box`、`letter`、`english`、`chinese`、`display_text` 叠加到视野里。默认使用模拟场景；允许浏览器摄像头权限后，会截取真实摄像头当前帧作为 `image_base64` 上传。

HUD 还支持：

- 自动识别：定时截帧并刷新画面标签。
- 本地缓存：成功识别后把最近结果保存到浏览器 `localStorage`。
- 离线兜底：请求失败时恢复最近一次识别结果，并显示 `Offline cache`。

## POST /api/vision/recognize

识别眼镜上传的图片帧。

请求：

```json
{
  "device_id": "glass_001",
  "frame_id": "f_123",
  "image_base64": "...",
  "last_scene_hash": "optional",
  "offline_ok": true
}
```

响应：

```json
{
  "scene_hash": "a1b2c3d4",
  "from_cache": false,
  "objects": [
    {
      "letter": "A",
      "name": "cup",
      "english": "cup",
      "meaning": "杯子",
      "chinese": "杯子",
      "phonetic": "/kʌp/",
      "sentence": "This is a cup.",
      "display_text": "A cup / 杯子",
      "speak_text": "A. cup. This is a cup.",
      "box": {
        "x": 120,
        "y": 80,
        "width": 90,
        "height": 110
      },
      "score": 0.92,
      "learning": {
        "english": "cup",
        "chinese": "杯子",
        "phonetic": "/kʌp/",
        "example_sentence": "This is a cup.",
        "example_meaning": "这是一个杯子。"
      }
    }
  ]
}
```

## 字段说明

- `letter`：画面标签，给眼镜 HUD 使用。
- `name` / `english`：英文单词。`name` 保留给旧客户端兼容，推荐新客户端使用 `english`。
- `meaning` / `chinese`：中文解释。`meaning` 保留给旧客户端兼容，推荐新客户端使用 `chinese`。
- `phonetic`：音标。
- `sentence`：学习例句。
- `display_text`：眼镜 HUD 可直接展示的中英文字，例如 `A cup / 杯子`。
- `speak_text`：TTS 可直接朗读的英文学习文本。
- `learning`：完整学习卡片，包含英文、中文、音标、例句和例句中文。
- `box`：目标在画面中的位置。
- `score`：识别置信度。
- `scene_hash`：场景缓存键。
- `from_cache`：结果是否来自缓存。

## Cloud Vision Provider Contract

当服务端配置为：

```env
VISION_PROVIDER=cloud
CLOUD_VISION_ENDPOINT=https://example.com/vision/recognize
```

后端会把眼镜帧转发到 `CLOUD_VISION_ENDPOINT`。

请求：

```json
{
  "device_id": "glass_001",
  "frame_id": "f_123",
  "image_base64": "data:image/jpeg;base64,...",
  "last_scene_hash": "optional"
}
```

响应：

```json
{
  "scene_hash": "scene_001",
  "objects": [
    {
      "english": "chair",
      "box": {
        "x": 10,
        "y": 20,
        "width": 80,
        "height": 90
      },
      "score": 0.91
    }
  ]
}
```

兼容字段：

- `english`：推荐字段，英文目标名。
- `name`：兼容字段，如果没有 `english`，服务端会读取 `name`。
