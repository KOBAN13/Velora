$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$ProtoFile = Join-Path $Root "shared\packets.proto"
$IncludeDir = Join-Path $Root "shared"
$OutDir = Join-Path $Root "server"
$ProtocGenGoVersion = "v1.36.11"

function Find-Protoc {
    $command = Get-Command protoc -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    $wingetGlob = Join-Path $env:LOCALAPPDATA "Microsoft\WinGet\Packages\Google.Protobuf*\bin\protoc.exe"
    $candidates = Get-ChildItem -Path $wingetGlob -ErrorAction SilentlyContinue
    if ($candidates) {
        return $candidates[0].FullName
    }

    throw "protoc not found. Install: winget install --id Google.Protobuf -e"
}

function Ensure-ProtocGenGo {
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) {
        throw "go not found in PATH"
    }

    $binDir = Join-Path (go env GOPATH) "bin"
    $plugin = Join-Path $binDir "protoc-gen-go.exe"
    if (-not (Test-Path $plugin)) {
        Write-Host "Installing protoc-gen-go $ProtocGenGoVersion..."
        go install "google.golang.org/protobuf/cmd/protoc-gen-go@$ProtocGenGoVersion"
    }

    return $plugin
}

$protoc = Find-Protoc
$protocGenGo = Ensure-ProtocGenGo

Write-Host "protoc: $protoc"
Write-Host "protoc-gen-go: $protocGenGo"

& $protoc `
    "-I=$IncludeDir" `
    "--plugin=protoc-gen-go=$protocGenGo" `
    "--go_out=$OutDir" `
    $ProtoFile

Write-Host "Done: server\pkg\packets\packets.pb.go"
