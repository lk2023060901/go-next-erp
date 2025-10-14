# HRM (人力资源管理) 模块

## 概述

HRM模块是一个独立的人力资源管理系统,专注于考勤管理功能,支持与第三方平台(钉钉、企业微信、飞书)和考勤机设备的数据对接。

## 模块结构

```
internal/hrm/
├── model/                      # 数据模型层 (7个文件)
│   ├── employee.go            # 员工模型
│   ├── attendance_record.go   # 考勤记录模型
│   ├── shift.go               # 班次模型
│   ├── attendance_rule.go     # 考勤规则模型
│   ├── device.go              # 考勤设备模型
│   ├── leave.go               # 请假模型
│   └── overtime.go            # 加班模型
│
├── repository/                 # Repository接口层 (6个接口)
│   ├── employee.go
│   ├── attendance_record.go
│   ├── shift.go
│   ├── device.go
│   ├── leave.go
│   └── overtime.go
│
├── repository/postgres/        # PostgreSQL实现层 (7个实现)
│   ├── hrm_employee_repo.go
│   ├── attendance_record_repo.go
│   ├── shift_repo.go
│   ├── schedule_repo.go
│   ├── attendance_rule_repo.go
│   ├── employee_sync_mapping_repo.go
│   └── employee_work_schedule_repo.go
│
├── integration/                # 第三方集成适配器
│   ├── adapter.go             # 集成接口定义
│   ├── dingtalk/              # 钉钉适配器
│   │   └── attendance_adapter.go
│   ├── wecom/                 # 企业微信适配器
│   │   └── attendance_adapter.go
│   └── feishu/                # 飞书适配器
│       └── attendance_adapter.go
│
├── service/                    # 业务逻辑层 (2个服务)
│   ├── employee_service.go
│   └── attendance_service.go
│
├── handler/                    # API处理器层 (4个Handler)
│   ├── attendance_handler.go
│   ├── shift_handler.go
│   ├── schedule_handler.go
│   └── attendance_rule_handler.go
│
├── dto/                        # 数据传输对象
│   └── employee_dto.go
│
└── README.md                   # 本文档
```

## 核心功能

### 1. 考勤管理
- ✅ 员工打卡 (支持多种方式: 设备、手机、网页、人脸识别)
- ✅ 考勤记录查询 (按员工、部门、异常情况)
- ✅ 考勤统计 (正常、迟到、早退、旷工等)
- ✅ 异常考勤处理

### 2. 班次管理
- ✅ 创建/更新/删除班次
- ✅ 支持固定班次、弹性班次、自由班次
- ✅ 配置迟到早退宽限期
- ✅ 支持跨天班次
- ✅ 加班设置

### 3. 排班管理
- ✅ 员工排班
- ✅ 批量排班
- ✅ 按日期查询排班
- ✅ 工作日类型标记 (工作日、周末、节假日)

### 4. 考勤规则
- ✅ 创建/更新/删除考勤规则
- ✅ 规则适用范围 (全员、部门、指定员工)
- ✅ 地理围栏配置
- ✅ WiFi打卡设置
- ✅ 打卡时间段配置

### 5. 第三方集成
- ✅ 钉钉考勤数据同步
- ✅ 企业微信考勤数据同步
- ✅ 飞书考勤数据同步
- ✅ 员工信息同步
- ⏳ 考勤机设备对接 (待实现)

## 技术特性

### 数据库设计
- **多租户支持**: 所有表包含 `tenant_id` 字段
- **软删除**: 使用 `deleted_at` 字段实现软删除
- **UUID主键**: 使用 UUID v7 作为主键
- **JSONB字段**: 灵活存储复杂数据(位置信息、原始数据)
- **表分区**: 考勤记录表按月分区,优化查询性能

### 性能优化
- 考勤记录表按月自动分区
- 索引优化(tenant_id, employee_id, clock_time等)
- JSONB字段支持高效查询

### 数据来源
支持多种数据来源类型:
- `system`: 系统内部
- `device`: 考勤机设备
- `dingtalk`: 钉钉
- `wecom`: 企业微信
- `feishu`: 飞书
- `manual`: 手动录入

## API接口

### 考勤服务 (AttendanceService)
- `POST /api/v1/hrm/attendance/clock-in` - 打卡
- `GET /api/v1/hrm/attendance/records/{id}` - 获取考勤记录
- `GET /api/v1/hrm/attendance/employee/{employee_id}` - 查询员工考勤
- `GET /api/v1/hrm/attendance/department/{department_id}` - 查询部门考勤
- `GET /api/v1/hrm/attendance/exceptions` - 查询异常考勤
- `GET /api/v1/hrm/attendance/statistics` - 考勤统计

### 班次服务 (ShiftService)
- `POST /api/v1/hrm/shifts` - 创建班次
- `GET /api/v1/hrm/shifts/{id}` - 获取班次
- `PUT /api/v1/hrm/shifts/{id}` - 更新班次
- `DELETE /api/v1/hrm/shifts/{id}` - 删除班次
- `GET /api/v1/hrm/shifts` - 列出班次
- `GET /api/v1/hrm/shifts/active` - 列出启用的班次

### 排班服务 (ScheduleService)
- `POST /api/v1/hrm/schedules` - 创建排班
- `POST /api/v1/hrm/schedules/batch` - 批量创建排班
- `GET /api/v1/hrm/schedules/{id}` - 获取排班
- `PUT /api/v1/hrm/schedules/{id}` - 更新排班
- `DELETE /api/v1/hrm/schedules/{id}` - 删除排班
- `GET /api/v1/hrm/schedules/employee/{employee_id}` - 查询员工排班
- `GET /api/v1/hrm/schedules/department/{department_id}` - 查询部门排班

### 考勤规则服务 (AttendanceRuleService)
- `POST /api/v1/hrm/attendance-rules` - 创建考勤规则
- `GET /api/v1/hrm/attendance-rules/{id}` - 获取考勤规则
- `PUT /api/v1/hrm/attendance-rules/{id}` - 更新考勤规则
- `DELETE /api/v1/hrm/attendance-rules/{id}` - 删除考勤规则
- `GET /api/v1/hrm/attendance-rules` - 列出考勤规则

## 数据模型

### 核心模型

#### 1. HRMEmployee (HRM员工)
```go
type HRMEmployee struct {
    ID              uuid.UUID  // UUID v7
    TenantID        uuid.UUID  // 租户ID
    EmployeeID      uuid.UUID  // 关联 organization.Employee
    EmployeeNumber  string     // 工号
    Department      string     // 部门
    Position        string     // 职位
    JobLevel        string     // 职级
    EntryDate       time.Time  // 入职日期
    ...
}
```

#### 2. AttendanceRecord (考勤记录)
```go
type AttendanceRecord struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    EmployeeID      uuid.UUID
    ClockTime       time.Time          // 打卡时间
    ClockType       AttendanceClockType // check_in, check_out
    Status          AttendanceStatus    // normal, late, early, absent
    CheckInMethod   AttendanceMethod    // device, mobile, web, face
    SourceType      SourceType          // 数据来源
    Location        *LocationInfo       // 位置信息 (JSONB)
    ...
}
```

#### 3. Shift (班次)
```go
type Shift struct {
    ID                uuid.UUID
    TenantID          uuid.UUID
    Code              string    // 班次编码
    Name              string    // 班次名称
    Type              ShiftType // fixed, flexible, free
    WorkStart         string    // 上班时间 HH:MM
    WorkEnd           string    // 下班时间 HH:MM
    LateGracePeriod   int       // 迟到宽限期(分钟)
    EarlyGracePeriod  int       // 早退宽限期(分钟)
    ...
}
```

#### 4. AttendanceRule (考勤规则)
```go
type AttendanceRule struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    Name            string
    ApplyType       string              // all, department, employee
    ApplyTo         []string            // 适用对象 (JSONB)
    CheckInSettings *CheckInSettings    // 打卡设置 (JSONB)
    Geofences       []Geofence          // 地理围栏 (JSONB)
    WiFiList        []WiFiInfo          // WiFi列表 (JSONB)
    ...
}
```

## 集成指南

### 钉钉集成
```go
import "github.com/lk2023060901/go-next-erp/internal/hrm/integration/dingtalk"

adapter := dingtalk.NewAttendanceAdapter(appKey, appSecret)
records, err := adapter.SyncAttendanceRecords(ctx, &integration.SyncAttendanceRequest{
    TenantID:   tenantID,
    EmployeeID: employeeID,
    StartDate:  startDate,
    EndDate:    endDate,
})
```

### 企业微信集成
```go
import "github.com/lk2023060901/go-next-erp/internal/hrm/integration/wecom"

adapter := wecom.NewAttendanceAdapter(corpID, corpSecret)
records, err := adapter.SyncAttendanceRecords(ctx, req)
```

### 飞书集成
```go
import "github.com/lk2023060901/go-next-erp/internal/hrm/integration/feishu"

adapter := feishu.NewAttendanceAdapter(appID, appSecret)
records, err := adapter.SyncAttendanceRecords(ctx, req)
```

## 数据库迁移

### 迁移脚本位置
```
migrations/
├── 008_create_hrm_tables.sql         # HRM基础表
├── 009_create_hrm_tables_part2.sql   # HRM扩展表(第2部分)
└── 010_create_hrm_tables_part3.sql   # HRM扩展表(第3部分)
```

### 表分区创建
考勤记录表按月自动分区,需定期创建新分区:
```sql
-- 创建2024年1月分区
CREATE TABLE hrm_attendance_records_2024_01 PARTITION OF hrm_attendance_records
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## 统计信息

### 代码量统计
- **总代码行数**: 6,550+ 行
- **Model层**: 1,032 行 (7个文件)
- **Repository接口**: 677 行 (6个文件)
- **Repository实现**: 2,078 行 (7个文件)
- **Service层**: 1,026 行 (2个文件)
- **Handler层**: 268 行 (4个文件)
- **Integration层**: 885 行 (4个文件)
- **DTO层**: 180 行 (1个文件)
- **数据库迁移**: 1,080 行 (3个SQL文件)
- **Protobuf定义**: 568 行 (1个proto文件)

### 功能实现进度
- ✅ Model层 (100%)
- ✅ Repository接口层 (100%)
- ✅ Repository PostgreSQL实现 (100%)
- ✅ Service层 (100%)
- ✅ Handler层 (100%)
- ✅ 第三方集成适配器 (100% - 钉钉/企业微信/飞书)
- ✅ API Protobuf定义 (100%)
- ✅ 数据库迁移脚本 (100%)
- ⏳ Wire依赖注入 (待完成)
- ⏳ 单元测试 (待完成)
- ⏳ 集成测试 (待完成)
- ⏳ 考勤机设备适配器 (待完成)

## 设计原则

1. **独立性**: HRM模块独立设计,便于日后拆分为微服务
2. **可扩展性**: 通过接口和适配器模式支持多种第三方平台
3. **多租户**: 所有数据隔离,支持SaaS模式
4. **性能优化**: 表分区、索引优化、JSONB字段
5. **数据一致性**: 事务处理、软删除、审计字段

## 下一步工作

1. ✅ 完善Handler层实现 (填充TODO部分)
2. ✅ 创建Wire依赖注入配置
3. ⏳ 编写单元测试
4. ⏳ 编写集成测试
5. ⏳ 实现考勤机设备适配器
6. ⏳ 性能测试和优化
7. ⏳ 完善API文档

## 参考资料

- [Kratos框架文档](https://go-kratos.dev/)
- [Google Wire](https://github.com/google/wire)
- [钉钉开放平台](https://open.dingtalk.com/)
- [企业微信API](https://work.weixin.qq.com/api/doc)
- [飞书开放平台](https://open.feishu.cn/)
