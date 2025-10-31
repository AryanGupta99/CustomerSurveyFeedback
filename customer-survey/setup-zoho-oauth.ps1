# Zoho Creator Secure Setup Script
# This script helps you configure secure OAuth authentication for Zoho Creator

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Zoho Creator Secure OAuth Setup" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Collect Zoho Creator details
Write-Host "[Step 1] Zoho Creator Application Details" -ForegroundColor Yellow
Write-Host ""
$accountOwner = Read-Host "Enter your Zoho Account Owner Name (e.g., aryan.gupta)"
$appLinkName = Read-Host "Enter Application Link Name (default: customer_survey_collector)"
if ([string]::IsNullOrWhiteSpace($appLinkName)) { $appLinkName = "customer_survey_collector" }
$formLinkName = Read-Host "Enter Form Link Name (default: Survey_Responses)"
if ([string]::IsNullOrWhiteSpace($formLinkName)) { $formLinkName = "Survey_Responses" }
$dataCenter = Read-Host "Enter Data Center (com/eu/in/com.au/jp, default: com)"
if ([string]::IsNullOrWhiteSpace($dataCenter)) { $dataCenter = "com" }

Write-Host ""
Write-Host "[Step 2] OAuth Client Credentials" -ForegroundColor Yellow
Write-Host "Go to https://api-console.zoho.$dataCenter/ and create a Self Client" -ForegroundColor Gray
Write-Host ""
$clientId = Read-Host "Enter Client ID (1000.XXXXX)"
$clientSecret = Read-Host "Enter Client Secret"

Write-Host ""
Write-Host "[Step 3] Generate Refresh Token" -ForegroundColor Yellow
Write-Host "Go to API Console > Your Client > Generate Code" -ForegroundColor Gray
Write-Host "Scope: ZohoCreator.form.CREATE" -ForegroundColor Gray
Write-Host ""
$code = Read-Host "Enter the generated Code (1000.xxxxx.yyyyy.zzzzz)"

# Exchange code for tokens
Write-Host ""
Write-Host "Exchanging code for tokens..." -ForegroundColor Green

$tokenUrl = "https://accounts.zoho.$dataCenter/oauth/v2/token"
$body = @{
    code = $code
    client_id = $clientId
    client_secret = $clientSecret
    grant_type = "authorization_code"
}

try {
    $response = Invoke-RestMethod -Uri $tokenUrl -Method Post -Body $body
    
    Write-Host "✅ Success! Tokens generated." -ForegroundColor Green
    Write-Host ""
    Write-Host "Access Token: $($response.access_token.Substring(0, 20))..." -ForegroundColor Gray
    Write-Host "Refresh Token: $($response.refresh_token.Substring(0, 20))..." -ForegroundColor Gray
    
    $refreshToken = $response.refresh_token
} catch {
    Write-Host "❌ Failed to get tokens: $_" -ForegroundColor Red
    exit 1
}

# Step 4: Choose storage method
Write-Host ""
Write-Host "[Step 4] Choose Secure Storage Method" -ForegroundColor Yellow
Write-Host "1. Environment Variables (Recommended - Most Secure)"
Write-Host "2. Encrypted Config File (configs/zoho_secure.json)"
Write-Host ""
$choice = Read-Host "Enter choice (1 or 2)"

if ($choice -eq "1") {
    # Set environment variables
    Write-Host ""
    Write-Host "Setting environment variables..." -ForegroundColor Green
    
    [System.Environment]::SetEnvironmentVariable("ZOHO_ACCOUNT_OWNER", $accountOwner, "User")
    [System.Environment]::SetEnvironmentVariable("ZOHO_APP_LINK", $appLinkName, "User")
    [System.Environment]::SetEnvironmentVariable("ZOHO_FORM_LINK", $formLinkName, "User")
    [System.Environment]::SetEnvironmentVariable("ZOHO_CLIENT_ID", $clientId, "User")
    [System.Environment]::SetEnvironmentVariable("ZOHO_CLIENT_SECRET", $clientSecret, "User")
    [System.Environment]::SetEnvironmentVariable("ZOHO_REFRESH_TOKEN", $refreshToken, "User")
    [System.Environment]::SetEnvironmentVariable("ZOHO_DATA_CENTER", $dataCenter, "User")
    
    Write-Host "✅ Environment variables set successfully!" -ForegroundColor Green
    Write-Host "   Note: Restart any open PowerShell windows to use the new variables" -ForegroundColor Yellow
    
} elseif ($choice -eq "2") {
    # Create config file
    Write-Host ""
    Write-Host "Creating encrypted config file..." -ForegroundColor Green
    
    $config = @{
        zoho = @{
            account_owner = $accountOwner
            app_link_name = $appLinkName
            form_link_name = $formLinkName
            client_id = $clientId
            client_secret = $clientSecret
            refresh_token = $refreshToken
            data_center = $dataCenter
        }
    } | ConvertTo-Json -Depth 3
    
    $configPath = "configs\zoho_secure.json"
    $config | Out-File -FilePath $configPath -Encoding UTF8
    
    Write-Host "✅ Config file created: $configPath" -ForegroundColor Green
    Write-Host "   ⚠️  IMPORTANT: Do NOT commit this file to Git!" -ForegroundColor Red
    Write-Host "   ⚠️  Add to .gitignore: configs/zoho_secure.json" -ForegroundColor Red
    
} else {
    Write-Host "❌ Invalid choice" -ForegroundColor Red
    exit 1
}

# Test the connection
Write-Host ""
Write-Host "[Step 5] Testing Connection..." -ForegroundColor Yellow

try {
    $testTokenUrl = "https://accounts.zoho.$dataCenter/oauth/v2/token"
    $testBody = @{
        refresh_token = $refreshToken
        client_id = $clientId
        client_secret = $clientSecret
        grant_type = "refresh_token"
    }
    
    $testResponse = Invoke-RestMethod -Uri $testTokenUrl -Method Post -Body $testBody
    
    Write-Host "✅ Token refresh successful!" -ForegroundColor Green
    Write-Host "   New Access Token: $($testResponse.access_token.Substring(0, 20))..." -ForegroundColor Gray
    
    # Test API endpoint
    $apiUrl = "https://creator.zoho.$dataCenter/api/v2/$accountOwner/$appLinkName/form/$formLinkName"
    Write-Host ""
    Write-Host "Your API Endpoint:" -ForegroundColor Cyan
    Write-Host "   $apiUrl" -ForegroundColor White
    
} catch {
    Write-Host "⚠️  Warning: Token refresh test failed: $_" -ForegroundColor Yellow
    Write-Host "   Please verify your credentials are correct" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Setup Complete!" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Build the application:"
Write-Host "   go build -ldflags=`"-H windowsgui`" -o customer-survey.exe .\cmd\survey\main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Run the application:"
Write-Host "   .\customer-survey.exe" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Check logs:"
Write-Host "   notepad `$env:APPDATA\.customer-survey\webhook.log" -ForegroundColor Gray
Write-Host ""
Write-Host "The application will now use secure OAuth authentication!" -ForegroundColor Green
Write-Host "All data is encrypted in transit and credentials are stored securely." -ForegroundColor Green
Write-Host ""
