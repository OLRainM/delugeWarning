# setup_db.ps1 —— 本地创建 deluge 数据库（建表由后端启动时自动迁移完成）
#
# 用法（在 backend 目录下运行）：
#   .\setup_db.ps1 -Password "你的postgres密码"
# 可选参数：-DbUser postgres -DbName deluge -DbHost 127.0.0.1 -Port 5432

param(
    [Parameter(Mandatory = $true)]
    [string]$Password,
    [string]$DbUser = "postgres",
    [string]$DbName = "deluge",
    [string]$DbHost = "127.0.0.1",
    [int]$Port = 5432
)

$ErrorActionPreference = "Stop"

# 自动定位 psql（兼容多版本安装）
$psql = Get-ChildItem 'C:\Program Files\PostgreSQL\*\bin\psql.exe' -ErrorAction SilentlyContinue |
    Sort-Object FullName -Descending | Select-Object -First 1 -ExpandProperty FullName
if (-not $psql) {
    Write-Error "未找到 psql.exe，请确认已安装 PostgreSQL"
    exit 1
}
Write-Host "使用 psql: $psql"

$env:PGPASSWORD = $Password
$env:PGCLIENTENCODING = "UTF8"

# 检测数据库是否存在
$exists = & $psql -U $DbUser -h $DbHost -p $Port -d postgres -tAc `
    "SELECT 1 FROM pg_database WHERE datname='$DbName'"

if ($LASTEXITCODE -ne 0) {
    Write-Error "连接数据库失败，请检查用户名/密码/端口"
    exit 1
}

if ($exists -eq "1") {
    Write-Host "数据库 '$DbName' 已存在，无需创建"
} else {
    & $psql -U $DbUser -h $DbHost -p $Port -d postgres -c "CREATE DATABASE $DbName"
    if ($LASTEXITCODE -eq 0) {
        Write-Host "已创建数据库 '$DbName'"
    } else {
        Write-Error "创建数据库失败"
        exit 1
    }
}

Write-Host ""
Write-Host "完成。接下来："
Write-Host "  1) 确认 config.yaml 的 database.dsn 中 user/password/dbname 与上面一致"
Write-Host "  2) 运行后端：go run ./cmd/server （启动时会自动建表）"
