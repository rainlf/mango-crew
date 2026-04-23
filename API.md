# Mango Crew API 文档

## 基础信息

- Base URL：`http://localhost:8080`
- API 前缀：`/api`
- Content-Type：除特别说明外，默认使用 `application/json`
- 用户标识：通过请求头 `X-User-ID` 或查询参数 `userId` 传递

## 响应格式

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

业务错误示例：

```json
{
  "code": 1,
  "message": "具体错误信息",
  "data": null
}
```

说明：

- 参数校验失败通常返回 HTTP `400` / `404`。
- 业务处理失败通常返回 HTTP `200`，并在响应体中设置 `code != 0`。

## 枚举说明

### 游戏类型 `game_type`

| 值 | 名称 |
| --- | --- |
| 1 | 平胡 |
| 2 | 自摸 |
| 3 | 一炮双响 |
| 4 | 一炮三响 |
| 5 | 相公 |
| 6 | 运动 |

### 玩家角色 `role`

| 值 | 名称 |
| --- | --- |
| 1 | 赢家 |
| 2 | 输家 |
| 3 | 记录者 |
| 4 | 参与者 |

### 游戏状态 `status`

| 值 | 名称 |
| --- | --- |
| 0 | 待确认 |
| 1 | 已确认 |
| 2 | 已取消 |

### 场次状态 `status`

| 值 | 名称 |
| --- | --- |
| 0 | 进行中 |
| 1 | 已结束 |

## 健康检查

### GET `/api/health`

返回纯文本 `OK`。

## 用户接口

### GET `/api/user/login`

通过微信小程序 `code` 登录或注册用户。

请求参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `code` | query | 是 | 微信登录临时凭证 |

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1
  }
}
```

### GET `/api/user/info`

获取指定用户信息与统计数据。

请求参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `userId` | query | 是 | 用户 ID |

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "nickname": "张三",
    "avatar_url": "https://example.com/avatar.jpg",
    "total_points": 150,
    "total_games": 20,
    "win_count": 8
  }
}
```

### POST `/api/user/update`

更新用户昵称或头像。

请求类型：`application/x-www-form-urlencoded` 或 `multipart/form-data`

请求参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `userId` | form | 是 | 用户 ID |
| `nickname` | form | 否 | 新昵称 |
| `avatar` | form | 否 | base64 图片内容 |

说明：

- 当前代码会读取 `avatar` 字段，但不会真正上传图片，头像 URL 暂未落库更新。

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "nickname": "张三",
    "avatar_url": ""
  }
}
```

### GET `/api/user/rank`

获取用户排行榜。

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "nickname": "张三",
      "avatar_url": "",
      "total_points": 200,
      "total_games": 25,
      "win_count": 10
    }
  ]
}
```

### GET `/api/user/list`

获取所有用户的基础信息列表。

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "nickname": "张三",
      "avatar_url": ""
    }
  ]
}
```

## 对局接口

### POST `/api/game`

创建一盘对局记录。

请求头或参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `X-User-ID` | header | 否 | 当前用户 ID，建议传递 |
| `userId` | query | 否 | 当前用户 ID，未传 Header 时可用 |

请求体：

```json
{
  "game_type": 2,
  "remark": "清一色自摸",
  "players": [
    {
      "user_id": 1,
      "seat": 1,
      "role": 1,
      "base_points": 10,
      "win_types": ["qing_yi_se"]
    },
    {
      "user_id": 2,
      "seat": 2,
      "role": 2
    }
  ]
}
```

字段说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `game_type` | 是 | 游戏类型，见上方枚举 |
| `remark` | 否 | 备注，最长 200 字符 |
| `players` | 是 | 玩家列表 |

`players` 子项说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `user_id` | 是 | 用户 ID |
| `seat` | 是 | 座位号，范围 1-4 |
| `role` | 是 | 玩家角色 |
| `base_points` | 否 | 基础分 |
| `win_types` | 否 | 番型 code 数组 |

校验规则：

- `game_type=6`（运动）时，`players` 必须只有 1 人。
- 非运动类型至少需要 2 名玩家。
- 必须至少存在 1 名 `role=1` 的赢家。

### POST `/api/game/players`

更新当前牌桌玩家。用于“加入牌桌 / 踢人后自己加入 / 固定 4 人后连续记局”的场景。

请求体：

```json
{
  "user_ids": [1, 2, 3, 4]
}
```

说明：

- 允许先保存 1-4 名当前玩家。
- 非运动类型真正记局时，后端会要求当前牌桌恰好为 4 人。
- 成功后返回最新的 `current_players` 和 `all_players` 摘要。

### POST `/api/game/record`

按麻将记牌场景直接记录一局“已结算”对局。该接口会更新当前牌桌玩家，并保留当前 4 人状态用于继续记录下一局。

请求体：

```json
{
  "gameType": 1,
  "players": [1, 2, 3, 4],
  "recorderId": 1,
  "winners": [
    {
      "userId": 2,
      "basePoints": 5,
      "winTypes": ["qing_yi_se"]
    }
  ],
  "losers": [3],
  "remark": "清一色点炮"
}
```

字段说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `gameType` | 是 | 游戏类型，1-平胡 2-自摸 3-一炮双响 4-一炮三响 5-相公 6-运动 |
| `players` | 是 | 当前牌桌玩家 ID 列表；非运动类型必须最终形成 4 人 |
| `recorderId` | 是 | 本次记录的操作人 |
| `winners` | 是 | 赢家列表 |
| `losers` | 否 | 输家列表，运动类型可为空 |
| `remark` | 否 | 备注 |

校验规则：

- 平胡：1 个赢家 + 1 个输家，其余玩家记为参与者，分数为 0。
- 自摸：1 个赢家 + 其余玩家全部输家。
- 一炮双响：2 个赢家 + 1 个输家。
- 一炮三响：3 个赢家 + 1 个输家。
- 相公：3 个赢家 + 1 个输家。
- 运动：只允许 1 个当前玩家，且没有输家。

积分规则：

- 赢家分数 = `basePoints * 所有番型倍数连乘`
- 平胡：输家 = `-赢家分数`
- 自摸：赢家 = `单家分 * 3`，3 个输家各自扣 `单家分`
- 一炮双响 / 一炮三响 / 相公：输家扣所有赢家总和
- 非运动类型下，如果记录者本人在当前牌桌中，会额外获得随机记录奖励：1% 为 `20`，99% 为 `1`

响应说明：

- 返回新创建的 `game` 主记录，状态直接为 `已确认`

积分规则说明：

- `final_points = base_points * 所有番型 multiplier 连乘`
- 记录者在非运动类型中会额外得到随机奖励：1% 概率 `20` 分，99% 概率 `1` 分

响应说明：

- 当前接口返回的是新建 `game` 记录本身，不包含完整 DTO 展开数据。

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "type": 2,
    "status": 0,
    "remark": "清一色自摸",
    "created_by": 1,
    "created_at": "2026-04-21T10:10:00Z"
  }
}
```

### POST `/api/game/cancel`

取消一盘对局。

请求体：

```json
{
  "game_id": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### GET `/api/game/recent`

分页获取最近的对局列表。

请求参数：

| 参数 | 位置 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `limit` | query | 否 | 10 | 每页数量 |
| `offset` | query | 否 | 0 | 偏移量 |

响应数据结构：

- `type`：游戏类型中文名
- `type_code`：游戏类型数值编码
- `created_by`：创建者基础信息
- `players[].role`：玩家角色中文名
- `players[].role_code`：玩家角色编码
- `players[].win_types`：玩家番型列表

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "type": "自摸",
      "type_code": 2,
      "status": 1,
      "remark": "清一色自摸",
      "created_by": {
        "id": 1,
        "nickname": "张三",
        "avatar_url": ""
      },
      "players": [
        {
          "id": 1,
          "user": {
            "id": 1,
            "nickname": "张三",
            "avatar_url": ""
          },
          "seat": 1,
          "role": "赢家",
          "role_code": 1,
          "base_points": 10,
          "final_points": 40,
          "win_types": [
            {
              "code": "qing_yi_se",
              "name": "清一色",
              "multiplier": 4
            }
          ]
        }
      ],
      "created_at": "2026-04-21 10:10:00",
      "settled_at": "2026-04-21 10:15:00"
    }
  ]
}
```

### GET `/api/game/user/list`

分页获取某个用户参与过的已结算对局。

请求参数：

| 参数 | 位置 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `userId` | query | 是 | - | 用户 ID |
| `limit` | query | 否 | 10 | 每页数量 |
| `offset` | query | 否 | 0 | 偏移量 |

响应结构与 `/api/game/recent` 一致。

### GET `/api/game/players`

获取当前牌桌玩家摘要。

响应字段说明：

| 字段 | 说明 |
| --- | --- |
| `current_players` | 当前活跃牌桌中的玩家 |
| `all_players` | 全量用户列表 |

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "current_players": [
      {
        "id": 1,
        "nickname": "张三",
        "avatar_url": ""
      }
    ],
    "all_players": [
      {
        "id": 1,
        "nickname": "张三",
        "avatar_url": ""
      },
      {
        "id": 2,
        "nickname": "李四",
        "avatar_url": ""
      }
    ]
  }
}
```

## 数据库关联说明

项目默认数据库为 `mango_crew`，接口数据主要来源于以下表：

| 表名 | 作用 |
| --- | --- |
| `user` | 用户主数据 |
| `game_player` | 当前牌桌玩家 |
| `game` | 单盘对局 |
| `game_record` | 每盘对局的记录明细，内含赢家番型 JSON |

推荐初始化方式：

```bash
mysql -uroot -p mango_crew < migrations/init.sql
```

## 当前实现说明

- 目前没有实现 `/api/win-type/list`、`/api/game/detail`、`/api/game/stats` 等接口。
- `POST /api/user/update` 会读取头像字段，但不会真正上传图片。
- `POST /api/game` 如果未传用户 ID，创建者会落成 `0`。
