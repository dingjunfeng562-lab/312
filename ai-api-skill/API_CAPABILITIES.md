# 好多米Ai API 能力缓存

> 同步时间：2026-07-11T00:59:42.768Z
> 来源：https://api.lk888.ai/api/v1/skills（使用项目中 haoduomi-ai 渠道的现有 API Key）
> API Key 位置：`config/config.yaml` 的 `haoduomi-ai` 渠道配置；本文不保存密钥值。

## 平台能力与接口

```json
{
  "auth": {
    "format": "Bearer {api_key}",
    "header": "Authorization",
    "method": "bearer"
  },
  "base_url": "https://api.lk888.ai/api",
  "categories": [
    {
      "endpoints": [
        {
          "description": "按类型查询平台所有可用模型。返回每个模型的名称、展示名称、类型、功能标签和简介。\n- 不传 type 参数返回所有类型的模型\n- type=chat 时只返回 gpt/o1/o3/chatgpt/claude/gemini 前缀的语言模型，并额外返回 api_format（调用格式：openai/anthropic/gemini）和 api_endpoint（对应的请求路径）\n- type=image/video/audio 返回对应类型的媒体模型（TTS 语音合成、音乐均归类为 audio）\n\n响应字段说明：\n- name: 模型标识名，调用接口时传此值\n- display_name: 展示用的中文名称\n- type: 模型类型（chat/image/video/audio）\n- tags: 功能标签数组，如[\"文生视频\",\"图生视频\"]\n- description: 模型简介\n- input_hint: 输入提示文案\n- api_format: [仅chat] 调用格式，openai/anthropic/gemini\n- api_endpoint: [仅chat] 对应请求路径，如 /v1/chat/completions",
          "id": "list_models",
          "method": "GET",
          "name": "获取模型列表",
          "params": [
            {
              "description": "按模型类型筛选，不传则返回全部",
              "enum": [
                "chat",
                "image",
                "video",
                "audio"
              ],
              "name": "type",
              "required": false,
              "type": "string"
            }
          ],
          "path": "/v1/skills/models",
          "response_example": {
            "models": [
              {
                "api_endpoint": "/v1/chat/completions",
                "api_format": "openai",
                "available_for_this_key": true,
                "description": "OpenAI旗舰模型",
                "display_name": "GPT-4o",
                "input_hint": "",
                "name": "gpt-4o",
                "tags": [
                  "对话",
                  "多模态"
                ],
                "type": "chat"
              },
              {
                "available_for_this_key": true,
                "description": "高质量AI视频生成",
                "display_name": "Grok Video 3",
                "input_hint": "描述视频内容",
                "name": "grok-video-3",
                "tags": [
                  "文生视频"
                ],
                "type": "video"
              },
              {
                "aliases": [
                  "Seedance",
                  "即梦"
                ],
                "available_for_this_key": true,
                "description": "字节跳动即梦团队推出的旗舰级视频生成模型 Seedance 2.0...",
                "display_name": "SD 2.0 首尾帧",
                "input_hint": "不传图=文生视频，1张图=首帧，2张图=首尾帧",
                "name": "kwvideo-v2",
                "tags": [
                  "文生视频",
                  "首尾帧"
                ],
                "type": "video"
              }
            ],
            "total": 3,
            "type": "video"
          },
          "tips": "1. chat 类型只返回主流语言模型（gpt/o1/o3/chatgpt/claude/gemini 前缀），其他特殊模型不在此列表中。\n2. 媒体模型（image/video/audio）返回全部可用模型。TTS 语音合成和 music 音乐模型统一归类为 audio 类型，使用 type=audio 可查询到。\n3. 每个 chat 模型的 api_format 告诉你该用哪种格式调用：openai 用 /v1/chat/completions，anthropic 用 /v1/messages，gemini 用 /v1beta/models/{model}:{action}。\n4. 要获取模型详细参数用 /v1/skills/models/{name}，要获取价格用 /v1/skills/models/{name}/pricing。\n5. aliases 字段：可选。当模型 display_name 是缩写但 description 含业内通用品牌名（如 Seedance / Veo / Hailuo / Kling 等）时，系统会自动把品牌名提到 aliases 数组里，方便按品牌名检索。display_name 已含的品牌名不会重复列出。例如按 \"seedance\" 检索可命中 display_name=\"SD 2.0 首尾帧\" 但 aliases 包含 \"Seedance\" 的模型。\n6. 响应顶层结构固定为 {\"models\": [...], \"total\": 数量, \"type\": 查询的类型}；不传 type 参数时 type 返回 \"all\"。模型字段（name/display_name/type/tags 等）在 models 数组的每个元素内。\n7. is_active 是平台级状态。若你的 API Key 渠道策略为“自定义”，实际能否创建任务取决于该 Key 的白名单：未在白名单中配置分组的模型调用 /v1/media/generate 会直接返回 403（该模型未在自定义渠道策略中配置可用渠道分组）。请以响应中的 available_for_this_key 字段为准判断当前 Key 可用性。",
          "when_to_use": "用户想知道有哪些模型可用，或需要选择模型时"
        },
        {
          "description": "查询单个模型的功能信息和参数列表，不含价格。\n\n响应字段说明：\n- name/display_name/type/tags/description: 同模型列表接口\n- input_hint: 提示用户输入什么，如\"描述视频内容\"\n- params: 参数定义数组，每个参数包含：\n  - name: 参数标识名，调用时传入 params 对象的 key\n  - label: 参数中文名称\n  - type: 参数类型，select=下拉选择，textarea=文本输入，number=数字输入，upload=文件上传，switch=开关\n  - required: 是否必填\n  - default: 默认值\n  - options: [仅select类型] 可选项数组，每项含 label(显示名)/value(传入值)/is_default\n  - description: 参数说明\n\n调用媒体生成接口时，将此处获取的参数放入请求体的 params 对象中。",
          "id": "model_detail",
          "method": "GET",
          "name": "获取模型功能与参数",
          "params": [
            {
              "description": "模型名称，如 grok-video-3",
              "in": "path",
              "name": "model_name",
              "required": true,
              "type": "string"
            }
          ],
          "path": "/v1/skills/models/{model_name}",
          "response_example": {
            "aliases": [
              "Seedance",
              "即梦"
            ],
            "description": "字节跳动即梦团队推出的旗舰级视频生成模型 Seedance 2.0",
            "display_name": "SD 2.0 首尾帧",
            "input_hint": "不传图=文生视频，1张图=首帧，2张图=首尾帧",
            "name": "kwvideo-v2",
            "params": [
              {
                "description": "描述视频内容",
                "label": "提示词",
                "name": "prompt",
                "required": true,
                "type": "textarea"
              },
              {
                "description": "生成时长",
                "label": "视频时长",
                "name": "duration",
                "options": [
                  {
                    "label": "自动",
                    "value": "auto"
                  },
                  {
                    "label": "5秒",
                    "value": "5"
                  },
                  {
                    "label": "10秒",
                    "value": "10"
                  }
                ],
                "required": true,
                "type": "select"
              }
            ],
            "tags": [
              "文生视频",
              "首尾帧"
            ],
            "type": "video"
          },
          "tips": "1. 此接口只返回功能和参数，不含价格。查价格请用 /v1/skills/models/{name}/pricing。\n2. params 数组定义了调用媒体生成接口时可传的参数。type=select 的参数必须从 options 中选取 value 值，不能自拟。\n3. type=upload 的参数表示需要传入文件（图片/视频/音频）的可公开访问 URL。单个文件可传字符串 \"https://example.com/image.jpg\"，多个文件传数组 [\"https://example.com/1.jpg\",\"https://example.com/2.jpg\"]，两种格式均支持。参数描述中会标注支持的文件数量范围。平台不提供文件上传/托管服务，请自行将文件上传至对象存储服务后传入 URL。\n4. 语言模型（chat类型）也会返回 params（常见如 attachments、web_search、enable_thinking）——它们描述的是该模型支持的能力，不是通过 params 对象调用。chat 模型一律走透传 Chat API（OpenAI /v1/chat/completions、Anthropic /v1/messages、Gemini /v1beta/models/{model}:{action}），这些能力要按对应上游的原生协议在请求体里启用：多模态输入（对应 attachments）按上游格式放进 messages（OpenAI 用 image_url 内容块、Anthropic 用 image source 块、Gemini 用 inline_data/file_data parts）；联网（web_search）、思考（enable_thinking）等同理按各上游原生字段/工具开启，不要把它们当作 media-generate 的 params 传。\n5. aliases 字段：可选，同 list_models 接口。当 display_name 是缩写但 description 含业内通用品牌名（如 Seedance / Veo / Hailuo / Kling 等）时，系统会自动把品牌名提到 aliases 数组里，可以按品牌名检索。display_name 已含的品牌名不会重复列出。\n6. 图片模型的 size（尺寸）代表目标宽高比与分辨率；部分上游渠道（如 gpt-image 系）会按所选宽高比返回近似分辨率（例如选 1920x1088(16:9)，实际可能返回 1666x944，仍是 16:9）。如需精确像素，请在拿到 result_url 后自行裁剪/缩放。",
          "when_to_use": "需要查看模型支持哪些参数、每个参数的选项时"
        },
        {
          "description": "查询模型所有渠道分组的完整价格信息，包括参数价格变动。默认返回全量渠道分组，每个分组含 is_active 字段标识当前是否启用。注意：is_active=false 的分组并非永久关闭，平台会根据供应商状态随时启用或关闭渠道分组，因此展示价格时应包含所有分组供用户参考。传 ?status=active 可仅获取当前正在运行的分组。",
          "id": "model_pricing",
          "method": "GET",
          "name": "获取模型完整价格",
          "params": [
            {
              "description": "模型名称",
              "in": "path",
              "name": "model_name",
              "required": true,
              "type": "string"
            },
            {
              "description": "筛选条件。不传或为空返回全部分组；传 active 仅返回当前启用的分组",
              "in": "query",
              "name": "status",
              "required": false,
              "type": "string"
            }
          ],
          "path": "/v1/skills/models/{model_name}/pricing",
          "response_example": {
            "available_for_this_key": true,
            "channel_groups": [
              {
                "avg_response_seconds": 42.3,
                "base_price": 1.5,
                "billing_method": "按次",
                "group_name": "标准渠道",
                "in_key_whitelist": true,
                "input_token_price": 0,
                "is_active": true,
                "option_prices": [
                  {
                    "final_price": 1.5,
                    "option_label": "5秒",
                    "option_value": "5",
                    "param_name": "duration",
                    "price_addition": 0,
                    "price_impact": "基础价格",
                    "price_multiplier": 1
                  },
                  {
                    "final_price": 3,
                    "option_label": "10秒",
                    "option_value": "10",
                    "param_name": "duration",
                    "price_addition": 0,
                    "price_impact": "x2",
                    "price_multiplier": 2
                  }
                ],
                "output_token_price": 0,
                "success_rate_24h": 95.5
              },
              {
                "avg_response_seconds": 0,
                "base_price": 2,
                "billing_method": "按次",
                "group_name": "高速渠道",
                "in_key_whitelist": true,
                "input_token_price": 0,
                "is_active": false,
                "option_prices": [],
                "output_token_price": 0,
                "success_rate_24h": 0
              }
            ],
            "display_name": "Grok Video 3",
            "filter": "",
            "key_channel_strategy": "综合最优",
            "name": "grok-video-3",
            "pricing_note": "默认返回所有渠道分组（含已关闭的），加 ?status=active 仅返回当前启用的分组。实际调用时不需要指定渠道分组，系统根据 API Key 的渠道策略自动选择。若某参数选项未出现在 option_prices 中，表示该选项使用分组的基础价格（base_price），无额外加价。",
            "type": "video"
          },
          "tips": "1. 默认返回所有渠道分组，每个分组有 is_active 字段。is_active=false 表示该分组当前暂停服务，但随时可能重新启用，不代表永久下线。\n2. 建议向终端用户展示全部分组价格（含暂停的），因为这些分组可能随时恢复。\n3. 传 ?status=active 仅返回当前正在运行的分组，适用于只关心实时可用渠道的场景。\n4. 只返回分组名称，不暴露上游供应商信息。价格已含代理商加成。\n5. 若某参数选项未出现在分组的 option_prices 中，表示该选项使用分组的基础价格（base_price），无额外加价。\n6. is_active 是平台级状态。若你的 API Key 渠道策略为“自定义”，实际能否创建任务取决于该 Key 的白名单：未在白名单中配置分组的模型调用 /v1/media/generate 会直接返回 403（该模型未在自定义渠道策略中配置可用渠道分组）。请以响应中的 available_for_this_key 字段为准判断当前 Key 可用性。",
          "when_to_use": "用户需要查看模型定价、比较不同渠道价格时"
        }
      ],
      "id": "models",
      "name": "模型查询"
    },
    {
      "endpoints": [
        {
          "description": "返回平台所有模型的通用调用指南，包含以下内容：\n\n1. 语言模型三种调用格式：\n   - OpenAI 格式：POST /v1/chat/completions，适用于 gpt/o1/o3/chatgpt 前缀模型\n   - Anthropic 格式：POST /v1/messages，适用于 claude 前缀模型\n   - Gemini 格式：POST /v1beta/models/{model}:{action}，适用于 gemini 前缀模型\n   每种格式含请求示例和响应示例\n\n2. 媒体模型异步轮询流程：\n   - 第一步：POST /v1/media/generate 提交任务，获取 task_id\n   - 第二步：GET /v1/skills/task-status?task_id=xxx 轮询状态\n   - 轮询间隔建议5秒，is_final=true 时停止\n\n3. 价格计算公式：\n   - 按次计费：最终价格 = 基础价格 × 参数系数 + 参数加价\n   - 按token计费：费用 = 输入token数 × 输入单价 + 输出token数 × 输出单价\n\n4. 渠道策略说明：\n   - 价格优先：自动选择最便宜的可用渠道\n   - 速度优先：自动选择响应最快的可用渠道\n   - 成功率优先：自动选择成功率最高的可用渠道\n   策略在用户的 API Key 设置中配置，调用时无需指定",
          "id": "guide",
          "method": "GET",
          "name": "通用调用说明",
          "path": "/v1/skills/guide",
          "tips": "1. 此接口返回的是通用调用指南，所有模型共用，不是某个具体模型的说明。\n2. 语言模型支持流式输出（stream:true），媒体模型只支持异步轮询，不支持流式。\n3. 渠道策略由用户在平台网站的 API Key 设置中配置，调用接口时无需且不能指定渠道或策略。\n4. 所有接口均需在 Header 中携带 Authorization: Bearer {API_KEY} 进行认证。",
          "when_to_use": "需要了解如何调用模型、选择哪个端点时"
        }
      ],
      "id": "calling",
      "name": "调用说明"
    },
    {
      "endpoints": [
        {
          "description": "查询媒体生成任务的实时状态。提交生成任务后，通过此接口轮询任务进度和结果。\n\n响应字段说明：\n- task_id: 任务ID\n- model: 使用的模型名称\n- status: 任务状态文本，如\"排队中\"\"生成中\"\"生成完成\"\"生成失败\"\n- status_group: 状态分组，\"等待中\"/\"处理中\"/\"已完成\"/\"失败\"\n- progress: 进度百分比，如\"0%\"、50%\"、\"100%\"\n- is_final: 是否为终态。true 表示任务已结束（成功或失败），必须停止轮询\n- result_url: 生成结果的下载地址，仅成功时有值\n- result_type: 结果类型，video/image/audio 等\n- cost: 实际扣费的算力值\n- channel_group: 实际使用的渠道分组名称\n- error: 失败时的错误信息\n- created_at: 任务创建时间\n- completed_at: 任务完成时间，未完成时为空\n- duration_seconds: 从创建到完成的耗时（秒）",
          "id": "task_status",
          "method": "GET",
          "name": "查询任务状态",
          "params": [
            {
              "description": "任务ID",
              "name": "task_id",
              "required": true,
              "type": "integer"
            }
          ],
          "path": "/v1/skills/task-status",
          "response_example": {
            "channel_group": "标准渠道",
            "completed_at": "2026-03-17T10:02:13Z",
            "cost": 1.5,
            "created_at": "2026-03-17T10:00:00Z",
            "duration_seconds": 133,
            "error": null,
            "input_files": [
              "https://example.com/ref.jpg"
            ],
            "is_final": true,
            "model": "grok-video-3",
            "progress": "100%",
            "refunded": false,
            "refunded_amount": 0,
            "result_type": "video",
            "result_url": "https://cdn.example.com/video/abc.mp4",
            "state": "success",
            "status": "生成完成",
            "status_group": "已完成",
            "task_id": 12345
          },
          "tips": "1. 【重要】判断任务是否完成只看 state 和 is_final，不要看 progress 数值。很多异步模型（如 nano-banana-2 / DM 通道、多数视频模型）上游不返回中间 progress，progress 会全程保持 \"0\"，仅在完成瞬间跳 \"100\"——这不是卡住，不要提前重试。\n2. 当 is_final=true 时必须停止轮询，不要继续请求。\n3. 推荐轮询节奏：提交后先等 5-10 秒再开始首次轮询（过早轮询会看到 state=pending、progress=0 是正常现象），之后每 5 秒轮询一次，不要太频繁。\n4. 如果轮询超过 7200 秒（7200秒=2小时）任务仍未完成，可认为超时，停止轮询并提示用户。\n5. cost 字段语义：代表本次任务的实际成本（单位：算力）。成功任务=运行中按实扣金额；失败任务平台会自动全额退款，此时 cost=0 且 refunded=true；pending 未运行时 cost=0。不需要再关联消费记录表才能知道是否已退款。\n6. refunded 字段：false=未退款或未发生退款（含成功/运行中/异常仅扣费未退款场景）；true=任务失败且平台已自动退回全额。可用来向用户明确交代“这笔任务失败了但钱已退”。\n7. refunded_amount 字段：refunded=true 时为实际退还金额（算力）；其他场景为 0。\n8. channel_group 返回实际使用的渠道分组名称，可用于账单展示和成本记录。\n9. 此接口是 /v1/media/status 的增强版，额外返回 model、created_at、completed_at、duration_seconds、channel_group、refunded、refunded_amount 字段。\n10. error 字段：无错误时为 null，有错误时为字符串。判断任务是否失败请检查 state==\"failed\"（或 status_group==\"失败\"），不要只靠 error != null。\n11. input_files 字段：始终为数组类型，无输入文件时为空数组 []。【名称重叠提醒】input_files 只是本响应的统一字段名，调 /v1/media/generate 提交任务时，上传参数名不是 input_files，而是模型详情里实际的 upload 参数名（如 images / image_url / videos / attachments 等），请通过 /v1/skills/models/{name} 查看。\n12. 【重要】duration_seconds 是【任务处理总耗时】（completed_at - created_at 的墙钟秒数，含排队/上游生成/下载落库），不是输出视频/音频本身的时长。输出媒体本身的时长由提交时 params.duration 决定（需查对应任务的提交参数），本接口不返回媒体本身时长。",
          "when_to_use": "提交媒体生成任务后需要轮询结果时"
        }
      ],
      "id": "task",
      "name": "任务管理"
    },
    {
      "endpoints": [
        {
          "description": "查询当前 API Key 对应用户的算力余额和 Key 额度使用情况。\n\n响应字段说明：\n- balance: 用户账户的算力余额（注意：单位是算力，不是人民币）\n- unit: 余额单位，固定为\"算力\"\n- api_key_quota: API Key 的额度信息\n  - used: 该 Key 已使用的算力\n  - limit: 该 Key 的总额度上限，0 表示不限额\n  - remaining: 该 Key 剩余可用额度，仅在 limit>0 时返回",
          "id": "balance",
          "method": "GET",
          "name": "查询算力余额",
          "path": "/v1/skills/balance",
          "response_example": {
            "api_key_quota": {
              "limit": 1000,
              "remaining": 849.7,
              "used": 150.3
            },
            "balance": 128.5,
            "unit": "算力"
          },
          "tips": "1. 余额单位是算力，不是人民币。展示时用\"算力\"而不是\"元\"。\n2. api_key_quota.limit=0 表示该 Key 不限额，此时不会返回 remaining 字段。\n3. 余额不足时应提示用户前往平台官网充值算力。\n4. 建议在调用付费接口前先查询余额，避免因余额不足导致任务失败。",
          "when_to_use": "调用付费接口前检查余额是否充足时"
        },
        {
          "description": "查询最近 N 天本 API Key（或跨 Key 按账户）的算力消费。默认返回按模型聚合的汇总（调用次数/成功数/失败数/实际扣费/退款金额）；detail=1 返回按任务倒序的最近消费记录。失败但已退款的任务 cost 在响应里会处理为 0，refunded_amount 会单独给出，与 /v1/skills/task-status 一致。",
          "id": "usage",
          "method": "GET",
          "name": "查询消费明细",
          "params": [
            {
              "description": "范围：key 仅本次调用使用的 API Key（默认）；user 同账户下所有 API Key 汇总。",
              "enum": [
                "key",
                "user"
              ],
              "in": "query",
              "name": "scope",
              "required": false,
              "type": "string"
            },
            {
              "description": "以当前时间为终点向前推 N 天（默认 1，最大 30）；start_time 和 end_time 同时传时本参数被忽略。",
              "in": "query",
              "name": "days",
              "required": false,
              "type": "integer"
            },
            {
              "description": "起始时间。支持 RFC3339、YYYY-MM-DD HH:MM:SS、YYYY-MM-DD；与 end_time 必须同时传。",
              "in": "query",
              "name": "start_time",
              "required": false,
              "type": "string"
            },
            {
              "description": "结束时间。格式同 start_time；范围不得超过 30 天。",
              "in": "query",
              "name": "end_time",
              "required": false,
              "type": "string"
            },
            {
              "description": "过滤到某个模型（传模型名称，取 /v1/skills/models 中的 name 字段）。",
              "in": "query",
              "name": "model",
              "required": false,
              "type": "string"
            },
            {
              "description": "detail=1 返回 records 记录列表（含 task_id），不传或 0 返回按模型聚合的 by_model 数组。",
              "enum": [
                "0",
                "1"
              ],
              "in": "query",
              "name": "detail",
              "required": false,
              "type": "string"
            },
            {
              "description": "仅 detail=1 生效；默认 50，最大 200。",
              "in": "query",
              "name": "limit",
              "required": false,
              "type": "integer"
            },
            {
              "description": "仅 detail=1 生效；从 0 起。",
              "in": "query",
              "name": "offset",
              "required": false,
              "type": "integer"
            }
          ],
          "path": "/v1/skills/usage",
          "request_example": null,
          "response_example": {
            "by_model": [
              {
                "cost": 12.6304,
                "count": 42,
                "failed_count": 2,
                "model": "gpt-image-2",
                "model_type": "image",
                "refunded_cost": 0.6315,
                "refunded_count": 2,
                "success_count": 40
              },
              {
                "cost": 2.8984,
                "count": 8,
                "failed_count": 0,
                "model": "grok-video-3",
                "model_type": "video",
                "refunded_cost": 0,
                "refunded_count": 0,
                "success_count": 8
              }
            ],
            "from": "2026-05-07T17:00:00+08:00",
            "grand_total": {
              "cost": 15.5288,
              "count": 50,
              "failed_count": 2,
              "refunded_cost": 0.6315,
              "refunded_count": 2,
              "success_count": 48
            },
            "scope": "key",
            "to": "2026-05-08T17:00:00+08:00",
            "unit": "算力"
          },
          "tips": "1) 默认 scope=key 只看本次调用使用的 API Key；scope=user 跨 Key 汇总。\n2) 默认 days=1（近 24h），最多 30 天；如需更精确使用 start_time/end_time（同时传），两者冲突时以 start_time/end_time 为准。\n3) 时间格式支持 RFC3339 / YYYY-MM-DD HH:MM:SS / YYYY-MM-DD；不传时区默认服务器本地时区。\n4) ?model=xxx 可过滤到某个模型。\n5) detail=1 时返回的财场变为 records 数组，每条含：task_id, model, model_type, state(success/failed/running/pending), cost, refunded, refunded_amount, created_at, completed_at；顶层额外带 total/limit/offset 翻页字段。limit 默认 50，最大 200；offset 从 0 起。\n6) 该接口不读实时并发资源，只提供历史消费汇总；实时余额调 /v1/skills/balance，实时任务状态调 /v1/skills/task-status。\n7) 按任务表聚合，后台以提交时间为范围过滤；正在排队/生成中的任务也计入，成功/失败后 cost 会随实际结算金额变动。\n8) 失败但已退款的任务，cost 在详情响应里会被处理为 0，refunded=true 且 refunded_amount 为实际退款算力；汇总中则在 refunded_count/refunded_cost 中汇总，by_model[i].cost 为净额（已减退款）。",
          "when_to_use": "定期查加费走势、对账、揭示哪个模型烧钱最多、给用户展示近期产生的调用记录。"
        }
      ],
      "id": "account",
      "name": "账户信息"
    },
    {
      "endpoints": [
        {
          "description": "AI 在使用平台接口过程中，遇到以下情况应主动提交反馈：1) 接口调用报错或返回异常，提交错误信息和调用参数便于平台定位修复；2) 文档说明不清晰或与实际行为不一致，帮助平台完善文档；3) 功能建议或缺少的能力，供平台开发团队参考。这是平台与AI之间的沟通桥梁，积极提交有助于平台持续改进。",
          "id": "feedback",
          "method": "POST",
          "name": "意见反馈与BUG提交",
          "params": [
            {
              "description": "反馈类型",
              "enum": [
                "文档疑问",
                "接口报错",
                "功能建议"
              ],
              "name": "type",
              "required": true,
              "type": "string"
            },
            {
              "description": "问题内容",
              "name": "question",
              "required": true,
              "type": "string"
            },
            {
              "description": "相关接口路径",
              "name": "endpoint",
              "required": false,
              "type": "string"
            },
            {
              "description": "当时操作背景",
              "name": "context",
              "required": false,
              "type": "string"
            },
            {
              "description": "AI工具名称",
              "name": "ai_tool",
              "required": false,
              "type": "string"
            }
          ],
          "path": "/v1/skills/feedback",
          "request_example": {
            "ai_tool": "cursor",
            "context": "用户要接入图生视频",
            "endpoint": "/v1/skills/models/grok-video-3",
            "question": "模型详情中 type=upload 的参数，文件大小限制是多少？",
            "type": "文档疑问"
          },
          "response_example": {
            "feedback_id": 42,
            "message": "反馈已收到，感谢！",
            "success": true
          },
          "tips": "每个API Key每小时最多10条，超过返回429。遇到报错时应尽量提供完整信息：调用的接口路径、请求参数、错误信息、操作步骤。不要因为“不确定是不是BUG”就不提交，平台开发团队会判断处理。",
          "when_to_use": "接口调用报错时主动上报BUG；文档的描述与实际行为不符时反馈；参数格式或返回结果看不懂时提问；觉得缺少某个能力时建议"
        },
        {
          "description": "通过反馈ID查询之前提交的反馈的处理状态和结果",
          "id": "feedback_query",
          "method": "GET",
          "name": "查询反馈处理结果",
          "params": [
            {
              "description": "反馈ID，提交反馈时返回的 feedback_id",
              "name": "id",
              "required": true,
              "type": "integer"
            }
          ],
          "path": "/v1/skills/feedback",
          "response_example": {
            "created_at": "2026-03-25T02:22:50+08:00",
            "endpoint": "/v1/media/generate",
            "feedback_id": 24,
            "question": "问题描述...",
            "resolution": "经排查非代码Bug，图片已正确传递至模型",
            "status": "已处理",
            "type": "接口报错",
            "updated_at": "2026-03-25T15:00:00+08:00"
          },
          "tips": "1. 参数 id 为之前提交反馈时返回的 feedback_id。\n2. 只能查询自己提交的反馈，API Key 不匹配会返回 403。\n3. status 可能的值：未处理、已处理、已忽略。\n4. resolution 字段在状态为“已处理”时包含修复说明。",
          "when_to_use": "提交反馈后想知道处理进度，或者之前反馈过问题想知道是否已修复"
        }
      ],
      "id": "feedback",
      "name": "反馈"
    },
    {
      "endpoints": [
        {
          "description": "完全兼容 OpenAI Chat Completions API。可直接使用 OpenAI 官方 SDK，只需将 base_url 指向本平台即可。\n\n适用模型：gpt/o1/o3/chatgpt 前缀的所有模型\n\n主要参数：\n- model: 模型名称（必填）\n- messages: 消息数组，每条含 role(system/user/assistant) 和 content（必填）\n- stream: 是否流式输出，true 为 SSE 流式，false 为一次性返回（默认false）\n- temperature: 温度参数 0-2（可选）\n- max_tokens: 最大输出 token 数（可选）\n\n响应字段：\n- choices[0].message.content: AI 回复内容\n- usage.prompt_tokens: 输入消耗的 token 数\n- usage.completion_tokens: 输出消耗的 token 数",
          "id": "openai_chat",
          "method": "POST",
          "name": "OpenAI Chat Completions",
          "path": "/v1/chat/completions",
          "request_example": {
            "messages": [
              {
                "content": "你好",
                "role": "user"
              }
            ],
            "model": "gpt-4o",
            "stream": false
          },
          "response_example": {
            "choices": [
              {
                "finish_reason": "stop",
                "index": 0,
                "message": {
                  "content": "你好！有什么可以帮助你的？",
                  "role": "assistant"
                }
              }
            ],
            "id": "chatcmpl-xxx",
            "model": "gpt-4o",
            "object": "chat.completion",
            "usage": {
              "completion_tokens": 12,
              "prompt_tokens": 8,
              "total_tokens": 20
            }
          },
          "tips": "1. 认证方式：Header 中携带 Authorization: Bearer {API_KEY}。\n2. stream=true 时返回 SSE 流，每行格式为 data: {json}，最后一行为 data: [DONE]。\n3. 若使用 OpenAI SDK，只需设置 base_url 和 api_key，其他代码与官方完全一致。\n4. 报错时返回 {\"error\": {\"message\": \"错误说明\", \"type\": \"错误类型\"}}。\n5. 401 表示 API Key 无效，402 表示余额不足，429 表示请求太频繁。\n6. 本平台不识别请求体里的 channel_group 字段，也不识别请求头里的 X-Channel-Group；分组路由完全由 API Key 的渠道策略（价格优先/速度优先/成功率优先）决定，见 /v1/skills/guide 的 channel_strategy。自带这些字段只会被透传给上游（可能触发上游自己的行为），不会改变本平台的路由结果。若需在同一程序里对同一模型做分组 fallback（例如官方→官转→直连），请为每个目标分组各建一把密钥、各配不同策略，然后在客户端自行切换。",
          "when_to_use": "需要调用 GPT 系列模型时"
        },
        {
          "description": "兼容 OpenAI 新版 Responses API 格式。相比 Chat Completions 更简洁，input 可直接传字符串。\n\n适用模型：gpt/o1/o3/chatgpt 前缀的所有模型\n\n主要参数：\n- model: 模型名称（必填）\n- input: 输入内容，可以是字符串或消息数组（必填）\n- stream: 是否流式输出（可选）\n\n响应字段：\n- output[0].content[0].text: AI 回复内容\n- usage.input_tokens: 输入 token 数\n- usage.output_tokens: 输出 token 数",
          "id": "openai_responses",
          "method": "POST",
          "name": "OpenAI Responses",
          "path": "/v1/responses",
          "request_example": {
            "input": "你好",
            "model": "gpt-4o"
          },
          "response_example": {
            "id": "resp-xxx",
            "model": "gpt-4o",
            "object": "response",
            "output": [
              {
                "content": [
                  {
                    "text": "你好！",
                    "type": "output_text"
                  }
                ],
                "type": "message"
              }
            ],
            "usage": {
              "input_tokens": 5,
              "output_tokens": 8
            }
          },
          "tips": "1. 认证方式：Header 中携带 Authorization: Bearer {API_KEY}。\n2. input 可以直接传字符串（简单场景）或消息数组（多轮对话）。\n3. 与 Chat Completions 用同样的模型，二者选其一即可。\n4. 报错格式和状态码与 Chat Completions 一致。\n6. 本平台不识别请求体里的 channel_group 字段，也不识别请求头里的 X-Channel-Group；分组路由完全由 API Key 的渠道策略（价格优先/速度优先/成功率优先）决定，见 /v1/skills/guide 的 channel_strategy。自带这些字段只会被透传给上游（可能触发上游自己的行为），不会改变本平台的路由结果。若需在同一程序里对同一模型做分组 fallback（例如官方→官转→直连），请为每个目标分组各建一把密钥、各配不同策略，然后在客户端自行切换。",
          "when_to_use": "使用 OpenAI 新版 API 格式时（包括GPT-5以上的系列模型）"
        },
        {
          "description": "完全兼容 Anthropic Messages API。可直接使用 Anthropic 官方 SDK，只需将 base_url 指向本平台。\n\n适用模型：claude 前缀的所有模型\n\n主要参数：\n- model: 模型名称（必填）\n- messages: 消息数组，每条含 role(user/assistant) 和 content（必填）\n- max_tokens: 最大输出 token 数（必填，Anthropic 格式强制要求）\n- system: 系统提示词，单独字段而非放在 messages 中（可选）\n- stream: 是否流式输出（可选）\n\n响应字段：\n- content[0].text: AI 回复内容\n- usage.input_tokens: 输入 token 数\n- usage.output_tokens: 输出 token 数\n- stop_reason: 停止原因，\"end_turn\" 表示正常结束",
          "id": "anthropic_messages",
          "method": "POST",
          "name": "Anthropic Messages",
          "path": "/v1/messages",
          "request_example": {
            "max_tokens": 1024,
            "messages": [
              {
                "content": "你好",
                "role": "user"
              }
            ],
            "model": "claude-4-sonnet"
          },
          "response_example": {
            "content": [
              {
                "text": "你好！有什么可以帮助你的？",
                "type": "text"
              }
            ],
            "id": "msg-xxx",
            "model": "claude-4-sonnet",
            "role": "assistant",
            "stop_reason": "end_turn",
            "type": "message",
            "usage": {
              "input_tokens": 10,
              "output_tokens": 15
            }
          },
          "tips": "1. 认证方式：Header 中携带 Authorization: Bearer {API_KEY}（注意：不是 x-api-key，本平台统一使用 Bearer Token）。\n2. max_tokens 是必填参数，不传会报错。建议设为 1024 或更高。\n3. system 提示词是单独的字段，不要放在 messages 数组中，这是 Anthropic 格式与 OpenAI 的主要区别。\n4. 若使用 Anthropic SDK，注意将 base_url 指向本平台，api_key 填写本平台的 API Key。\n5. 报错时返回 {\"error\": {\"message\": \"错误说明\", \"type\": \"错误类型\"}}。\n6. 本平台不识别请求体里的 channel_group 字段，也不识别请求头里的 X-Channel-Group；分组路由完全由 API Key 的渠道策略（价格优先/速度优先/成功率优先）决定，见 /v1/skills/guide 的 channel_strategy。自带这些字段只会被透传给上游（可能触发上游自己的行为），不会改变本平台的路由结果。若需在同一程序里对同一模型做分组 fallback（例如官方→官转→直连），请为每个目标分组各建一把密钥、各配不同策略，然后在客户端自行切换。",
          "when_to_use": "需要调用 Claude 系列模型时"
        },
        {
          "description": "完全兼容 Google Gemini API。可直接使用 Google AI SDK，只需将 base_url 指向本平台。\n\n适用模型：gemini 前缀的所有模型\n\nURL 格式：/v1beta/models/{model}:{action}\n- {model}: 模型名称，如 gemini-3-pro\n- {action}: 操作类型\n  - generateContent: 非流式，一次性返回完整结果\n  - streamGenerateContent: 流式输出\n\n主要参数：\n- contents: 消息数组，每条含 role(user/model) 和 parts（必填）\n  - parts 支持的类型：\n    - {\"text\": \"文本内容\"}: 纯文本\n    - {\"inlineData\": {\"mimeType\": \"类型\", \"data\": \"base64编码\"}}: 图片/视频/音频/PDF 等文件\n- generationConfig: 生成配置，含 temperature/maxOutputTokens 等（可选）\n\n支持的附件类型（通过 inlineData 传入）：\n- 图片：image/jpeg, image/png, image/gif, image/webp\n- 视频：video/mp4, video/webm, video/mov\n- 音频：audio/mp3, audio/wav, audio/ogg, audio/flac\n- 文档：application/pdf\n\n响应字段：\n- candidates[0].content.parts[0].text: AI 回复内容\n- usageMetadata.promptTokenCount: 输入 token 数\n- usageMetadata.candidatesTokenCount: 输出 token 数",
          "id": "gemini_generate",
          "method": "POST",
          "name": "Gemini Generate Content",
          "path": "/v1beta/models/{model}:{action}",
          "request_example": {
            "contents": [
              {
                "parts": [
                  {
                    "text": "请描述这张图片的内容"
                  },
                  {
                    "inlineData": {
                      "data": "/9j/4AAQ...(base64编码的图片数据)",
                      "mimeType": "image/jpeg"
                    }
                  }
                ],
                "role": "user"
              }
            ]
          },
          "response_example": {
            "candidates": [
              {
                "content": {
                  "parts": [
                    {
                      "text": "你好！有什么可以帮助你的？"
                    }
                  ],
                  "role": "model"
                },
                "finishReason": "STOP"
              }
            ],
            "usageMetadata": {
              "candidatesTokenCount": 10,
              "promptTokenCount": 5,
              "totalTokenCount": 15
            }
          },
          "tips": "1. 认证方式：URL 参数 key={API_KEY} 或 Header 中 Authorization: Bearer {API_KEY}，两种方式均支持。\n2. 模型名和操作在 URL 路径中指定，不在请求体中。例如：/v1beta/models/gemini-3-pro:generateContent。\n3. Gemini 的角色名称是 user 和 model（不是 assistant）。\n4. 若使用 Google AI SDK，将 base_url/api_endpoint 指向本平台，api_key 填写本平台的 API Key。\n5. streamGenerateContent 返回的流格式与 Gemini 官方一致。\n6. 传入附件（图片/视频/音频/PDF）时，使用 Gemini 原生的 inlineData 格式：在 parts 数组中添加 {\"inlineData\": {\"mimeType\": \"文件MIME类型\", \"data\": \"base64编码内容\"}}。与 Gemini 官方 API 格式完全一致，无需额外适配。\n7. 附件必须使用 base64 编码内联传入，不支持直接传 URL。若文件在远程服务器，需先下载并转为 base64 后再传入。\n8. 本平台不识别请求体里的 channel_group 字段，也不识别请求头里的 X-Channel-Group；分组路由完全由 API Key 的渠道策略决定，见 /v1/skills/guide 的 channel_strategy。若需分组 fallback，请为不同策略各建一把密钥。",
          "when_to_use": "需要调用 Gemini 系列模型时"
        }
      ],
      "id": "chat_api",
      "name": "语言模型调用"
    },
    {
      "endpoints": [
        {
          "description": "返回所有可用的媒体生成模型及其参数定义。每个模型包含 name、type、label、description 和 params 字段。\n\nparams 定义了调用 /v1/media/generate 时可传的参数，包括名称、类型、选项、默认值等。\n\n注意：建议使用 /v1/skills/models 接口替代，信息更完整（含功能标签、展示名称等）。",
          "id": "media_models",
          "method": "GET",
          "name": "获取媒体模型列表",
          "path": "/v1/media/models",
          "response_example": {
            "code": 200,
            "data": [
              {
                "description": "AI视频生成",
                "label": "高质量视频",
                "name": "grok-video-3",
                "params": [
                  {
                    "default": "5",
                    "label": "时长",
                    "name": "duration",
                    "options": [
                      {
                        "label": "5秒",
                        "value": "5"
                      },
                      {
                        "label": "10秒",
                        "value": "10"
                      }
                    ],
                    "required": true,
                    "type": "select"
                  }
                ],
                "type": "video"
              }
            ]
          },
          "tips": "1. 此接口是早期版本，建议优先使用 /v1/skills/models 获取模型列表。\n2. 两个接口返回的模型数据一致，但 skills 版本额外含 tags、display_name 等字段。",
          "when_to_use": "查看可用的图片/视频/音频生成模型时"
        },
        {
          "description": "提交图片/视频/音频/TTS/音乐生成任务。提交后返回 task_id，通过轮询接口查询结果。\n\n请求体参数：\n- model: 模型名称（必填），从 /v1/skills/models 获取\n- prompt: 提示词/文本描述（必填）\n- params: 参数对象（可选），从 /v1/skills/models/{name} 获取可用参数\n\nparams 用法说明：\n- 先调用 /v1/skills/models/{model_name} 获取模型的 params 定义\n- 将需要的参数组装为对象，key 是参数的 name，value 是参数值\n- 例如模型有 duration 参数（type=select，options含\"5\"和\"10\"），则传 {\"duration\": \"5\"}\n- type=select 的参数必须从 options 中选取 value 值\n- type=upload 的参数传入图片/视频的 URL 地址\n- 未传的参数使用默认值\n\n响应：\n- data.任务id: 任务ID，用于轮询状态",
          "id": "media_generate",
          "method": "POST",
          "name": "提交媒体生成任务",
          "params": [
            {
              "description": "模型名称，通过 /v1/skills/models 获取",
              "name": "model",
              "required": true,
              "type": "string"
            },
            {
              "description": "生成内容的文字描述",
              "name": "prompt",
              "required": true,
              "type": "string"
            },
            {
              "description": "模型专属参数，字段定义来自 /v1/skills/models/{name}。type=upload 的参数传入可公开访问的文件URL",
              "name": "params",
              "required": false,
              "type": "object"
            },
            {
              "description": "已废弃，请勿传：无论传何值，每次请求固定只创建 1 个任务（传入值会被忽略）。需要一次生成多张图/多个结果，请并发发送多个独立请求。此字段仅为兼容老客户端保留",
              "name": "count",
              "required": false,
              "type": "integer"
            }
          ],
          "path": "/v1/media/generate",
          "request_example": {
            "_示例1_文生视频": {
              "model": "viduq3",
              "params": {
                "aspect_ratio": "16:9",
                "duration": "4",
                "model_variant": "turbo",
                "off_peak": "false",
                "resolution": "720p"
              },
              "prompt": "a golden retriever running on the beach at sunset"
            },
            "_示例2_图生图多参考图": {
              "model": "gemini-3.1-flash-image-preview",
              "params": {
                "aspectRatio": "16:9",
                "imageSize": "2K",
                "images": [
                  "https://your-cdn.example.com/ref1.png",
                  "https://your-cdn.example.com/ref2.png"
                ]
              },
              "prompt": "以这些参考图为基础生成一张同风格的主人公走进咖啡馆场景"
            },
            "_示例3_图生视频单参考图": {
              "model": "wan2.6",
              "params": {
                "aspect_ratio": "9:16",
                "duration": "5",
                "img_url": "https://your-cdn.example.com/first-frame.jpg"
              },
              "prompt": "镜头推进，主人公转身微笑"
            }
          },
          "response_example": {
            "code": 200,
            "data": {
              "task_id": 12345,
              "任务ids": [
                12345
              ],
              "对话组ID": "abc123",
              "成功数量": 1
            },
            "msg": "Task created successfully"
          },
          "tips": "1. 认证方式：Header 中携带 Authorization: Bearer {API_KEY}。\n2. 提交后立即返回 task_id（同时为了兼容老版本，仍会返回 任务ids 数组 / 对话组ID 等中文字段），不会等待生成完成。需通过 /v1/skills/task-status?task_id=xxx 轮询结果。\n3. 所有模型特定参数必须放在 params 对象内，不要放在请求体顶层。params 中的参数定义来自 /v1/skills/models/{name} 接口，type=select 的参数只能从 options 中选值，不能自拟。\n4. type=upload 的参数需要传入可公开访问的文件 URL。平台不提供文件上传/托管服务，请自行将文件上传至 COS、CDN 或其他对象存储服务。upload 参数同时支持“单字符串”和“字符串数组”两种形式（如 \"https://x/a.png\" 或 [\"https://x/a.png\",\"https://x/b.png\"]）；禁止传对象数组如 [{\"url\":\"...\"}]，会被拒绝。\n5. 【参数名 vs 响应字段名 并不相同】请求体里的 upload 参数名由模型决定（常见：images / image_url / img_url / videos / video_url / attachments / reference_urls / audio_url 等），task-status 响应里统一叫 input_files 仅作为取回查看，“input_files” 不是请求参数名。在请求里传 input_files 会被静默忽略。\n6. 音乐模型（music-2.5、music-2.5+）在歌曲模式（is_instrumental 不为 instrumental）下，必须在 params 中传入 lyrics（歌词文本），否则会报「歌词不能为空」。\n7. TTS 语音合成模型（如 speech-2.8、doubao-tts-2.0、gemini-2.5-pro-preview-tts）需要先通过 /v1/skills/voices 接口获取可用音色列表，并在 params 中传入对应的 voice_id。speech-2.8 还支持通过 /v1/skills/voices/clone 接口克隆自定义音色。\n8. 提交前建议先查询余额（/v1/skills/balance），余额不足会导致任务失败。\n9. 同一个 API Key 可同时提交多个任务，并行轮询各自的 task_id 即可。",
          "when_to_use": "需要生成图片、视频、音频等媒体内容时"
        },
        {
          "description": "查询媒体生成任务的实时状态（早期版本）。返回任务进度、结果地址、扣费等信息。\n\n建议使用 /v1/skills/task-status 替代，增强版额外返回：\n- model: 模型名称\n- created_at: 创建时间\n- completed_at: 完成时间\n- duration_seconds: 耗时\n- channel_group: 渠道分组名称",
          "id": "media_status",
          "method": "GET",
          "name": "查询任务状态（原版）",
          "params": [
            {
              "description": "任务ID",
              "name": "task_id",
              "required": true,
              "type": "integer"
            }
          ],
          "path": "/v1/media/status",
          "response_example": {
            "code": 200,
            "data": {
              "cost": 1.5,
              "error": "",
              "is_final": true,
              "progress": "100%",
              "result_type": "video",
              "result_url": "https://cdn.example.com/video/abc.mp4",
              "status": "生成完成",
              "task_id": 12345
            }
          },
          "tips": "1. 建议优先使用增强版 /v1/skills/task-status，信息更完整。\n2. 两个接口的基础字段一致（task_id/status/progress/is_final/result_url/cost等）。",
          "when_to_use": "查询任务进度时（建议用 skills 版本）"
        },
        {
          "description": "获取当前用户可用的 TTS 音色列表，支持按模型筛选",
          "id": "voices_list",
          "method": "GET",
          "name": "获取可用音色列表",
          "params": [
            {
              "description": "Filter by model name: speech-2.8 or gemini-2.5-pro-preview-tts. Omit to get all voices.",
              "in": "query",
              "name": "model",
              "required": false,
              "type": "string"
            }
          ],
          "path": "/v1/skills/voices",
          "response_example": {
            "total": 2,
            "voices": [
              {
                "created_at": "2026-04-10T10:00:00+08:00",
                "days_left": 5,
                "demo_audio": "https://example.com/demo.mp3",
                "is_permanent": false,
                "model": "speech-2.8",
                "name": "My Voice",
                "status": "active",
                "type": "cloned",
                "voice_id": "LK_123_1712345678"
              },
              {
                "demo_audio": "https://example.com/kore.mp3",
                "description": "clear and bright",
                "gender": "female",
                "model": "gemini-2.5-pro-preview-tts",
                "name": "Kore",
                "type": "preset",
                "voice_id": "Kore"
              }
            ]
          },
          "tips": "1. 支持 model 参数筛选：?model=speech-2.8 返回用户克隆的音色，?model=gemini-2.5-pro-preview-tts 返回预设音色，不传 model 参数则返回全部。\n2. speech-2.8 的音色为用户自己克隆的音色（type=cloned），需通过 /v1/skills/voices/clone 接口创建。\n3. gemini-2.5-pro-preview-tts 的音色为平台预设音色（type=preset），无需创建。\n4. 克隆音色默认有效期 7 天，首次使用后转为永久。已过期的音色不会在列表中返回。\n5. 调用 /v1/media/generate 时，把列表里的 voice_id 放在请求体 params.voice_id 字段；gemini 系 TTS（gemini-2.5-pro-preview-tts、gemini-2.5-flash-preview-tts）会自动转换为内部 voice_name；minimax（speech-2.8）和 vidu-jieshuoman 直接消费 voice_id。",
          "when_to_use": "调用 TTS 语音合成模型前，先查询可用音色，获取 voice_id 用于生成参数"
        },
        {
          "description": "上传音频文件克隆自定义音色，用于 speech-2.8 模型",
          "id": "voices_clone",
          "method": "POST",
          "name": "克隆自定义音色",
          "path": "/v1/skills/voices/clone",
          "request_example": {
            "audio_url": "https://example.com/my-voice-sample.mp3",
            "name": "My Custom Voice"
          },
          "response_example": {
            "demo_audio": "https://example.com/demo.mp3",
            "expires_at": "2026-04-18T10:00:00+08:00",
            "model": "speech-2.8",
            "name": "My Custom Voice",
            "voice_id": "LK_123_1712345678"
          },
          "tips": "1. 仅支持 speech-2.8 模型的音色克隆。\n2. audio_url 必须是可公开访问的音频文件 URL，支持 mp3/wav/flac 等格式，文件不超过 20MB。\n3. 克隆会扣除 0.1 算力，余额不足时会返回 402 错误。\n4. 音色名称不能重复（同一用户下），不超过 50 个字符。\n5. 新克隆的音色有效期 7 天，首次用于生成后自动转为永久。\n6. 音频建议 10-60 秒，清晰无杂音的单人语音效果最佳。",
          "when_to_use": "需要使用自定义音色进行 TTS 语音合成时，先克隆音色获取 voice_id"
        }
      ],
      "id": "media_api",
      "name": "媒体生成"
    }
  ],
  "platform": "好多米Ai",
  "version": "2026-07-05"
}
```

## 模型调用指南

```json
{
  "call_modes": [
    {
      "applicable_models": "模型名称以 gpt、o1、o3、chatgpt 开头的语言模型",
      "auth": "Authorization: Bearer {api_key}",
      "description": "100% 兼容 OpenAI Chat Completions API，支持流式输出。",
      "endpoints": [
        {
          "method": "POST",
          "name": "Chat Completions",
          "path": "/v1/chat/completions"
        },
        {
          "method": "POST",
          "name": "Responses",
          "path": "/v1/responses"
        }
      ],
      "mode": "realtime_openai",
      "name": "OpenAI 格式（GPT 系列）",
      "request_example": {
        "messages": [
          {
            "content": "你好",
            "role": "user"
          }
        ],
        "model": "gpt-4o",
        "stream": false
      },
      "tips": [
        "设置 stream: true 开启流式输出",
        "兼容 OpenAI SDK，修改 base_url 即可",
        "流式响应以 data: [DONE] 结尾"
      ]
    },
    {
      "applicable_models": "模型名称以 claude 开头的语言模型",
      "auth": "Authorization: Bearer {api_key}",
      "description": "兼容 Anthropic Messages API。",
      "endpoints": [
        {
          "method": "POST",
          "name": "Messages",
          "path": "/v1/messages"
        }
      ],
      "mode": "realtime_anthropic",
      "name": "Anthropic 格式（Claude 系列）",
      "request_example": {
        "max_tokens": 1024,
        "messages": [
          {
            "content": "你好",
            "role": "user"
          }
        ],
        "model": "claude-4-sonnet"
      },
      "tips": [
        "必须传 max_tokens 参数",
        "兼容 Anthropic SDK，修改 base_url 即可",
        "支持 stream: true 流式输出"
      ]
    },
    {
      "applicable_models": "模型名称以 gemini 开头的语言模型",
      "auth": "Authorization: Bearer {api_key}",
      "description": "兼容 Google Gemini API，路径中包含模型名和操作。",
      "endpoints": [
        {
          "method": "POST",
          "name": "Generate Content",
          "path": "/v1beta/models/{model}:generateContent"
        },
        {
          "method": "POST",
          "name": "Stream Generate Content",
          "path": "/v1beta/models/{model}:streamGenerateContent"
        }
      ],
      "mode": "realtime_gemini",
      "name": "Gemini 格式（Gemini 系列）",
      "request_example": {
        "contents": [
          {
            "parts": [
              {
                "text": "你好"
              }
            ],
            "role": "user"
          }
        ]
      },
      "tips": [
        "模型名称在 URL 路径中，不在请求体里",
        "streamGenerateContent 自动添加 ?alt=sse",
        "兼容 Google AI SDK"
      ]
    },
    {
      "applicable_types": [
        "image",
        "video",
        "audio"
      ],
      "applicable_types_note": "audio 涵盖 TTS 语音合成、语音克隆、音乐生成等全部音频类模型；不存在独立的 tts/music 取值。",
      "description": "用于媒体生成，提交任务后轮询获取结果。",
      "file_upload_note": "平台不提供文件上传/托管服务。模型参数中 type=upload 的字段需要传入可公开访问的文件URL，请自行将文件上传至COS、CDN或其他对象存储服务后，将URL作为参数值传入。⚠️ 必须是文件本身的直链（直接 GET 即可下载到文件字节、Content-Type 为 image/* 或 video/* 等），不要传文件分享页/预览页的网页URL（如 filebin.net/xxx、各类网盘分享页），否则上游无法下载该文件，任务会失败并报「媒体类型不支持或链接地址无效」。另外，临时/短效图床（如 tmpfiles.org、catbox.moe/litterbox 等）即使你本地 HEAD 测试返回 200，上游服务端仍可能因地域不可达、反爬或限速而无法下载（报「图像下载失败」），请优先使用 COS/CDN 等稳定对象存储的长期直链。",
      "generate_endpoint": {
        "method": "POST",
        "path": "/v1/media/generate"
      },
      "max_wait_seconds": 1800,
      "max_wait_seconds_note": "1800 秒（30 分钟）只是覆盖绝大多数任务的保守建议值，不是硬死线。判定任务结束请以 is_final=true 为准，不要按墙钟时长一刀切。后端各模型实际轮询兜底差异较大：图片类常见 10 分钟内、视频类常见 25–80 分钟、少数排队型模型（sora2、seedance、vidu_mv 等）兜底可达数小时；只要 is_final=false 就请继续轮询，无需重复提交任务。",
      "mode": "async_poll",
      "name": "异步轮询模式",
      "poll_interval_seconds": 5,
      "status_endpoint": {
        "method": "GET",
        "path": "/v1/skills/task-status?task_id={task_id}"
      },
      "steps": [
        "POST /v1/media/generate 提交任务，返回 task_id",
        "GET /v1/skills/task-status?task_id={task_id} 轮询状态",
        "当 is_final=true 时，从 result_url 获取结果"
      ],
      "tips": [
        "先通过 GET /v1/skills/models/{name} 查看参数",
        "建议轮询间隔 5 秒",
        "判定任务结束以 is_final=true 为准，不要按墙钟时长一刀切超时",
        "视频/复杂图任务常见 5–30 分钟；少数模型（sora2、seedance、pixverse、vidu_mv 等）或排队场景，后端兜底可达 80+ 分钟甚至数小时，请耐心轮询，不要重复提交相同任务",
        "长时间看到 progress=0 不代表任务卡死：部分上游（Gemini Image、Sora 等）在生成期间不推送中间进度，只在最后阶段把 progress 跳到 100",
        "任务真的失败时 state 会变成 failed 或 is_final=true，此时再处理；上游故障已自动退款（refunded=true），无需在客户端做超时退款",
        "cost 字段会随任务推进变化，属正常：任务进行中显示的是提交时锁定的【预扣费/预估】金额，任务成功后会按实际用量多退少补、回写为真实结算金额（两者可能不同）；任务失败退款后 cost=0、refunded=true、refunded_amount 为退回金额。判断是否扣费请以 is_final=true 后的 cost / refunded 为准。",
        "type=upload 的参数需传入可公开访问的文件URL（自行托管至COS/CDN等服务）；务必是文件直链而非分享页/预览页网页URL（如 filebin.net/xxx 这类页面会被上游判为「媒体类型不支持或链接地址无效」）",
        "TTS语音合成模型（如 speech-2.8、gemini-2.5-pro-preview-tts）需先通过 /v1/skills/voices 获取音色列表，speech-2.8 可通过 /v1/skills/voices/clone 克隆自定义音色，在 params 中传入 voice_id 参数",
        "单 Key 无每模型并发上限：同一 API Key 可同时提交多个 /v1/media/generate 任务，并行轮询各自的 task_id；唯一限制是 API Key 的「限额」（如已设置）和账户余额。",
        "task-status 返回的 duration_seconds 是【任务处理总耗时】（提交→完成的墙钟秒数，含排队/上游生成/下载落库），不是输出视频/音频本身的时长。媒体本身的时长由提交时 params.duration 决定。",
        "不想轮询可在提交时传 notify_url，任务终态平台会主动 POST 回调（body 同 task-status，详见 webhook_note）；回调不保证必达，仍建议保留 task_id 轮询兜底。"
      ],
      "webhook_note": "可选：在 POST /v1/media/generate 时传 notify_url（公网可访问的 http/https 地址，禁内网/本机），任务到终态（成功/失败）平台会主动 POST 推送结果到该地址，body 与 GET /v1/skills/task-status 完全一致（用 state/is_final 判定）。成功会等结果转存完成后再推；投递失败按 立即/10秒/30秒/1分钟 重试 4 次，需返回 2xx 视为成功。回调为尽力而为不保证必达：请保留用 task_id 轮询 task-status 的兜底方式，长时间没收到回调就改用轮询。"
    }
  ],
  "channel_strategy": {
    "description": "渠道策略说明，用户在 API Key 设置中选择，调用时自动路由。",
    "is_active_note": "渠道分组的 is_active 字段会随上游可用性实时变化。某次请求实际命中的是任务创建瞬间 is_active=1 且按策略排序靠前的分组；事后再查 /v1/skills/models/{name}/pricing 可能会看到那个分组已经 is_active=false。billed 金额匹配任务命中分组的价格，属正常现象，不代表扣错费。",
    "note": "策略在 API Key 管理页面设置，调用失败时自动切换到下一个渠道重试。",
    "per_request_override": "本平台不支持按单次请求切换渠道分组：请求体里的 channel_group 字段、请求头里的 X-Channel-Group 均不被识别，分组选择完全由 API Key 绑定的策略决定。如果需要在同一个应用里对同一模型做分组 fallback（例如 官方 → 官转 → 直连），建议为每条 fallback 链路各创建一把 API Key、各自配置不同策略，然后在客户端按顺序重试切换，不要期望通过请求参数覆盖路由。",
    "strategies": [
      {
        "description": "综合价格、成功率、速度与实时拥堵四个维度智能分流，并自动避开拥堵渠道（推荐）",
        "name": "综合最优"
      },
      {
        "description": "自动选择价格最低的可用渠道",
        "name": "价格优先"
      },
      {
        "description": "自动选择响应最快的可用渠道",
        "name": "速度优先"
      },
      {
        "description": "自动选择成功率最高的可用渠道",
        "name": "成功率优先"
      },
      {
        "description": "仅在用户为该模型预先勾选的渠道分组中按用户设定的顺序逐个 fallback；若该模型未配置任何分组，将直接返回 403（不会自动切换到价格优先等其它策略）。配置入口：API Key 管理 → 创建/编辑密钥 → 渠道策略=自定义 → 配置渠道。",
        "name": "自定义"
      }
    ]
  },
  "error_format": {
    "description": "所有 /v1/skills/* 和 /v1/media/models 接口失败时返回统一结构，AI 请优先按 error.type 判断错误类别，而不要依赖 error.message 文本",
    "example": {
      "error": {
        "message": "task_id 参数无效",
        "type": "invalid_request_error"
      }
    },
    "schema": {
      "error": {
        "message": "人类可读的错误描述（可能为中文或英文）",
        "type": "错误类别稳定枚举，参见 types 字段"
      }
    },
    "special_note": "注意：/v1/media/generate 接口因历史兼容原因使用 {code,msg,data} 格式，msg 字段直接给出中文描述，AI 请判断 code=200 为成功。聊天透传接口（/v1/chat/completions /v1/messages /v1beta/models/:action /v1/responses）错误格式完全保持上游原始结构不变。",
    "types": [
      {
        "ai_action": "修正参数后重试，直接重试无效",
        "http_status": "400",
        "meaning": "请求参数或入参错误",
        "type": "invalid_request_error"
      },
      {
        "ai_action": "检查 Authorization 头",
        "http_status": "401/403",
        "meaning": "API Key 无效、越权或已禁用",
        "type": "authentication_error"
      },
      {
        "ai_action": "充值或取消调用",
        "http_status": "402",
        "meaning": "账户余额或 API Key 额度不足",
        "type": "insufficient_balance"
      },
      {
        "ai_action": "检查 URL 参数、模型名是否拼错",
        "http_status": "404",
        "meaning": "模型、任务、记录不存在",
        "type": "not_found"
      },
      {
        "ai_action": "退避重试",
        "http_status": "429",
        "meaning": "频率限制",
        "type": "rate_limit_exceeded"
      },
      {
        "ai_action": "5-30 秒后重试；用户余额已自动退回",
        "http_status": "5xx",
        "meaning": "上游供应商临时故障",
        "type": "upstream_error"
      },
      {
        "ai_action": "稍后重试；反复出现请提交 /v1/skills/feedback",
        "http_status": "500",
        "meaning": "平台内部错误",
        "type": "server_error"
      }
    ]
  },
  "not_supported_endpoints": {
    "description": "本平台不提供以下 OpenAI 兼容端点，所有图像/视频/音频/语音生成统一走 POST /v1/media/generate（见上方 async_poll 模式）。例外：POST /v1/images/generations 已支持（仅 model=gpt-image-2 / gpt-image-2-guan，同步返回，支持顶层 image/images 或 params.images 传参考图 URL 走图生图）。",
    "endpoints": [
      {
        "alternative": "POST /v1/media/generate，把底稿与参考图一并放入对应模型的 upload 类参数（如 gpt-image-2 的 params.images，是数组，最多 10 张）",
        "path": "/v1/images/edits"
      },
      {
        "alternative": "POST /v1/media/generate，配合提示词描述变化方向",
        "path": "/v1/images/variations"
      },
      {
        "alternative": "POST /v1/media/generate (model=kling-v3-omni-cankao / wan2.6 / grok-video-3 等视频模型，先调 GET /v1/skills/models?type=video 列出可用模型)",
        "path": "/v1/video/generations"
      },
      {
        "alternative": "POST /v1/media/generate，所有视频生成统一走这个端点；本平台不区分文生视频/图生视频路径，由 model 与 params 决定",
        "path": "/v1/videos/generations"
      },
      {
        "alternative": "POST /v1/media/generate (model=speech-2.8 等 TTS 模型，先调 GET /v1/skills/voices 取音色)",
        "path": "/v1/audio/speech"
      },
      {
        "alternative": "本平台暂不提供语音转文字（ASR）能力",
        "path": "/v1/audio/transcriptions"
      },
      {
        "alternative": "本平台暂不提供语音翻译能力",
        "path": "/v1/audio/translations"
      },
      {
        "alternative": "本平台暂不提供 embedding 能力",
        "path": "/v1/embeddings"
      }
    ],
    "image_edit_note": "本平台没有\"主底稿/参考图\"分离字段。所有参考图统一放在该模型的 upload 类参数里（数组）。如需明确角色/主体来源，请在 prompt 文本中说明（如 \"以第 1 张图的人物为主体，参考第 2 张图的风格\"）。要避免主体漂移，建议：1) 第 1 张就是主体；2) prompt 中显式指代；3) 必要时减少参考图数量。"
  },
  "pricing_guide": {
    "billing_methods": [
      {
        "formula": "最终价格 = 基础价格 × 所有选项系数的乘积 + 所有选项加成的总和",
        "method": "按次"
      },
      {
        "formula": "最终价格 = (输入token数 × 输入token价格 + 输出token数 × 输出token价格) ÷ 1000000",
        "method": "按token"
      },
      {
        "formula": "最终价格 = 时长秒数 × 每秒价格",
        "method": "按秒"
      }
    ],
    "description": "价格计算说明，每个模型的价格通过 GET /v1/skills/models/{name}/pricing 获取"
  }
}
```

## 当前视频模型列表

```json
{
  "models": [
    {
      "name": "grok-imagine-video-1.5-preview",
      "display_name": "grok Imagine video1.5",
      "type": "video",
      "description": "xAI official Imagine 1.5 video model, focused on image-to-video: upload a single first-frame reference image to generate a 1-15s high-quality short video with built-in audio. Aspect ratio and duration are flexibly adjustable. Fast response and low cost.",
      "input_hint": "Upload a first-frame reference image and describe the desired motion and camera movement (this model only supports image-to-video; a first-frame reference image is required)",
      "tags": [
        "Image-to-video",
        "First-frame reference",
        "Built-in audio",
        "1-15s",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "sora-2",
      "display_name": "Sora-2 Official",
      "type": "video",
      "description": "sora-2 is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Describe the action, scene, camera movement, and atmosphere. Use matching landscape/portrait reference images for best results and avoid restricted public-figure or copyright-IP content.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Stable"
      ],
      "available_for_this_key": true
    },
    {
      "name": "grok-video-3",
      "display_name": "grok-video-3",
      "type": "video",
      "description": "grok-video-3 is a Grok video generation model for fast image-to-video and text-to-video creation with short-video friendly durations and aspect ratios.",
      "input_hint": "Describe the video action, scene, and atmosphere. Upload a first-frame reference image for better control when supported.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "First-frame reference",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kwvideo-v2-ref",
      "display_name": "SD 2.0 Reference-to-Video",
      "type": "video",
      "description": "kwvideo-v2-ref is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "Reference-to-Videovideo",
        "Seedream",
        "720p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "doubao-seedance-1-5-pro-251215",
      "display_name": "Seedream 3.5 Pro",
      "type": "video",
      "description": "Seedream 3.5 Pro is a ByteDance Seedance video model for high-quality video generation with sound, motion, and reference-frame control.",
      "input_hint": "Describe your request clearly and include any files or context needed for the model to complete the task.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "First-frame reference",
        "First/last frames",
        "Video with audio",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-v3-omni-cankao",
      "display_name": "Kling-Omni Reference-to-Video",
      "type": "video",
      "description": "Kling-Omni Reference-to-Video is a Kling video generation model for text-to-video, image-to-video, reference-based video, motion control, and high-quality cinematic output.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "Reference-to-Videovideo",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kwvideo-v2-quannengcankao",
      "display_name": "SD 2.0 All-purpose Reference",
      "type": "video",
      "description": "kwvideo-v2-quannengcankao is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-r2v",
      "display_name": "HappyHorse-Reference-to-Video",
      "type": "video",
      "description": "HappyHorse-Reference-to-Video is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kwvideo-v2",
      "display_name": "SD 2.0 First/Last Frames",
      "type": "video",
      "description": "kwvideo-v2 is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "First/last frames",
        "Seedream",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-video-edit",
      "display_name": "HappyHorse-Video Editing",
      "type": "video",
      "description": "HappyHorse-Video Editing is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Upload the source video to edit, optionally add reference images, and describe what to change, such as outfit, background, color, style, or local details.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-1.1-r2v",
      "display_name": "HappyHorse 1.1-Reference-to-Video",
      "type": "video",
      "description": "HappyHorse-Reference-to-Video is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "viduq3-turbo-cankaosheng",
      "display_name": "Vidu Q3 Turbo Reference-to-Video",
      "type": "video",
      "description": "Vidu Q3 Turbo reference-to-video: upload 1-7 reference images and AI generates a subject-consistent audio video based on the subjects; without images it falls back to text-to-video. Billed per second.",
      "input_hint": "Upload 1-7 reference images and describe the video; AI generates an audio video based on the subjects. Without images it generates from text. Supports 540P/720P/1080P, 3-16 seconds.",
      "tags": [
        "Reference-to-Video",
        "Text-to-Video",
        "Multi-image Reference",
        "Audio Video",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "viduq3",
      "display_name": "Vidu Q3",
      "type": "video",
      "description": "Vidu Q3 is a Vidu video model for text-to-video, image-to-video, reference-based generation, character consistency, and audio-synchronized video output.",
      "input_hint": "Describe the video content, characters, actions, camera movement, and rhythm. Upload reference assets when the selected mode requires them.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "First-frame reference",
        "First/last frames",
        "Video with audio",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "viduq3-cankaosheng",
      "display_name": "Vidu Q3 Reference-to-Video",
      "type": "video",
      "description": "Vidu Q3 Reference-to-Video is a Vidu video model for text-to-video, image-to-video, reference-based generation, character consistency, and audio-synchronized video output.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-1.1-t2v",
      "display_name": "HappyHorse 1.1-Text-to-video",
      "type": "video",
      "description": "HappyHorse-Text-to-video is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Describe the scene, motion, camera movement, and visual style. Upload required first-frame, reference, or video-edit assets based on the selected mode.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "omni_flash-10s",
      "display_name": "Omni Flash 10s",
      "type": "video",
      "description": "Powered by Google's Gemini multimodal model, Omni Flash 10s turns a single prompt into a 10-second 720P video, and lets you upload up to 7 reference images to lock in characters, objects and scenes. Videos come with built-in sound, smooth motion and precise prompt understanding—ideal for quickly producing short videos, ads and creative content.",
      "input_hint": "Describe the video action, camera and scene in text. Optionally upload up to 7 reference images for character/object/scene reference; leave empty for text-to-video. Supports landscape 16:9 and portrait 9:16, fixed 10s 720P.",
      "tags": [
        "Text-to-video",
        "Reference-to-video",
        "Multi-image reference",
        "With audio",
        "10s",
        "720P",
        "Google Gemini"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-motion-control-v3",
      "display_name": "Kling-Motion Control V3",
      "type": "video",
      "description": "Kling-Motion Control V3 is a Kling video generation model for text-to-video, image-to-video, reference-based video, motion control, and high-quality cinematic output.",
      "input_hint": "Upload a reference image and action video, then describe how the subject should move or follow the reference motion.",
      "tags": [
        "Motion Control",
        "videoGenerate ",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-v3-omni-shouweizhen",
      "display_name": "Kling-Omni First/Last Frames",
      "type": "video",
      "description": "kling-v3-omni-shouweizhen is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Upload one image for the first frame or two images for first and last frames, then describe the video content and motion.",
      "tags": [
        "Image-to-video",
        "Video with audio",
        "First/last frames",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "veo3.1",
      "display_name": "veo3.1",
      "type": "video",
      "description": "veo3.1 is a Google Veo video generation model with strong camera-control, first/last-frame guidance, and high-quality cinematic output.",
      "input_hint": "Describe the video content, camera movement, and visual evolution. Upload first/last frames when needed and avoid extreme gaps between reference frames.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "First-frame reference",
        "First/last frames",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "viduq3-drama",
      "display_name": "Vidu Q3 Drama",
      "type": "video",
      "description": "Vidu Q3 Drama is a cinematic story-video model built for premium short dramas and AI anime dramas. Upload 1-7 reference images and the AI generates subject-consistent, voiced cinematic videos with accurate character blocking, dialogue and stronger shot-by-shot storytelling. Images are required.",
      "input_hint": "Upload 1-7 reference images and describe the scene and dialogue. The AI will generate a subject-consistent, voiced cinematic story video. Images are required.",
      "tags": [
        "Reference-to-video",
        "Drama",
        "Multi-image reference",
        "Voiced video",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "pixverse-v6-shouweizhen",
      "display_name": "Pix V6 First/Last Frames",
      "type": "video",
      "description": "pixverse-v6-shouweizhen is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Upload one image for the first frame or two images for first and last frames, then describe the video content and motion.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "First/last frames",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-i2v",
      "display_name": "HappyHorse-First Frame",
      "type": "video",
      "description": "HappyHorse-First Frame is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Describe the scene, motion, camera movement, and visual style. Upload required first-frame, reference, or video-edit assets based on the selected mode.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-v3-video",
      "display_name": "Kling-V3-video",
      "type": "video",
      "description": "Kling-V3-video is a Kling video generation model for text-to-video, image-to-video, reference-based video, motion control, and high-quality cinematic output.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "First/last frames",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-motion-control",
      "display_name": "Kling-Motion Control",
      "type": "video",
      "description": "Kling-Motion Control is a Kling video generation model for text-to-video, image-to-video, reference-based video, motion control, and high-quality cinematic output.",
      "input_hint": "Upload a reference image and action video, then describe how the subject should move or follow the reference motion.",
      "tags": [
        "Motion Control",
        "videoGenerate "
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-t2v",
      "display_name": "HappyHorse-Text-to-video",
      "type": "video",
      "description": "HappyHorse-Text-to-video is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Describe the scene, motion, camera movement, and visual style. Upload required first-frame, reference, or video-edit assets based on the selected mode.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-v3-omni-videoref",
      "display_name": "Kling-Omni Video Reference",
      "type": "video",
      "description": "Kling-Omni Video Reference is a Kling video generation model for text-to-video, image-to-video, reference-based video, motion control, and high-quality cinematic output.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "Video Reference",
        "Video Editing",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-avatar-image2video",
      "display_name": "Kling-Digital Human",
      "type": "video",
      "description": "Kling-Digital Human is a Kling image model for text-to-image generation and image editing with strong semantic understanding and reference-image consistency.",
      "input_hint": "Describe the image you want to generate, including subject, scene, style, lighting, and composition. You may upload reference images when supported.",
      "tags": [
        "Digital Human",
        "videoGenerate "
      ],
      "available_for_this_key": true
    },
    {
      "name": "vidu-mv",
      "display_name": "VIDU-Music MV",
      "type": "video",
      "description": "VIDU-Music MV is a Vidu music video model that generates narrated or music-driven videos from audio and reference images, billed by actual duration.",
      "input_hint": "After selecting an audio file, describe the storyboard or visual style. For best results, follow the examples in the usage guide instead of copying lyrics directly.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "wan2.2-animate-mix",
      "display_name": "Wanxiang-Video Face Swap",
      "type": "video",
      "description": "Wanxiang-Video Face Swap is an Alibaba Tongyi Wanxiang video model for text-to-video, image-to-video, reference-based generation, and cinematic video creation.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "Video Face Swap"
      ],
      "available_for_this_key": true
    },
    {
      "name": "wan2.6-cankaosheng",
      "display_name": "Wanxiang 2.6 Reference-to-Video",
      "type": "video",
      "description": "Wanxiang 2.6 Reference-to-Video is an Alibaba Tongyi Wanxiang video model for text-to-video, image-to-video, reference-based generation, and cinematic video creation.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "Reference-to-Videovideo",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "happyhorse-1.1-i2v",
      "display_name": "HappyHorse 1.1-First Frame",
      "type": "video",
      "description": "HappyHorse-First Frame is an Alibaba Bailian HappyHorse video model for text-to-video, image-to-video, reference-based generation, or video editing depending on the mode.",
      "input_hint": "Describe the scene, motion, camera movement, and visual style. Upload required first-frame, reference, or video-edit assets based on the selected mode.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "wan2.6-shouzheng",
      "display_name": "Wanxiang 2.6 First Frame",
      "type": "video",
      "description": "Wanxiang 2.6 First Frame is an Alibaba Tongyi Wanxiang video model for text-to-video, image-to-video, reference-based generation, and cinematic video creation.",
      "input_hint": "Upload one image for the first frame or two images for first and last frames, then describe the video content and motion.",
      "tags": [
        "Image-to-video",
        "Text-to-video",
        "Video with audio",
        "1080p",
        "First-frame reference",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "pixverse-c1-cankaosheng",
      "display_name": "Pix C1 Reference-to-Video",
      "type": "video",
      "description": "Pix C1 Reference-to-Video is a PixVerse video generation model for text-to-video, image-to-video, reference-based generation, and dynamic short-video creation.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "viduq3-turbo",
      "display_name": "Vidu Q3 Turbo",
      "type": "video",
      "description": "Vidu Q3 Turbo video model: upload 1 image for image-to-video, 2 images for first-last frame transition, with synchronized audio. Billed per second, fast generation.",
      "input_hint": "Upload images to generate an audio video: 1 image = image-to-video (first frame), 2 images = first-last frame transition. Supports 540P/720P/1080P, 1-16 seconds.",
      "tags": [
        "Image-to-Video",
        "First-Last Frame",
        "Audio Video",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "viduq2-cankaosheng",
      "display_name": "Vidu Q2 Reference-to-Video",
      "type": "video",
      "description": "Vidu Q2 Reference-to-Video is a Vidu video model for text-to-video, image-to-video, reference-based generation, character consistency, and audio-synchronized video output.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "Reference-to-Videovideo",
        "1080p",
        "HD",
        "Video with audio"
      ],
      "available_for_this_key": true
    },
    {
      "name": "pixverse-c1-shouweizhen",
      "display_name": "Pix C1 First/Last Frames",
      "type": "video",
      "description": "pixverse-c1-shouweizhen is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Upload one image for the first frame or two images for first and last frames, then describe the video content and motion.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "wan2.7-cankaosheng",
      "display_name": "Wanxiang 2.7 Reference-to-Video",
      "type": "video",
      "description": "Wanxiang 2.7 Reference-to-Video is an Alibaba Tongyi Wanxiang video model for text-to-video, image-to-video, reference-based generation, and cinematic video creation.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "Text-to-video",
        "Reference-to-Videovideo",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "kling-v2-6",
      "display_name": "Kling 2.6 Pro",
      "type": "video",
      "description": "Kling 2.6 Pro is a Kling video generation model for text-to-video, image-to-video, reference-based video, motion control, and high-quality cinematic output.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "vidu-jieshuoman",
      "display_name": "VIDU-Narrated Comic",
      "type": "video",
      "description": "VIDU-Narrated Comic is a Vidu video model for text-to-video, image-to-video, reference-based generation, character consistency, and audio-synchronized video output.",
      "input_hint": "Describe the video content, characters, actions, camera movement, and rhythm. Upload reference assets when the selected mode requires them.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "pixverse-v5.6-r2v",
      "display_name": "Pix V5.6 Reference-to-Video",
      "type": "video",
      "description": "Pix V5.6 Reference-to-Video is a PixVerse video generation model for text-to-video, image-to-video, reference-based generation, and dynamic short-video creation.",
      "input_hint": "Upload reference images or videos as required, then describe the scene, characters, actions, camera movement, and desired style.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "pixverse-v5.6-shouweizhen",
      "display_name": "Pix V5.6 First/Last Frames",
      "type": "video",
      "description": "pixverse-v5.6-shouweizhen is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Upload one image for the first frame or two images for first and last frames, then describe the video content and motion.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "Video with audio",
        "First/last frames",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "wan2.7-shouweizhen",
      "display_name": "Wanxiang 2.7 First/Last Frames",
      "type": "video",
      "description": "wan2.7-shouweizhen is an AI video model for text-to-video, image-to-video, reference generation, or video editing workflows.",
      "input_hint": "Upload one image for the first frame or two images for first and last frames, then describe the video content and motion.",
      "tags": [
        "Image-to-video",
        "First/last frames",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    },
    {
      "name": "omni-flash",
      "display_name": "omni-flash",
      "type": "video",
      "description": "Gemini Omni Flash is Google's multimodal AI video model. Supports text-to-video and multi-image reference-to-video, understands real-world physics, and generates high-definition short videos with audio. Generate from text alone, or upload 1-3 reference images to guide characters / objects / scene style. Supports 16:9 / 9:16 aspect ratios and 6 / 8 / 10 second durations.",
      "input_hint": "Describe the video action, camera and scene in text. Optionally upload 1-3 reference images for characters / objects / scene reference; leave empty for pure text-to-video. Select 16:9 or 9:16 aspect ratio.",
      "tags": [
        "Text-to-video",
        "Reference-to-video",
        "Multi-image reference",
        "Audio video",
        "6/8/10s",
        "Google Gemini"
      ],
      "available_for_this_key": true
    },
    {
      "name": "wan2.7-xuxie",
      "display_name": "Wanxiang 2.7 Video Extension",
      "type": "video",
      "description": "Wanxiang 2.7 Video Extension is an Alibaba Tongyi Wanxiang video model for text-to-video, image-to-video, reference-based generation, and cinematic video creation.",
      "input_hint": "Describe the video content, subject, action, scene, camera movement, and style. Upload required reference images, videos, or audio based on the selected mode.",
      "tags": [
        "AI feature"
      ],
      "available_for_this_key": true
    },
    {
      "name": "hailuo-2.3",
      "display_name": "Hailuo 2.3",
      "type": "video",
      "description": "Hailuo 2.3 is a MiniMax Hailuo creative model for speech, music, or video generation depending on the selected workflow.",
      "input_hint": "Describe your request clearly and include any files or context needed for the model to complete the task.",
      "tags": [
        "Text-to-video",
        "Image-to-video",
        "1080p",
        "HD"
      ],
      "available_for_this_key": true
    }
  ],
  "total": 45,
  "type": "video"
}
```

## Grok Imagine Video 1.5 参数

```json
{
  "description": "xAI official Imagine 1.5 video model, focused on image-to-video: upload a single first-frame reference image to generate a 1-15s high-quality short video with built-in audio. Aspect ratio and duration are flexibly adjustable. Fast response and low cost.",
  "display_name": "grok Imagine video1.5",
  "input_hint": "Upload a first-frame reference image and describe the desired motion and camera movement (this model only supports image-to-video; a first-frame reference image is required)",
  "name": "grok-imagine-video-1.5-preview",
  "params": [
    {
      "name": "prompt",
      "label": "Prompt",
      "type": "textarea",
      "required": true,
      "description": "Describe the video scene and action you want to generate"
    },
    {
      "name": "images",
      "label": "First-frame reference",
      "type": "upload",
      "required": true,
      "description": "Upload 1 image as the video first frame (image-to-video only, required)。（仅接受可公开访问的 URL；多文件可传 URL 数组。平台不提供文件托管，请自行将文件上传至 COS/CDN 等对象存储服务后传入 URL）"
    },
    {
      "name": "aspect_ratio",
      "label": "Aspect ratio",
      "type": "radio",
      "required": true,
      "description": "Choose the video aspect ratio",
      "options": [
        {
          "value": "16:9",
          "label": "Landscape 16:9"
        },
        {
          "value": "9:16",
          "label": "Portrait 9:16"
        },
        {
          "value": "1:1",
          "label": "Square 1:1"
        },
        {
          "value": "3:2",
          "label": "Landscape 3:2"
        },
        {
          "value": "2:3",
          "label": "Portrait 2:3"
        }
      ]
    },
    {
      "name": "resolution",
      "label": "Resolution",
      "type": "radio",
      "required": true,
      "description": "Choose the video resolution",
      "options": [
        {
          "value": "720p",
          "label": "HD 720p"
        },
        {
          "value": "480p",
          "label": "SD 480p"
        }
      ]
    },
    {
      "name": "duration",
      "label": "Duration",
      "type": "select",
      "required": true,
      "description": "Choose the video duration (1-15s, billed per second)",
      "options": [
        {
          "value": "1",
          "label": "1s"
        },
        {
          "value": "2",
          "label": "2s"
        },
        {
          "value": "3",
          "label": "3s"
        },
        {
          "value": "4",
          "label": "4s"
        },
        {
          "value": "5",
          "label": "5s"
        },
        {
          "value": "6",
          "label": "6s"
        },
        {
          "value": "7",
          "label": "7s"
        },
        {
          "value": "8",
          "label": "8s"
        },
        {
          "value": "9",
          "label": "9s"
        },
        {
          "value": "10",
          "label": "10s"
        },
        {
          "value": "11",
          "label": "11s"
        },
        {
          "value": "12",
          "label": "12s"
        },
        {
          "value": "13",
          "label": "13s"
        },
        {
          "value": "14",
          "label": "14s"
        },
        {
          "value": "15",
          "label": "15s"
        }
      ]
    }
  ],
  "tags": [
    "Image-to-video",
    "First-frame reference",
    "Built-in audio",
    "1-15s",
    "HD"
  ],
  "type": "video"
}
```

## Grok Imagine Video 1.5 计费

```json
{
  "available_for_this_key": true,
  "channel_groups": [
    {
      "group_name": "MC-gork1.5官转分组",
      "is_active": true,
      "in_key_whitelist": true,
      "billing_method": "按秒",
      "base_price": 0.1035,
      "input_token_price": 0,
      "output_token_price": 0,
      "success_rate_24h": 98.93,
      "avg_response_seconds": 73.81,
      "sample_count_1h": 187,
      "total_success": 733234,
      "total_fail": 70724,
      "option_prices": []
    },
    {
      "group_name": "ZZ-优质grok",
      "is_active": false,
      "in_key_whitelist": true,
      "billing_method": "按秒",
      "base_price": 0.5175,
      "input_token_price": 0,
      "output_token_price": 0,
      "success_rate_24h": -1,
      "avg_response_seconds": 0,
      "sample_count_1h": 0,
      "total_success": 3388,
      "total_fail": 1685,
      "option_prices": [
        {
          "param_name": "resolution",
          "option_label": "高清 720p",
          "option_value": "720p",
          "price_multiplier": 1.5,
          "price_addition": 0,
          "final_price": 0.7762499999999999,
          "price_impact": "x1.5"
        }
      ]
    }
  ],
  "display_name": "grok-video-3.5",
  "filter": "",
  "key_channel_strategy": "价格优先",
  "name": "grok-imagine-video-1.5-preview",
  "pricing_note": "默认返回所有渠道分组（含已关闭的），加 ?status=active 仅返回当前启用的分组。实际调用时不需要指定渠道分组（请求体 channel_group 字段、请求头 X-Channel-Group 均被忽略），系统根据 API Key 的渠道策略自动选择。三种策略说明参见 GET /v1/skills/guide 的 channel_strategy 部分。is_active 状态会随上游可用性实时切换，某次任务实际命中的分组是任务创建瞬间 is_active=1 且符合策略排序的分组，事后查询可能看到该分组已 is_active=false；扣费金额匹配任务瞬间命中分组的价格，不一定能在当前 is_active=true 的列表里找到对应档位。若某参数选项未出现在 option_prices 中，表示该选项使用分组的基础价格（base_price），无额外加价。",
  "type": "video"
}
```

