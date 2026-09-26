# 2026/09/26

## Go 后端项目目录结构分析 & Spring Boot 对照（cc-ai-code-mother）

> 结论先行：这套 Go 目录结构就是 Spring Boot 分层架构的"Go 翻译版"（连 `HealthController`、`BusinessException`、`UserConstant` 的命名都一一对应）。
> Spring Boot 的后端思维可以完全平移，只需要学"Go 的表达方式"。

### 一、Go 项目目录总览

```text
go-backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口（总装车间）
├── internal/                    # 内部包（编译器强制：外部模块无法 import）
│   ├── common/                  # 通用工具
│   │   ├── response.go          # 统一响应 {code, data, message}
│   │   └── page.go              # 分页封装
│   ├── config/                  # 配置管理（加载 .env）
│   │   └── config.go
│   ├── constant/                # 常量定义
│   │   └── user_constant.go
│   ├── controller/              # 控制器层（处理 HTTP 请求）
│   │   ├── health_controller.go
│   │   └── user_controller.go
│   ├── errno/                   # 错误码和业务异常
│   │   ├── error_code.go
│   │   └── business_error.go
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # 登录态/权限校验
│   │   ├── cors.go              # 跨域
│   │   └── recovery.go          # 捕获 panic，防止进程崩溃
│   ├── model/                   # 数据模型
│   │   ├── dto/                 # 请求参数
│   │   │   └── user.go
│   │   ├── entity/              # 数据库实体
│   │   │   └── user.go
│   │   └── vo/                  # 响应视图（脱敏）
│   │       └── user.go
│   ├── repository/              # 数据访问层
│   │   └── user_repository.go
│   ├── router/                  # 路由配置（集中注册所有路由）
│   │   └── router.go
│   └── service/                 # 业务逻辑层
│       └── user_service.go
├── .env                         # 环境变量配置（含密码，必须进 .gitignore）
├── go.mod                       # Go Module 文件（模块定义 + 依赖清单）
└── go.sum                       # 依赖校验文件（密码学校验和，自动生成勿手改）
```

### 二、与 Spring Boot 项目的逐层对照

对照对象：`cc-ai-code-mother`（MyBatis-Plus 分层架构）

| Go 目录/文件 | Spring Boot 对应物 | 干的事 |
|---|---|---|
| `cmd/server/main.go` | `CcAiCodeMotherApplication.java` | 程序入口 |
| `internal/` | `src/main/java/com/zjcc/...` 整个包树 | 项目代码根 |
| `common/response.go` | `BaseResponse.java` + `ResultUtils.java` | 统一响应 `{code, data, message}` |
| `common/page.go` | `PageRequest.java` | 分页封装 |
| `config/config.go` | `application.yml` + `config/` 下的 `@Configuration` 类 | 读取并集中管理配置 |
| `constant/user_constant.go` | `UserConstant.java` | 常量（如用户角色） |
| `controller/user_controller.go` | `UserController.java`（`@RestController`） | 处理 HTTP 请求、参数校验 |
| `errno/error_code.go` | `BusinessException` 携带的 errorCode | 错误码定义 |
| `errno/business_error.go` | `BusinessException.java` | 业务异常对象 |
| `middleware/auth.go` | `@AuthCheck` 注解 + `aop/AuthInterceptor.java` | 登录态/权限校验 |
| `middleware/cors.go` | `CorsConfig.java` | 跨域配置 |
| `middleware/recovery.go` | `GlobalExceptionHandler.java` 的兜底职责 | 捕获 panic 防崩溃 |
| `model/dto/user.go` | `model/dto/user/*Request.java` | 请求参数对象 |
| `model/entity/user.go` | `model/entity/User.java` | 数据库实体 |
| `model/vo/user.go` | `UserVO.java` / `LoginUserVO.java` | 响应视图（脱敏） |
| `repository/user_repository.go` | `mapper/UserMapper.java`（MyBatis-Plus） | 数据访问层 |
| `service/user_service.go` | `UserService.java` + `impl/UserServiceImpl.java` | 业务逻辑 |
| `router/router.go` | 各 Controller 上的 `@RequestMapping` 注解 | **集中**注册所有路由 |
| `.env` | `application.yml` + `application-local.yml` | 环境变量配置 |
| `go.mod` | `pom.xml` | 模块定义 + 依赖清单 |
| `go.sum` | （无对应物，类似 `package-lock.json`） | 依赖的密码学校验和 |

注：Spring 项目里的 `ThrowUtils.java` 在 Go 中无对应物——Go 用 `if err != nil` 显式处理错误。

### 三、四个核心思维差异

#### 1. `internal/` 是编译器强制的私有（Java 没有这东西）

Java 的包私有（不加 `public`）只对同包生效，跨包靠自觉。
Go 规定：**任何外部模块 import 你项目 `internal/` 下的包，直接编译报错**。
含义："这些都是实现细节，外人别碰"。看开源项目时，`internal/` 越大说明私有实现越多。

#### 2. 没有注解和魔法——一切显式

Spring：`@RestController`、`@RequestMapping("/user")`、`@Autowired`，框架扫描注解自动装配。
Go **没有注解**，全部是明明白白的代码：

- 路由不是散在各 Controller 的注解，而是集中在 `router.go` 里逐行注册：
  `userGroup.POST("/login", userController.Login)`
- 依赖注入不是 `@Autowired`，而是在 `main.go` / `router.go` 里手动构造、逐层传入

代价：写得多。收益：跳转定义就能看清一切，没有"这个 Bean 是哪来的"的疑惑。

#### 3. `main.go` 要自己"组装汽车"

Spring 的启动类几乎是空的（框架接管一切）；Go 的 `main.go` 是**总装车间**：

```go
func main() {
    config.Init()                                // 加载 .env
    db := database.Init()                        // 连数据库
    repo := repository.NewUserRepo(db)
    svc := service.NewUserService(repo)
    ctrl := controller.NewUserController(svc)
    r := router.Setup(ctrl, middleware.Auth())   // 注册路由 + 中间件
    r.Run(":8080")                               // 启动 HTTP 服务
}
```

依赖链条（repo → service → controller）亲手串起来，一眼看清应用骨架。

#### 4. 错误处理：`error` 是值，不是异常（最大的思维转换）

Java：`throw new BusinessException(...)` 靠 `GlobalExceptionHandler` 兜住。
Go：**error 是普通返回值**，逐层手动传递：

```go
user, err := userService.GetUser(id)
if err != nil {
    return nil, err    // 显式把错误交还给调用方
}
```

- `errno/business_error.go` ≈ `BusinessException`（带错误码的错误对象），但不会被"抛出"，而是作为返回值一路传到 controller
- `middleware/recovery.go` 只兜底 **panic**（相当于 Java 的 Error 级灾难，如空指针），不负责业务错误

### 四、其他对应关系细节

- **`repository` vs `mapper`**：职责相同。MyBatis-Plus 用 `BaseMapper<User>` + 条件构造器；Go 教程标配 **GORM**（`db.Where("id = ?", id).First(&user)` 链式调用），零 SQL、纯代码，体验很像 MyBatis-Plus 的 Wrapper
- **`.env` vs `application.yml`**：Go 没有多 profile 机制（没有 `spring.profiles.active` 那套），惯例是环境变量 + `.env` 区分环境；⚠️ `.env` 含密码必须进 `.gitignore`
- **枚举**：Go 没有枚举类型，用 `constant/` 定义常量 + 自定义类型模拟（对应 `UserRoleEnum.java`）
- **`health_controller.go`**：与 `HealthController.java` 相同，`GET /health` 探活
- **框架**：这套结构（router / middleware / controller 签名）几乎必然搭配 **Gin** 框架

### 五、建议的阅读顺序

按**一次 HTTP 请求的完整路径**读，比按目录读有效：

```text
main.go → router.go → middleware（请求先过中间件）→
user_controller.go（取参数）→ user_service.go（业务逻辑）→
user_repository.go（查库）→ 一路 return → response.go（包装成统一响应）
```

跟着一条 `/user/login` 请求走一遍就通了——本质上和 Spring Boot 里走过的同款路径一样，只是换了语言表达。
