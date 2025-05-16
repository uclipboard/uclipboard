
$ucipboardDirectory = "" 


Add-Type -AssemblyName System.Windows.Forms

$balloon = New-Object System.Windows.Forms.NotifyIcon
$path = (Get-Process -id $pid).Path
$icon = [System.Drawing.Icon]::ExtractAssociatedIcon($path)
$balloon.Icon =$icon
$balloon.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
$balloon.BalloonTipTitle = "Uclipboard Startup"
$balloon.Visible =$true

if ($ucipboardDirectory -eq "") {
    $balloon.BalloonTipText = "Please set uclipboard directory"
    $balloon.ShowBalloonTip(10000)
    exit
}

$uclipboardExecutablePath = Join-Path -Path $ucipboardDirectory -ChildPath "uclipboard.exe"
$uclipboardLogPath = Join-Path -Path $ucipboardDirectory -ChildPath "uclipboard.log"
$uclipboardArguments = "--mode client --log $uclipboardLogPath --log-level info"


if (-not (Get-Process "uclipboard" -ErrorAction SilentlyContinue)) {
    Start-Process $uclipboardExecutablePath -ArgumentList $uclipboardArguments -WindowStyle Hidden
    $balloon.BalloonTipText = "uclipboard has been started"
} else {
    $balloon.BalloonTipText = "uclipboard is already running"
}


# show the balloon tip for 10 seconds
$balloon.ShowBalloonTip(10000)