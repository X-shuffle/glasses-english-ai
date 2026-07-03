# Glasses English AI

面向智能眼镜的英语学习视觉识别服务。目标是让眼镜快速识别画面中的物体或图形，为每个目标打上英文字母标签，并返回英文单词、中文释义、音标和例句。

## 当前状态

这是项目初始化版本，包含：

- Go HTTP 服务端
- `/healthz` 健康检查接口
- `/api/vision/recognize` 识别接口
- 内存场景缓存
- Mock 视觉识别器
- API、架构、路线图文档
- Git 初始化准备

当前版本不依赖外部 Go 包，方便先跑通服务结构。

## 快速开始

```bash
cp .env.example .env
make test
make run
```

服务默认监听：

```text
http://localhost:8080
```

测试识别接口：

```bash
curl -X POST http://localhost:8080/api/vision/recognize \
  -H 'Content-Type: application/json' \
  -d '{"device_id":"glass_001","frame_id":"f_1","image_base64":"demo"}'
```

## 核心能力规划

1. 眼镜本地抽帧并上传关键帧。
2. 服务端识别画面中的物体、图形和文字。
3. 对每个目标分配 `A/B/C` 等字母标签。
4. 返回英语学习内容：单词、中文意思、音标、例句。
5. 本地缓存相似场景，眼镜端快速响应。
6. 网络断开时使用本地缓存和轻量模型兜底。

## 推荐开发顺序

1. 替换 Mock 识别器，接入云端多模态视觉 API。
2. 增加图片帧存储和 scene hash 去重。
3. 增加 SQLite 或 PostgreSQL 保存学习记录。
4. 增加眼镜端 WebSocket 长连接。
5. 增加区域变化检测，只识别变化区域。
6. 积累数据后训练专用 YOLO/ONNX 模型。

## 目录结构

```text
cmd/server             服务入口
internal/config        环境变量配置
internal/httpapi       HTTP API
internal/cache         场景缓存
internal/vision        识别接口和 Mock 实现
docs                   架构、API、路线图
```
