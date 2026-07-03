# API

## GET /healthz

健康检查。

响应：

```json
{
  "status": "ok"
}
```

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
      "meaning": "杯子",
      "phonetic": "/kʌp/",
      "sentence": "This is a cup.",
      "box": {
        "x": 120,
        "y": 80,
        "width": 90,
        "height": 110
      },
      "score": 0.92
    }
  ]
}
```

## 字段说明

- `letter`：画面标签，给眼镜 HUD 使用。
- `name`：英文单词。
- `meaning`：中文解释。
- `phonetic`：音标。
- `sentence`：学习例句。
- `box`：目标在画面中的位置。
- `score`：识别置信度。
- `scene_hash`：场景缓存键。
- `from_cache`：结果是否来自缓存。
