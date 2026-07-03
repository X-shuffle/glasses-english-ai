# Glasses English AI

面向智能眼镜的英语学习视觉识别服务。系统从眼镜视频帧中识别物体、图形和文字，为每个目标分配 `A/B/C` 等英文标签，并返回英文单词、中文释义、音标、例句和画面位置，让眼镜可以快速显示和朗读学习内容。

项目会按照 DDD（Domain-Driven Design，领域驱动设计）演进：业务规则沉在领域层，HTTP、缓存、云 API、数据库和模型推理都作为基础设施适配器接入，避免核心识别学习逻辑被外部技术细节绑死。

## 当前状态

DDD 初版已经包含：

- Go HTTP 服务端。
- `/healthz` 健康检查接口。
- `/api/vision/recognize` 识别接口。
- `RecognizeFrame` 应用用例。
- `SceneRecognition`、`VisualObject`、`LearningCard` 领域模型。
- 内存场景仓储，用于相似场景快速返回。
- Mock 视觉识别器，用于模拟眼镜视野里的多个物体。
- 静态中英词典，用于返回中文释义、音标和例句。
- API、架构、路线图文档。

当前代码暂时保持轻量，没有引入外部 Go 包，方便先跑通“识别目标 -> 分配标签 -> 中英学习内容 -> 缓存复用”的完整闭环。

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
  -d '{"device_id":"glass_001","frame_id":"f_1","image_base64":"desk demo"}'
```

返回结果会包含类似：

```json
{
  "letter": "A",
  "english": "cup",
  "chinese": "杯子",
  "display_text": "A cup / 杯子",
  "speak_text": "A. cup. This is a cup."
}
```

## DDD 分层

项目当前按下面的方向组织代码：

```text
cmd/server
  服务启动入口，只负责组装依赖和启动 HTTP 服务。

internal/interfaces/httpapi
  用户接口层。放 HTTP handler、WebSocket handler、请求响应 DTO。

internal/application
  应用层。当前实现 RecognizeFrame 用例。

internal/domain
  领域层。放核心业务模型、聚合、仓储接口和外部能力端口。

internal/infrastructure
  基础设施层。当前实现内存缓存、Mock 视觉识别、静态学习词典。

internal/config
  配置读取和运行参数。

docs
  架构、API、路线图和训练方案文档。
```

云视觉 API、ONNX 推理、Redis/SQLite 和文件存储后续都作为 `infrastructure` 下的适配器接入。

## 领域模型

### 核心领域：眼镜英语识别学习

核心问题不是单纯“识别图片”，而是把视觉结果转成可学习、可显示、可缓存、可复习的英语知识。

### 主要实体和值对象

- `Device`：眼镜设备，包含设备 ID、在线状态、能力信息。
- `Frame`：视频帧，包含帧 ID、图片数据、时间戳、上一场景 hash。
- `Scene`：识别场景，包含 scene hash、画面目标、缓存状态。
- `VisualObject`：画面中的目标，包含标签字母、英文名、置信度和位置。
- `BoundingBox`：目标位置值对象。
- `LearningCard`：英语学习卡片，包含单词、中文释义、音标、例句。
- `RecognitionResult`：一次识别结果，连接场景、目标和学习内容。

### 聚合设计

- `SceneRecognition` 聚合：管理一帧画面识别出的目标、标签分配、结果可信度和缓存键。
- `LearningSession` 聚合：管理用户在一次眼镜使用过程中的学习记录、复习历史和错词。
- `DeviceSession` 聚合：管理眼镜连接状态、最近场景和离线兜底策略。

## 限界上下文

### Vision Context

负责视觉识别和目标定位。

- 输入：视频帧、局部变化区域、历史场景 hash。
- 输出：物体名称、位置、置信度、场景 hash。
- 外部依赖：云视觉 API、自训练模型、ONNX/NCNN 推理。

### Learning Context

负责英语学习内容生成。

- 输入：识别出的英文目标。
- 输出：中文释义、音标、例句、发音文本、学习卡片。
- 外部依赖：词典、LLM、TTS。

### Cache Context

负责快速响应和离线可用。

- 输入：场景 hash、图像 embedding、局部区域 hash。
- 输出：缓存识别结果、离线候选结果。
- 外部依赖：内存缓存、Redis、SQLite、本地文件。

### Device Context

负责眼镜端协作。

- 输入：设备状态、网络状态、帧上传策略。
- 输出：识别响应、预加载任务、离线策略。
- 外部依赖：WebSocket、设备 SDK、HUD 显示层。

## 应用用例

1. `RecognizeFrame`
   接收眼镜上传的图片帧，优先查缓存，未命中时调用视觉识别，再生成英语学习结果。

2. `GetCachedScene`
   根据 `last_scene_hash` 或图像相似度复用历史结果，让眼镜本地快速响应。

3. `RecognizeChangedRegion`
   只识别画面变化区域，例如手伸进来或局部物体移动。

4. `PrepareNextScene`
   根据摄像头缓慢移动方向预测下一帧可能出现的区域，提前请求云 API。

5. `StudyRecognizedObject`
   用户看向某个标签时，返回该目标的发音、解释、例句和复习记录。

## 依赖方向

DDD 代码依赖方向应该保持为：

```text
interfaces -> application -> domain
infrastructure -> domain
application -> domain
```

领域层不依赖 HTTP、数据库、Redis、云 API 或具体模型 SDK。外部能力通过接口注入，例如：

- `SceneRepository`
- `RecognitionProvider`
- `LearningContentProvider`
- `FrameStorage`
- `DeviceNotifier`

这样未来从 Mock 识别器切换到云 API、自训练模型或眼镜本地模型时，不需要重写核心业务规则。

## 核心能力规划

1. 眼镜本地抽帧并上传关键帧。
2. 服务端识别画面中的物体、图形和文字。
3. 对每个目标分配 `A/B/C` 等字母标签。
4. 返回英语学习内容：单词、中文意思、音标、例句。
5. 本地缓存相似场景，眼镜端快速响应。
6. 网络断开时使用本地缓存和轻量模型兜底。

## 推荐开发顺序

1. 接入云端多模态视觉 API，替换 Mock 视觉识别器。
2. 增加图片帧压缩、大小限制和上传限流。
3. 增加 SQLite 或 PostgreSQL 保存学习记录。
4. 增加眼镜端 WebSocket 长连接。
5. 增加区域变化检测，只识别变化区域。
6. 增加本地轻量模型和离线候选结果。
7. 积累数据后训练专用 YOLO/ONNX 模型。

## 当前目录结构

```text
cmd/server             服务入口
internal/config        环境变量配置
internal/interfaces    HTTP API 和请求响应 DTO
internal/application   RecognizeFrame 应用用例
internal/domain        领域模型、仓储接口、外部能力端口
internal/infrastructure/cache      内存场景仓储
internal/infrastructure/learning   静态中英学习词典
internal/infrastructure/vision     Mock 视觉识别器
docs                   架构、API、路线图
```

## 设计原则

- 领域模型先表达业务语言，再考虑技术实现。
- 应用层只编排流程，不写底层 API 细节。
- 基础设施层可以替换，领域层保持稳定。
- 缓存命中时优先返回，云端结果异步修正。
- 离线模式是核心能力，不是附加功能。
