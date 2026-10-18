package config

import (
	"bufio"   // 按行读取文件的缓冲 IO
	"errors"  // 创建错误对象
	"flag"    // 命令行参数解析
	"os"      // 文件操作、环境变量
	"strconv" // 字符串与数字互转
	"strings" // 字符串处理工具
)

// 各配置项的兜底默认值：命令行参数、环境变量、.env 都没提供时使用
const (
	defaultServerPort    = "8123"           // 服务监听端口
	defaultContextPath   = "/api"           // 接口统一前缀
	defaultRedisAddr     = "127.0.0.1:6379" // Redis 地址
	defaultRedisDB       = 0                // Redis 库编号（0~15）
	defaultSessionMaxAge = 2592000          // session 有效期（秒）= 30 天
	defaultSessionName   = "SESSION"        // Cookie 里 session id 的字段名
)

// 配置加载 负责从 .env 文件和命令行参数中读取配置
// 配置加载核心逻辑：命令行参数 > 环境变量 > .env 文件 > 默认值  优先级从高到低
type Config struct {
	ServerPort    string // 服务监听端口          ← SERVER_PORT
	ContextPath   string // 接口统一前缀          ← SERVER_CONTEXT_PATH
	MySQLDSN      string // MySQL 连接串（必填）  ← MYSQL_DSN
	RedisAddr     string // Redis 地址            ← REDIS_ADDR
	RedisUsername string // Redis 6.0+ ACL 用户名，留空 = 默认用户 ← REDIS_USERNAME
	RedisPassword string // Redis 密码，留空 = 无密码             ← REDIS_PASSWORD
	RedisDB       int    // Redis 库编号 0~15     ← REDIS_DB

	SessionName         string // Cookie 中 session id 的字段名    ← SESSION_NAME
	SessionSecret       string // session 签名密钥（必填）         ← SESSION_SECRET
	SessionMaxAge       int    // session 有效期（秒）             ← SESSION_MAX_AGE_SECONDS
	EncryptionSecretKey string // BYOK：加密用户 AI API Key 的密钥 ← ENCRYPTION_SECRET_KEY
}

// Load 配置加载总入口：读 .env → 解析命令行参数 → 按优先级组装 Config。
// 返回值 (*Config, error)：Go 不用异常机制，错误作为普通返回值交还调用方处理
func Load() (*Config, error) {
	// 1. 读取 .env 文件中的环境变量
	baseEnv := currentEnvKeys()    // 进程已有环境变量键名快照，防止稍后被 .env 覆盖
	envName := detectEnvName()     // 当前环境名，如 "dev"
	loadEnvFiles(envName, baseEnv) // 依次加载 .env、.env.dev 并注入进程环境变量

	// 2. 解析命令行参数
	// flag.String(参数名, 默认值, 说明)：注册一个 string 型命令行参数，返回 *string。
	// 返回指针是因为注册时值还未确定，flag.Parse() 之后才被填充，
	// 之后通过 *serverPortFlag 解引用读到的才是解析后的值
	serverPortFlag := flag.String("server-port", "", "server port")
	contextPathFlag := flag.String("server-context-path", "", "server context path")
	envFlag := flag.String("env", "", "runtime env name, such as dev or prod")
	mysqlDSNFlag := flag.String("mysql-dsn", "", "mysql dsn")
	redisAddrFlag := flag.String("redis-addr", "", "redis address")
	redisUsernameFlag := flag.String("redis-username", "", "redis username")
	redisPasswordFlag := flag.String("redis-password", "", "redis password")
	// flag.Int：int 版本。默认值用 -1 当"未设置"哨兵——不能用 0，
	// 因为 RedisDB=0（0 号库）是合法值
	redisDBFlag := flag.Int("redis-db", -1, "redis db index")
	sessionNameFlag := flag.String("session-name", "", "session cookie name")
	sessionSecretFlag := flag.String("session-secret", "", "session secret")
	sessionMaxAgeFlag := flag.Int("session-max-age-seconds", -1, "session max age in seconds")

	encryptionSecretKeyFlag := flag.String("encryption-secret-key", "", "secret key for BYOK api key encryption")

	// flag.Parse()：真正扫描 os.Args（相当于 Java 的 main(String[] args)），
	// 把值填进上面注册的所有指针。必须在读取 *flag 之前调用
	flag.Parse()
	if *envFlag != "" && *envFlag != envName {
		// 命令行 --env 与 APP_ENV 不一致时，以命令行为准重新加载对应的 .env.<env>
		loadEnvFiles(*envFlag, baseEnv)
	}

	// pickString(命令行值, 环境变量值, 默认值)：三级优先取值，取第一个非空
	// os.Getenv("MYSQL_DSN")：读进程环境变量（相当于 Java 的 System.getenv()）
	mysqlDSN := pickString(*mysqlDSNFlag, os.Getenv("MYSQL_DSN"), "")
	if mysqlDSN == "" {
		// errors.New：创建错误对象。Go 没有 throw，错误直接作为返回值上抛
		return nil, errors.New("MYSQL_DSN is required (or pass --mysql-dsn)")
	}
	sessionSecret := pickString(*sessionSecretFlag, os.Getenv("SESSION_SECRET"), "")
	if sessionSecret == "" {
		return nil, errors.New("SESSION_SECRET is required (or pass --session-secret)")
	}

	// 3. 按优先级组装配置；&Config{...} 取结构体指针返回
	return &Config{
		ServerPort:    pickString(*serverPortFlag, os.Getenv("SERVER_PORT"), "8123"),
		ContextPath:   pickString(*contextPathFlag, os.Getenv("SERVER_CONTEXT_PATH"), "/api"),
		MySQLDSN:      mysqlDSN,
		RedisAddr:     pickString(*redisAddrFlag, os.Getenv("REDIS_ADDR"), defaultRedisAddr),
		RedisUsername: pickString(*redisUsernameFlag, os.Getenv("REDIS_USERNAME"), ""),
		RedisPassword: pickString(*redisPasswordFlag, os.Getenv("REDIS_PASSWORD"), ""),
		// pickInt：int 版三级优先，>= 0 视为有效值（-1 = 未设置）
		RedisDB:             pickInt(*redisDBFlag, getIntEnvOrDefault("REDIS_DB", defaultRedisDB), defaultRedisDB),
		SessionName:         pickString(*sessionNameFlag, os.Getenv("SESSION_NAME"), defaultSessionName),
		SessionSecret:       sessionSecret,
		SessionMaxAge:       pickInt(*sessionMaxAgeFlag, getIntEnvOrDefault("SESSION_MAX_AGE_SECONDS", defaultSessionMaxAge), defaultSessionMaxAge),
		EncryptionSecretKey: pickString(*encryptionSecretKeyFlag, os.Getenv("ENCRYPTION_SECRET_KEY"), ""),
	}, nil
}

// loadEnvFiles 按顺序加载 .env 和 .env.<环境名>（相当于 Spring 的 profile 叠加：
// application.yml + application-dev.yml），后加载的文件可覆盖先加载的同名键，
// 但都不能覆盖真实的进程环境变量
func loadEnvFiles(envName string, baseEnv map[string]struct{}) {
	paths := []string{
		".env",
		".env." + envName, // envName="dev" → ".env.dev"
	}
	for _, path := range paths { // for range：遍历切片，相当于 Java 的 for-each
		loadEnvFile(path, baseEnv)
	}
}

// loadEnvFile 手写的 .env 解析器：逐行读取 KEY=VALUE 注入进程环境变量；
// 文件不存在时静默跳过（.env 属于可选配置）
func loadEnvFile(path string, baseEnv map[string]struct{}) {
	// os.Open：只读打开文件，返回 (*File, error)。
	// Java 靠抛 FileNotFoundException，Go 把错误当普通返回值，需当场判断处理
	file, err := os.Open(path)
	if err != nil {
		return // 文件不存在就跳过（.env 是可选的）
	}
	// defer：函数返回时（无论走哪个分支）执行关闭，相当于 Java 的 try-with-resources
	defer file.Close()

	//  Java 的 BufferedReader.readLine() 循环，一行一行读文件
	// bufio.NewScanner：按行扫描；scanner.Scan() 读到新行返回 true、读完返回 false；
	// scanner.Text() 取当前行内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// strings.TrimSpace：去掉两端空白，相当于 Java 的 String.trim()
		// 例："  SERVER_PORT=8123  " → "SERVER_PORT=8123"
		line := strings.TrimSpace(scanner.Text())
		// strings.HasPrefix：判断前缀，相当于 Java 的 startsWith
		if line == "" || strings.HasPrefix(line, "#") {
			continue // 跳过空行和注释行
		}
		// strings.SplitN(line, "=", 2)：在第一个 "=" 处切一刀、最多切成 2 段。
		// 例："MYSQL_DSN=root:pwd@tcp(127.0.0.1:3306)/db?parseTime=true"
		//       → parts[0] = "MYSQL_DSN"
		//         parts[1] = "root:pwd@tcp(127.0.0.1:3306)/db?parseTime=true"
		// 值里再含 "="（如 parseTime=true）也不会被切碎；
		// Java 的 split("=") 是全切、会把值截断，所以这里必须用 SplitN
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 不含 "=" 的行直接跳过
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// strings.Trim(value, `"'`)：剥掉两端成对出现的双引号或单引号。
		// 反引号 `"'` 是原始字符串字面量，内容就是双引号、单引号两个字符。
		// 例：value 为 "123456"（两端带双引号）→ Trim 后为 123456
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		// Do not override real process-level env,
		// but allow later .env files to override earlier .env files.
		// 例：baseEnv 中已存在 "REDIS_ADDR"（部署时真实设置过）→ 跳过，不覆盖
		if _, exists := baseEnv[key]; exists {
			continue
		}
		// os.Setenv：写入进程环境变量，之后 os.Getenv 即可读到
		// _ = ：空白标识符，显式丢弃 Setenv 返回的 error（此处失败不影响主流程）
		_ = os.Setenv(key, value)
	}
}

// detectEnvName 识别运行环境名：读环境变量 APP_ENV，未设置时默认 "dev"
func detectEnvName() string {
	envName := strings.TrimSpace(os.Getenv("APP_ENV"))
	if envName == "" {
		return "dev"
	}
	return envName
}

// currentEnvKeys 收集进程当前全部环境变量的键名，做成一个 Set。
// Go 没有 Set 类型，惯用 map[键]struct{} 模拟——struct{} 是零字节空类型，
// 只借 map 的键、不占用 value 空间
func currentEnvKeys() map[string]struct{} {
	result := make(map[string]struct{}) // make：创建并初始化 map
	for _, item := range os.Environ() { // os.Environ：返回所有环境变量，每项形如 "KEY=value"
		parts := strings.SplitN(item, "=", 2) // 同上：只在第一个 "=" 处切，防止值被截断
		if len(parts) == 0 {
			continue
		}
		result[parts[0]] = struct{}{} // 只存键，值用零字节占位
	}
	return result
}

// pickString 三级优先取 string 值：命令行参数 → 环境变量 → 默认值，取第一个非空。
// 例：pickString("", "8123", "9000") → "8123"（命令行没传，落到环境变量）
func pickString(flagValue, envValue, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if envValue != "" {
		return envValue
	}
	return defaultValue
}

// pickInt pickString 的 int 版本。用 >= 0 而不是 != 0 判断"是否设置"：
// 因为 0 是合法值（RedisDB=0 即 0 号库），哨兵只能取 -1。
// 例：pickInt(-1, 2, 0) → 2；pickInt(-1, -1, 0) → 0
func pickInt(flagValue, envValue, defaultValue int) int {
	if flagValue >= 0 {
		return flagValue
	}
	if envValue >= 0 {
		return envValue
	}
	return defaultValue
}

// getIntEnvOrDefault 读环境变量并转成 int；未设置或格式非法时返回默认值。
// 例：REDIS_DB="3" → 3；REDIS_DB="abc" → 返回默认值
func getIntEnvOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// strconv.Atoi：字符串转 int，相当于 Java 的 Integer.parseInt；
	// 转换失败不抛异常，而是返回 (0, error) 由调用方判断
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// getBoolEnvOrDefault 读环境变量并转成 bool：识别 1/true/yes/y 与 0/false/no/n，
// 其余情况返回默认值。（当前暂无调用方，为将来的布尔开关配置预留）
func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	// strings.ToLower：转小写，相当于 Java 的 toLowerCase
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return defaultValue
	}
	if value == "1" || value == "true" || value == "yes" || value == "y" {
		return true
	}
	if value == "0" || value == "false" || value == "no" || value == "n" {
		return false
	}
	return defaultValue
}
