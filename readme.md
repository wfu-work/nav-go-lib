# nav-go-lib

`nav-go-lib` 是一个用于合并 GNSS 广播星历的 Go 库。它从 RTCM3 原始数据文件中读取导航电文，解析多系统星历，去重排序后输出 RINEX 3.02 mixed NAV 文件；同时也可以构建成 C 共享库，供其他语言或桌面程序调用。

## 功能概览

- 读取 RTCM3 帧并校验 CRC24Q。
- 解析广播星历消息：
  - GPS: RTCM 1019
  - GLONASS: RTCM 1020
  - BDS: RTCM 1042
  - Galileo: RTCM 1045, 1046
- 按时间范围、站点前缀和文件后缀查找 RTCM 文件。
- 按卫星和 Toe 去重，按 Toe 排序。
- 追加写入 RINEX 3.02 `N: GNSS NAV DATA` mixed 导航文件，并跳过已存在的星历记录。
- 导出 `mergeNav` C ABI 方法，可通过 `go build -buildmode=c-shared` 生成动态库。

## 项目结构

```text
.
├── main.go            # c-shared 动态库导出函数 mergeNav / FreeC
├── cmd/               # 命令行调试入口
├── libs/              # RTCM 读取、消息分发、星历解析、RINEX 写入、合并主流程
├── domains/           # Nav、Ephe、GTime、SatType 等领域模型
├── utils/             # 文件检索、时间系统转换、CRC24Q、RINEX 辅助值等工具
├── makefile           # 多平台动态库构建命令
└── build/             # 默认构建产物目录
```

## 环境要求

- Go 1.24 或以上。
- 构建 C 共享库时需要启用 CGO。
- 交叉编译 Windows/Linux 动态库时，需要本机安装对应的 C 交叉编译器，例如 `x86_64-w64-mingw32-gcc` 或 `x86_64-linux-gnu-gcc`。

## 快速开始

### 运行示例

```bash
go run ./cmd "2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@PSNMJC014,PSNMJC074@1@"
```

`cmd` 是开发调试和人工验证入口，接收与 `libs.Nav` 相同的 `@` 拼接参数串。

### 查看帮助

```bash
go run ./cmd --help
```

### 构建动态库

```bash
# 按当前系统构建
make

# 指定平台构建
make build-macos
make build-linux
make build-windows

# 清理构建产物
make clean
```

默认产物：

| 平台 | 文件 |
| --- | --- |
| macOS | `build/libNavMerge.dylib` |
| Linux | `build/libNavMerge.so` |
| Windows | `build/libNavMerge.dll` |

## 参数格式

核心入口 `libs.Nav(argv string)` 和导出的 `mergeNav(char*)` 使用同一套字符串参数：

```text
startTime@endTime@rawPath@outPath@sites@fileFormat@ext
```

示例：

```text
2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@PSNMJC014,PSNMJC074@1@
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `startTime` | 开始时间，格式为 `YYYY/MM/DD HH:mm:ss` |
| `endTime` | 结束时间，格式为 `YYYY/MM/DD HH:mm:ss` |
| `rawPath` | RTCM 原始数据根目录 |
| `outPath` | RINEX NAV 输出根目录 |
| `sites` | 站点文件名前缀，多个站点用英文逗号分隔 |
| `fileFormat` | 原始数据目录结构类型，见下表 |
| `ext` | 文件后缀过滤条件；为空时使用默认后缀规则 |

`fileFormat` 支持：

| 值 | 目录结构 | 扫描步长 |
| --- | --- | --- |
| `0` | `rawPath` | 按天推进 |
| `1` | `rawPath/YYYY/DDD/HH` | 按小时推进 |
| `2` | `rawPath/YYYY/DDD` | 按天推进 |

默认后缀规则：

| `fileFormat` | 默认后缀 |
| --- | --- |
| `0` | `.YYYYDDDbinRTCM3` |
| `1` | `.YYYYDDDbinRTCM3` |
| `2` | `.YYYYo` |

其中 `YYYY` 是四位年份，`DDD` 是年积日。

## 输出规则

`libs.Nav` 会将结果写入：

```text
<outPath>/<YYYY>/BRDM<DDD>0.rnx
```

例如结束时间为 `2025/12/07 09:00:00` 时，输出文件为：

```text
<outPath>/2025/BRDM3410.rnx
```

函数返回输出文件路径；如果没有解析到有效星历，或写入失败，则返回空字符串。

## Go 调用示例

```go
package main

import (
	"fmt"

	"navfirst.com/nav-go-lib/libs"
)

func main() {
	argv := "2025/12/07 00:00:00@2025/12/07 09:00:00@/data/raw@/data/nav@PSNMJC014,PSNMJC074@1@"
	out := libs.Nav(argv)
	fmt.Println(out)
}
```

如果只需要内存中的星历数据，可以使用：

```go
nav := libs.NavData(argv)
```

## C 共享库接口

`main.go` 导出了：

```c
char* mergeNav(char* argv);
void FreeC(char* s);
```

构建动态库后，可以从 C/C++、Java、Python、C# 等语言通过 FFI 调用。`mergeNav` 返回值由 Go 内部使用 `C.CString` 分配，调用方在使用完后应调用 `FreeC` 释放。

## 主要流程

1. 解析 `argv` 参数，得到时间范围、原始文件目录、输出目录、站点列表和文件格式。
2. 根据 `fileFormat` 扫描匹配的 RTCM 文件。
3. 逐帧读取 RTCM3 数据，校验 CRC24Q。
4. 按消息类型分发到对应解析器。
5. 将解析出的星历写入 `domains.Nav`，按卫星和 Toe 去重。
6. 按 Toe 排序后追加写入 RINEX 3.02 mixed NAV 文件。

## 注意事项

- 当前库聚焦广播星历合并，不解析观测值数据。
- 时间字符串按 `YYYY/MM/DD HH:mm:ss` 解析。
- `sites` 使用文件名前缀匹配，例如 `PSNMJC014` 会匹配以该字符串开头的文件。
- 输出文件采用追加写入模式；如果文件已存在，会读取已有记录并跳过重复星历。
- 动态库构建依赖 CGO，交叉编译环境需要提前准备对应编译器。
