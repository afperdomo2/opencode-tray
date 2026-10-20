; Instalador de OpenCode Tray (Inno Setup 6.3+).
; Instalación POR USUARIO: no requiere administrador.
;
; Compilar:
;   iscc /DMyAppVersion=1.3.0 build\windows\installer\opencode-tray.iss
; o usar scripts\build-installer.ps1.

#ifndef MyAppVersion
  #define MyAppVersion "1.3.0"
#endif

#define MyAppName "OpenCode Tray"
#define MyAppPublisher "Andrés Felipe Perdomo"
#define MyAppURL "https://opencode.ai"
#define MyAppExeName "opencode-tray.exe"
#define WebView2Url "https://go.microsoft.com/fwlink/p/?LinkId=2124703"

[Setup]
AppId={{15946D28-DC9E-4391-8CF2-166E14314FB6}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={localappdata}\Programs\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\..\bin
OutputBaseFilename=OpenCodeTray-Setup-{#MyAppVersion}
SetupIconFile=..\icon.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
CloseApplications=yes
RestartApplications=no
SetupLogging=yes
MinVersion=10.0

[Languages]
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"

[Tasks]
Name: "desktopicon"; Description: "Crear un acceso directo en el escritorio"; GroupDescription: "Accesos directos:"; Flags: unchecked
Name: "launch"; Description: "Ejecutar {#MyAppName}"; GroupDescription: "Al terminar:"; Flags: unchecked

[Files]
Source: "..\..\bin\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Desinstalar {#MyAppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Ejecutar {#MyAppName}"; Flags: nowait postinstall skipifsilent; Tasks: launch

[Code]
const
  WebView2Client = '{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}';

function WebView2Installed(): Boolean;
var
  Version: String;
begin
  Result :=
    RegQueryStringValue(HKLM, 'SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\' + WebView2Client, 'pv', Version) or
    RegQueryStringValue(HKLM, 'SOFTWARE\Microsoft\EdgeUpdate\Clients\' + WebView2Client, 'pv', Version) or
    RegQueryStringValue(HKCU, 'Software\Microsoft\EdgeUpdate\Clients\' + WebView2Client, 'pv', Version);
end;

function OpenCodeOnPath(): Boolean;
var
  ResultCode: Integer;
begin
  Result :=
    Exec(ExpandConstant('{sys}\cmd.exe'), '/C where opencode >nul 2>&1', '', SW_HIDE,
      ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
end;

function EnsureWebView2(): Boolean;
var
  Bootstrapper: String;
  ResultCode: Integer;
begin
  if WebView2Installed() then
  begin
    Result := True;
    Exit;
  end;

  if MsgBox(
    'OpenCode Tray necesita Microsoft Edge WebView2, que no está instalado.' + #13#10 +
    '¿Quieres descargarlo e instalarlo ahora?',
    mbConfirmation, MB_YESNO) = IDNO then
  begin
    MsgBox('Puedes instalarlo manualmente desde:' + #13#10 + '{#WebView2Url}',
      mbInformation, MB_OK);
    Result := False;
    Exit;
  end;

  Bootstrapper := ExpandConstant('{tmp}\MicrosoftEdgeWebview2Setup.exe');
  try
    DownloadTemporaryFile('{#WebView2Url}', 'MicrosoftEdgeWebview2Setup.exe', '', nil);
  except
    MsgBox('No se pudo descargar WebView2. Instálalo manualmente desde:' + #13#10 +
      '{#WebView2Url}', mbError, MB_OK);
    Result := False;
    Exit;
  end;

  if (not Exec(Bootstrapper, '/silent /install', '', SW_SHOW, ewWaitUntilTerminated, ResultCode))
    or (ResultCode <> 0) or (not WebView2Installed()) then
  begin
    MsgBox('No se pudo instalar WebView2. Instálalo manualmente desde:' + #13#10 +
      '{#WebView2Url}', mbError, MB_OK);
    Result := False;
    Exit;
  end;

  Result := True;
end;

function InitializeSetup(): Boolean;
begin
  Result := EnsureWebView2();
  if not Result then
    Exit;

  if not OpenCodeOnPath() then
    MsgBox(
      'No se encontró el comando "opencode" en el PATH.' + #13#10 + #13#10 +
      'OpenCode Tray puede instalarse, pero necesitará la CLI de OpenCode para' + #13#10 +
      'controlar el servicio. Instálala y vuelve a abrir la app.',
      mbInformation, MB_OK);
end;
