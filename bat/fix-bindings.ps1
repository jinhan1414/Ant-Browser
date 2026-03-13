# 修复 wailsjs 绑定文件的 PowerShell 脚本

$modelsFile = "frontend/src/wailsjs/go/models.ts"
$appJsFile = "frontend/src/wailsjs/go/main/App.js"
$appDtsFile = "frontend/src/wailsjs/go/main/App.d.ts"

Write-Host "正在修复 models.ts..."

# 读取 models.ts
$modelsContent = Get-Content $modelsFile -Raw

# 检查是否已包含 PersonalNote
if ($modelsContent -notmatch "export class PersonalNote") {
    Write-Host "添加 PersonalNote 类型..."
    
    # 在 TaskWithChannel 之前插入 PersonalNote 和 AIChatMessage
    $insertContent = @"

	export class PersonalNote {
	    id: string;
	    news_id: string;
	    title: string;
	    content: string;
	    tags: string;
	    created_at: any;
	    updated_at: any;
	    news_title?: string;

	    static createFrom(source: any = {}) {
	        return new PersonalNote(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.news_id = source["news_id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.tags = source["tags"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.news_title = source["news_title"];
	    }
	}

	export class AIChatMessage {
	    id: string;
	    news_id: string;
	    session_id: string;
	    role: string;
	    content: string;
	    created_at: any;

	    static createFrom(source: any = {}) {
	        return new AIChatMessage(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.news_id = source["news_id"];
	        this.session_id = source["session_id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.created_at = source["created_at"];
	    }
	}
"@
    
    $modelsContent = $modelsContent -replace "(export class TaskWithChannel \{)", "$insertContent`r`n`$1"
    Set-Content -Path $modelsFile -Value $modelsContent -Encoding UTF8
}

# 修复 TaskLog 添加 query 字段
if ($modelsContent -match "export class TaskLog" -and $modelsContent -notmatch "query: string;") {
    Write-Host "为 TaskLog 添加 query 字段..."
    $modelsContent = Get-Content $modelsFile -Raw
    $modelsContent = $modelsContent -replace "(message: string;)", "`$1`r`n`t    query: string;"
    $modelsContent = $modelsContent -replace '(this\.message = source\["message"\];)', "`$1`r`n`t        this.query = source[""query""];"
    Set-Content -Path $modelsFile -Value $modelsContent -Encoding UTF8
}

Write-Host "正在修复 App.js..."

# 读取 App.js
$appJsContent = Get-Content $appJsFile -Raw

# 修复 FormatNewsContent 参数
if ($appJsContent -match "export function FormatNewsContent\(arg1\) \{") {
    Write-Host "修复 FormatNewsContent 参数..."
    $appJsContent = $appJsContent -replace "export function FormatNewsContent\(arg1\) \{", "export function FormatNewsContent(arg1, arg2) {"
    $appJsContent = $appJsContent -replace "window\['go'\]\['main'\]\['App'\]\['FormatNewsContent'\]\(arg1\);", "window['go']['main']['App']['FormatNewsContent'](arg1, arg2);"
}

# 修复 RunTaskManually 参数
if ($appJsContent -match "export function RunTaskManually\(arg1\) \{") {
    Write-Host "修复 RunTaskManually 参数..."
    $appJsContent = $appJsContent -replace "export function RunTaskManually\(arg1\) \{", "export function RunTaskManually(arg1, arg2, arg3) {"
    $appJsContent = $appJsContent -replace "window\['go'\]\['main'\]\['App'\]\['RunTaskManually'\]\(arg1\);", "window['go']['main']['App']['RunTaskManually'](arg1, arg2, arg3);"
}

# 添加缺失的 AI 函数
if ($appJsContent -notmatch "export function TestAIConnection") {
    Write-Host "添加缺失的 AI 函数..."
    $aiFunction = @"


export function TestAIConnection(arg1, arg2) {
  return window['go']['main']['App']['TestAIConnection'](arg1, arg2);
}

export function TestAIConfigConnection(arg1) {
  return window['go']['main']['App']['TestAIConfigConnection'](arg1);
}

export function GetAIModels(arg1, arg2) {
  return window['go']['main']['App']['GetAIModels'](arg1, arg2);
}

export function GetAIConfigModels(arg1) {
  return window['go']['main']['App']['GetAIConfigModels'](arg1);
}
"@
    $appJsContent = $appJsContent + $aiFunction
}

Set-Content -Path $appJsFile -Value $appJsContent -Encoding UTF8

Write-Host "正在修复 App.d.ts..."

# 读取 App.d.ts
$appDtsContent = Get-Content $appDtsFile -Raw

# 修复 FormatNewsContent 参数
if ($appDtsContent -match "export function FormatNewsContent\(arg1:string\):Promise<model\.News>;") {
    Write-Host "修复 FormatNewsContent 类型..."
    $appDtsContent = $appDtsContent -replace "export function FormatNewsContent\(arg1:string\):Promise<model\.News>;", "export function FormatNewsContent(arg1:string,arg2?:string):Promise<model.News>;"
}

# 修复 RunTaskManually 参数
if ($appDtsContent -match "export function RunTaskManually\(arg1:string\):Promise<void>;") {
    Write-Host "修复 RunTaskManually 类型..."
    $appDtsContent = $appDtsContent -replace "export function RunTaskManually\(arg1:string\):Promise<void>;", "export function RunTaskManually(arg1:string,arg2?:boolean,arg3?:string):Promise<void>;"
}

# 添加缺失的 AI 函数类型
if ($appDtsContent -notmatch "export function TestAIConnection") {
    Write-Host "添加缺失的 AI 函数类型..."
    $aiFunctionTypes = @"


export function TestAIConnection(arg1:string,arg2:string):Promise<void>;

export function TestAIConfigConnection(arg1:string):Promise<void>;

export function GetAIModels(arg1:string,arg2:string):Promise<Array<string>>;

export function GetAIConfigModels(arg1:string):Promise<Array<string>>;
"@
    $appDtsContent = $appDtsContent + $aiFunctionTypes
}

Set-Content -Path $appDtsFile -Value $appDtsContent -Encoding UTF8

Write-Host "✓ 所有绑定文件已修复"
