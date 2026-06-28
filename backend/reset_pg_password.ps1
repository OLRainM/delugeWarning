# reset_pg_password.ps1 (run as Administrator)
# Temporarily set pg_hba.conf to trust, reset postgres password, then restore.
$ErrorActionPreference = "Stop"
$log = "$PSScriptRoot\reset_log.txt"
"START $(Get-Date)" | Out-File $log -Encoding utf8

try {
    $svc = "postgresql-x64-18"
    $hba = "C:\Program Files\PostgreSQL\18\data\pg_hba.conf"
    $bak = "$hba.bak_reset"
    $newPassword = "postgres"

    # 1) backup
    Copy-Item $hba $bak -Force
    "Backed up to $bak" | Out-File $log -Append -Encoding utf8

    # 2) rewrite method -> trust for active lines
    $lines = Get-Content $hba
    $patched = foreach ($ln in $lines) {
        $t = $ln.Trim()
        if ($t -eq "" -or $t.StartsWith("#")) { $ln; continue }
        $parts = $ln -split '\s+'
        if ($parts.Length -ge 2) {
            $parts[$parts.Length - 1] = "trust"
            ($parts -join " ")
        } else { $ln }
    }
    $patched | Set-Content $hba -Encoding ascii
    "pg_hba patched to trust" | Out-File $log -Append -Encoding utf8

    # 3) restart + reset password
    Restart-Service $svc -Force
    Start-Sleep -Seconds 3
    $psql = "C:\Program Files\PostgreSQL\18\bin\psql.exe"
    & $psql -U postgres -h 127.0.0.1 -p 5432 -d postgres -c "ALTER USER postgres PASSWORD '$newPassword'" 2>&1 | Out-File $log -Append -Encoding utf8
    "Password reset attempted (exit=$LASTEXITCODE)" | Out-File $log -Append -Encoding utf8

    # 4) restore original hba + restart
    Copy-Item $bak $hba -Force
    Restart-Service $svc -Force
    Start-Sleep -Seconds 3
    "Restored pg_hba and restarted" | Out-File $log -Append -Encoding utf8
    "DONE OK" | Out-File $log -Append -Encoding utf8
}
catch {
    "ERROR: $_" | Out-File $log -Append -Encoding utf8
}
