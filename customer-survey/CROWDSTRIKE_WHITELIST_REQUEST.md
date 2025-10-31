# Crowdstrike Whitelist Request - ACH Customer Survey Application

## Executive Summary
The ACH Customer Survey Application is a legitimate internal business tool that needs to be whitelisted in Crowdstrike to function properly. Crowdstrike is currently blocking the application execution.

## Application Details

**Application Name:** ACH Customer Survey  
**Purpose:** Internal customer feedback collection and survey tool  
**Type:** Desktop GUI Application  
**Developer:** ACH Real Time Data Services  

## File Information

**Filename:** `customer-survey.exe`  
**Location:** `c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\`  
**File Size:** 6.54 MB  
**SHA256 Hash:** `a402925d065ebad80cbbedefe081d7afe66b9035177ea1fcb4bdfe861b219534`  

## Technical Details

- **Language:** Go (compiled statically)
- **Architecture:** Windows x64
- **Mode:** GUI application (no console window)
- **Network Activity:** Connects only to Zoho Flow webhook for survey submissions
- **Dependencies:** Built-in (no external DLLs required)
- **Source Code:** Available on request

## Request

Please add the following to Crowdstrike whitelist/exclusions:

1. **File Path Exclusion:**
   ```
   c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\customer-survey.exe
   ```

2. **SHA256 Hash Exclusion:**
   ```
   a402925d065ebad80cbbedefe081d7afe66b9035177ea1fcb4bdfe861b219534
   ```

3. **Process Name:**
   ```
   customer-survey.exe
   ```

## Impact of Blocking

Currently, the application is blocked and cannot execute. This is preventing internal user surveys and feedback collection which is critical for service improvement.

## Contacts

**Submitter:** Aryan Gupta  
**Department:** Real Time Data Services  
**Email:** aryan.gupta@myrealdatain.in  

---

## Instructions for IT Team

To whitelist this application in Crowdstrike Falcon:

1. Log into the Crowdstrike Falcon console
2. Navigate to: **Host Setup** → **Response & Isolation** → **Allowlist**
3. Add a new entry with:
   - **Match Value:** File path or SHA256 hash (provided above)
   - **Category:** Application
   - **Description:** "ACH Customer Survey Tool - Internal Business Application"
4. Set appropriate scope/IOA rules

Alternatively, add to local Windows Defender exclusions:
```powershell
Add-MpPreference -ExclusionPath "c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\customer-survey.exe"
```

---

**Note:** This is a safe, internal business application. Crowdstrike is being overly cautious with unsigned executables. Once whitelisted, it should execute without any issues.
