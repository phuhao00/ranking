# Middleware Test Script

Write-Host "Starting middleware tests..." -ForegroundColor Green
Write-Host ""

# Test 1: Health check
Write-Host "1. Testing health endpoint" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "Success: Health check passed" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json)" -ForegroundColor Cyan
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Basic endpoint
Write-Host "2. Testing basic endpoint" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/test" -Method GET
    Write-Host "Success: Basic test passed" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json)" -ForegroundColor Cyan
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 3: Rate limiting
Write-Host "3. Testing rate limiting (rapid requests)" -ForegroundColor Yellow
$successCount = 0
$rateLimitCount = 0

for ($i = 1; $i -le 15; $i++) {
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/rate-limit-test" -Method GET -ErrorAction Stop
        $successCount++
        Write-Host "Request $i : Success" -ForegroundColor Green
    } catch {
        if ($_.Exception.Response.StatusCode -eq 429) {
            $rateLimitCount++
            Write-Host "Request $i : Rate limited (429)" -ForegroundColor Yellow
        } else {
            Write-Host "Request $i : Error $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 50
}

Write-Host "Successful requests: $successCount, Rate limited: $rateLimitCount" -ForegroundColor Cyan
Write-Host ""

# Test 4: Recovery middleware
Write-Host "4. Testing recovery middleware (panic endpoint)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/panic-test" -Method GET
    Write-Host "Failed: Should have returned 500 error" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 500) {
        Write-Host "Success: Recovery middleware caught panic and returned 500" -ForegroundColor Green
    } else {
        Write-Host "Failed: Status code $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "Middleware tests completed!" -ForegroundColor Green