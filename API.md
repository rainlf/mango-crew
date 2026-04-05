# Mango Crew API 接口文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **Content-Type**: `application/json`
- **认证方式**: 通过 Header `X-User-ID` 或 Query 参数 `userId` 传递用户ID

---

## 用户接口

### 1. 用户登录

**POST** `/user/login`

微信小程序登录，通过 code 换取用户身份。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 微信登录临时凭证 |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1
  }
}
```

---

### 2. 获取用户信息

**GET** `/user/info`

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | int | 是 | 用户ID |

#### 响应示例

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
    "win_count": 8,
    "tags": ["清一色", "自摸"]
  }
}
```

---

### 3. 更新用户信息

**POST** `/user/update`

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userId | int | 是 | 用户ID |
| nickname | string | 否 | 昵称 |
| avatar | string | 否 | 头像 base64 |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "nickname": "张三",
    "avatar_url": "https://example.com/avatar.jpg"
  }
}
```

---

### 4. 用户排名

**GET** `/user/rank`

获取所有用户按积分排名的列表。

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "nickname": "张三",
      "avatar_url": "https://example.com/avatar.jpg",
      "total_points": 200,
      "total_games": 25,
      "win_count": 10
    },
    {
      "id": 2,
      "nickname": "李四",
      "avatar_url": "https://example.com/avatar2.jpg",
      "total_points": 150,
      "total_games": 20,
      "win_count": 8
    }
  ]
}
```

---

### 5. 获取所有用户

**GET** `/user/list`

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "nickname": "张三",
      "avatar_url": "https://example.com/avatar.jpg"
    },
    {
      "id": 2,
      "nickname": "李四",
      "avatar_url": "https://example.com/avatar2.jpg"
    }
  ]
}
```

---

## 场次接口

### 1. 创建场次

**POST** `/session`

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 场次名称，如"周五晚场" |

#### 请求示例

```json
{
  "name": "周五晚场"
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "周五晚场",
    "status": 0,
    "created_by": {
      "id": 1,
      "nickname": "张三"
    },
    "game_count": 0,
    "created_at": "2024-01-15 18:00:00"
  }
}
```

---

### 2. 结束场次

**POST** `/session/end`

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionId | int | 是 | 场次ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 3. 获取场次列表

**GET** `/session/list`

#### 请求参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| limit | int | 否 | 10 | 每页数量 |
| offset | int | 否 | 0 | 偏移量 |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "周五晚场",
      "status": 0,
      "created_by": {
        "id": 1,
        "nickname": "张三"
      },
      "game_count": 5,
      "created_at": "2024-01-15 18:00:00"
    }
  ]
}
```

---

### 4. 获取进行中的场次

**GET** `/session/active`

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "周五晚场",
      "status": 0,
      "created_by": {
        "id": 1,
        "nickname": "张三"
      },
      "game_count": 5,
      "created_at": "2024-01-15 18:00:00"
    }
  ]
}
```

---

## 游戏接口

### 1. 创建游戏

**POST** `/game`

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| session_id | int | 是 | 场次ID |
| game_type | int | 是 | 游戏类型：1平胡 2自摸 3一炮双响 4一炮三响 5相公 6运动 |
| remark | string | 否 | 备注 |
| players | array | 是 | 玩家列表 |

#### players 结构

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | int | 是 | 用户ID |
| seat | int | 是 | 座位号 1-4 |
| role | int | 是 | 角色：1赢家 2输家 3记录者 |
| base_points | int | 否 | 基础分（赢家必填） |
| win_types | array | 否 | 番型code列表 |

#### 请求示例 - 自摸

```json
{
  "session_id": 1,
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
    },
    {
      "user_id": 3,
      "seat": 3,
      "role": 2
    },
    {
      "user_id": 4,
      "seat": 4,
      "role": 2
    }
  ]
}
```

#### 请求示例 - 平胡（一炮两响）

```json
{
  "session_id": 1,
  "game_type": 1,
  "remark": "",
  "players": [
    {
      "user_id": 1,
      "seat": 1,
      "role": 1,
      "base_points": 5,
      "win_types": ["peng_peng_hu"]
    },
    {
      "user_id": 2,
      "seat": 2,
      "role": 2
    },
    {
      "user_id": 3,
      "seat": 3,
      "role": 1,
      "base_points": 5,
      "win_types": ["wu_hua_guo"]
    },
    {
      "user_id": 4,
      "seat": 4,
      "role": 2
    }
  ]
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "session_id": 1,
    "type": 2,
    "status": 0,
    "remark": "清一色自摸",
    "created_by": {
      "id": 1,
      "nickname": "张三"
    },
    "players": [
      {
        "id": 1,
        "user_id": 1,
        "seat": 1,
        "role": 1,
        "base_points": 10,
        "final_points": 40,
        "win_types": [
          {
            "code": "qing_yi_se",
            "name": "清一色",
            "multiplier": 4
          }
        ],
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar_url": "https://example.com/avatar.jpg"
        }
      },
      {
        "id": 2,
        "user_id": 2,
        "seat": 2,
        "role": 2,
        "base_points": 0,
        "final_points": -10,
        "user": {
          "id": 2,
          "nickname": "李四",
          "avatar_url": "https://example.com/avatar2.jpg"
        }
      }
    ],
    "created_at": "2024-01-15 19:30:00"
  }
}
```

---

### 2. 结算游戏

**POST** `/game/settle`

确认游戏结果并计算最终积分。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| gameId | int | 是 | 游戏ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "session_id": 1,
    "type": 2,
    "status": 1,
    "remark": "清一色自摸",
    "players": [
      {
        "id": 1,
        "user_id": 1,
        "seat": 1,
        "role": 1,
        "base_points": 10,
        "final_points": 40,
        "is_settled": true
      },
      {
        "id": 2,
        "user_id": 2,
        "seat": 2,
        "role": 2,
        "base_points": 0,
        "final_points": -10,
        "is_settled": true
      }
    ],
    "settled_at": "2024-01-15 19:35:00"
  }
}
```

---

### 3. 取消游戏

**POST** `/game/cancel`

取消未结算的游戏。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| gameId | int | 是 | 游戏ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 4. 获取场次游戏列表

**GET** `/game/list`

获取指定场次下的所有游戏记录。

#### 请求参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| sessionId | int | 是 | - | 场次ID |
| limit | int | 否 | 10 | 每页数量 |
| offset | int | 否 | 0 | 偏移量 |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "session_id": 1,
      "type": 2,
      "status": 1,
      "remark": "清一色自摸",
      "players": [
        {
          "id": 1,
          "user_id": 1,
          "seat": 1,
          "role": 1,
          "final_points": 40,
          "is_settled": true,
          "user": {
            "id": 1,
            "nickname": "张三",
            "avatar_url": "https://example.com/avatar.jpg"
          }
        }
      ],
      "created_at": "2024-01-15 19:30:00",
      "settled_at": "2024-01-15 19:35:00"
    }
  ]
}
```

---

### 5. 获取游戏详情

**GET** `/game/detail`

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| gameId | int | 是 | 游戏ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "session_id": 1,
    "type": 2,
    "status": 1,
    "remark": "清一色自摸",
    "created_by": {
      "id": 1,
      "nickname": "张三"
    },
    "players": [
      {
        "id": 1,
        "user_id": 1,
        "seat": 1,
        "role": 1,
        "base_points": 10,
        "final_points": 40,
        "is_settled": true,
        "win_types": [
          {
            "code": "qing_yi_se",
            "name": "清一色",
            "base_multi": 4,
            "description": "同一花色"
          }
        ],
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar_url": "https://example.com/avatar.jpg"
        }
      }
    ],
    "created_at": "2024-01-15 19:30:00",
    "settled_at": "2024-01-15 19:35:00"
  }
}
```

---

### 6. 获取场次统计

**GET** `/game/stats`

获取指定场次的玩家积分统计。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionId | int | 是 | 场次ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 1,
    "game_count": 10,
    "player_stats": [
      {
        "user_id": 1,
        "nickname": "张三",
        "avatar_url": "https://example.com/avatar.jpg",
        "total_points": 150,
        "win_count": 5,
        "game_count": 10
      },
      {
        "user_id": 2,
        "nickname": "李四",
        "avatar_url": "https://example.com/avatar2.jpg",
        "total_points": -50,
        "win_count": 2,
        "game_count": 10
      }
    ]
  }
}
```

---

## 番型接口

### 1. 获取番型列表

**GET** `/win-type/list`

获取所有支持的番型列表。

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "code": "wu_hua_guo",
      "name": "无花果",
      "base_multi": 1,
      "description": "无番型"
    },
    {
      "id": 2,
      "code": "peng_peng_hu",
      "name": "碰碰胡",
      "base_multi": 2,
      "description": "全部由碰牌组成"
    },
    {
      "id": 3,
      "code": "yi_tiao_long",
      "name": "一条龙",
      "base_multi": 2,
      "description": "同一花色1-9"
    },
    {
      "id": 4,
      "code": "hun_yi_se",
      "name": "混一色",
      "base_multi": 2,
      "description": "同一花色加字牌"
    },
    {
      "id": 5,
      "code": "qing_yi_se",
      "name": "清一色",
      "base_multi": 4,
      "description": "同一花色"
    },
    {
      "id": 6,
      "code": "xiao_qi_dui",
      "name": "小七对",
      "base_multi": 4,
      "description": "七个对子"
    },
    {
      "id": 7,
      "code": "long_qi_dui",
      "name": "龙七对",
      "base_multi": 8,
      "description": "小七对加一根"
    },
    {
      "id": 8,
      "code": "da_diao_che",
      "name": "大吊车",
      "base_multi": 2,
      "description": "单吊将牌"
    },
    {
      "id": 9,
      "code": "men_qian_qing",
      "name": "门前清",
      "base_multi": 2,
      "description": "未碰未吃"
    },
    {
      "id": 10,
      "code": "gang_kai_hua",
      "name": "杠开花",
      "base_multi": 2,
      "description": "杠牌后自摸"
    }
  ]
}
```

---

## 数据类型定义

### 枚举类型

#### GameType (游戏类型)

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | 平胡 | 点炮胡牌 |
| 2 | 自摸 | 自摸胡牌 |
| 3 | 一炮双响 | 一炮双响 |
| 4 | 一炮三响 | 一炮三响 |
| 5 | 相公 | 相公 |
| 6 | 运动 | 运动 |

#### PlayerRole (玩家角色)

| 值 | 名称 | 说明 |
|----|------|------|
| 1 | 赢家 | 胡牌玩家 |
| 2 | 输家 | 输分玩家 |
| 3 | 记录者 | 记录者 |

#### GameStatus (游戏状态)

| 值 | 名称 | 说明 |
|----|------|------|
| 0 | 进行中 | 游戏进行中 |
| 1 | 已结算 | 已结算 |
| 2 | 已取消 | 已取消 |

#### SessionStatus (场次状态)

| 值 | 名称 | 说明 |
|----|------|------|
| 0 | 进行中 | 场次进行中 |
| 1 | 已结束 | 已结束 |

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 1001 | 用户不存在 |
| 1002 | 微信登录失败 |
| 2001 | 场次不存在 |
| 2002 | 场次已结束 |
| 3001 | 游戏不存在 |
| 3002 | 游戏已结算 |
| 3003 | 游戏已取消 |
| 3004 | 玩家数量错误 |
| 3005 | 番型不存在 |

---

## 通用响应格式

所有接口统一返回以下格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 错误响应示例

```json
{
  "code": 400,
  "message": "请求参数错误：缺少必填字段 session_id",
  "data": null
}
```
