# ABAC 策略表达式示例

本文档展示基于 Expr 的 ABAC 策略表达式用法。

## 语法说明

- 使用类 Go 语法
- 结果必须为布尔值
- 支持属性访问：`User.属性名`、`Resource.属性名`
- 支持运算符：`==`、`!=`、`>`、`<`、`>=`、`<=`、`&&`、`||`、`!`、`in`

## 上下文变量

### User（用户属性）
```go
User.ID           // 用户 ID（字符串）
User.Username     // 用户名
User.Email        // 邮箱
User.TenantID     // 租户 ID
User.Status       // 状态
User.DepartmentID // 部门 ID（自定义）
User.Level        // 级别（自定义）
User.Roles        // 角色列表（自定义）
```

### Resource（资源属性）
```go
Resource.OwnerID      // 资源所有者 ID
Resource.DepartmentID // 资源所属部门
Resource.Status       // 资源状态
Resource.Level        // 资源级别
```

### Environment（环境属性）
```go
Environment.IPAddress // 请求 IP
Environment.Location  // 地理位置
```

### Time（时间属性）
```go
Time.Hour    // 小时 (0-23)
Time.Day     // 日期 (1-31)
Time.Weekday // 星期 (0-6, 0=周日)
Time.Month   // 月份 (1-12)
Time.Year    // 年份
```

---

## 表达式示例

### 1. 基础权限检查

#### 同部门可访问
```javascript
User.DepartmentID == Resource.DepartmentID
```

#### 用户级别大于等于 3
```javascript
User.Level >= 3
```

#### 组合条件
```javascript
User.DepartmentID == Resource.DepartmentID && User.Level >= 3
```

---

### 2. 角色检查

#### 检查用户是否为管理员
```javascript
"admin" in User.Roles
```

#### 检查用户是否为经理或管理员
```javascript
"manager" in User.Roles || "admin" in User.Roles
```

---

### 3. 资源所有权

#### 资源所有者可访问
```javascript
User.ID == Resource.OwnerID
```

#### 所有者或同部门经理可访问
```javascript
User.ID == Resource.OwnerID ||
(User.DepartmentID == Resource.DepartmentID && "manager" in User.Roles)
```

---

### 4. 时间限制

#### 工作时间（9:00-18:00）
```javascript
Time.Hour >= 9 && Time.Hour <= 18
```

#### 工作日（周一到周五）
```javascript
Time.Weekday >= 1 && Time.Weekday <= 5
```

#### 工作时间且工作日
```javascript
Time.Hour >= 9 && Time.Hour <= 18 &&
Time.Weekday >= 1 && Time.Weekday <= 5
```

---

### 5. 状态检查

#### 已发布的资源可访问
```javascript
Resource.Status == "published"
```

#### 已发布或用户是所有者
```javascript
Resource.Status == "published" || User.ID == Resource.OwnerID
```

---

### 6. 级别控制

#### 用户级别必须高于资源级别
```javascript
User.Level > Resource.Level
```

#### 机密资源仅经理可访问
```javascript
Resource.Level == "confidential" && "manager" in User.Roles
```

---

### 7. 地理位置限制

#### 仅中国大陆可访问
```javascript
Environment.Location == "CN"
```

#### 仅内网 IP 可访问
```javascript
Environment.IPAddress startsWith "192.168."
```

---

### 8. 复杂组合

#### 综合策略示例 1
```javascript
// 同部门且级别>=3，或者是管理员，工作时间内
(User.DepartmentID == Resource.DepartmentID && User.Level >= 3) ||
"admin" in User.Roles &&
Time.Hour >= 9 && Time.Hour <= 18
```

#### 综合策略示例 2
```javascript
// 资源所有者，或同部门经理，且资源未删除
(User.ID == Resource.OwnerID ||
 (User.DepartmentID == Resource.DepartmentID && "manager" in User.Roles)) &&
Resource.Status != "deleted"
```

#### 综合策略示例 3
```javascript
// 已发布资源任何人可读，草稿仅所有者可读
Resource.Status == "published" ||
(Resource.Status == "draft" && User.ID == Resource.OwnerID)
```

---

## 函数支持

Expr 支持自定义函数，可扩展以下功能：

```javascript
// 检查用户是否在组
inGroup("sales")

// 检查用户是否有角色
hasRole("manager")

// 时间范围检查
timeInRange("09:00", "18:00")

// IP 范围检查
ipInRange("192.168.0.0/16")

// 字符串匹配
matches(User.Email, ".*@company.com")
```

---

## 实际应用场景

### 场景 1：文档管理系统

**策略 1**：同部门可读
```javascript
User.DepartmentID == Resource.DepartmentID
```

**策略 2**：所有者可编辑
```javascript
User.ID == Resource.OwnerID
```

**策略 3**：管理员全权限
```javascript
"admin" in User.Roles
```

**策略 4**：机密文档仅经理可读
```javascript
Resource.Level != "confidential" || "manager" in User.Roles
```

---

### 场景 2：审批系统

**策略 1**：直属上级可审批
```javascript
User.ID == Resource.ManagerID
```

**策略 2**：金额超过 10000 需总监审批
```javascript
Resource.Amount <= 10000 || "director" in User.Roles
```

**策略 3**：工作时间审批
```javascript
Time.Hour >= 9 && Time.Hour <= 18 && Time.Weekday >= 1 && Time.Weekday <= 5
```

---

### 场景 3：数据访问控制

**策略 1**：用户只能访问自己的数据
```javascript
User.ID == Resource.UserID
```

**策略 2**：部门数据仅部门成员可访问
```javascript
User.DepartmentID == Resource.DepartmentID
```

**策略 3**：数据分析师可读所有数据
```javascript
"analyst" in User.Roles
```

---

## 性能优化建议

1. **表达式缓存**：Expr 自动缓存编译后的程序，无需手动优化
2. **简化表达式**：避免过于复杂的嵌套，拆分为多个策略
3. **优先级设置**：高频匹配的策略设置高优先级
4. **属性预加载**：将常用属性存入 User.Metadata，避免多次查询

---

## 调试技巧

### 1. 验证表达式语法
```go
err := engine.ValidatePolicyExpression("User.Level >= 3")
if err != nil {
    log.Fatal(err)
}
```

### 2. 测试策略
```go
matched, err := engine.EvaluatePolicy(ctx, policy, userID, resourceAttrs, envAttrs)
fmt.Printf("Matched: %v, Error: %v\n", matched, err)
```

### 3. 打印上下文
```go
fmt.Printf("User: %+v\n", evalCtx.User)
fmt.Printf("Resource: %+v\n", evalCtx.Resource)
```
