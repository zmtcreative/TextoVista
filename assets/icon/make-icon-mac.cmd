@echo off
SETLOCAL ENABLEDELAYEDEXPANSION

:: It is assumed that ImageMagick is in the system path.
:: Download: https://www.imagemagick.org/script/download.php

SET MD_MASTER_SVG=texto-vista-icon.svg
SET MD_ICON_FILENAME=

SET OCD=%CD%
SET SRCDIR=%~dp0
IF !SRCDIR:~-1!==\ SET SRCDIR=!SRCDIR:~0,-1!

IF [%~1] NEQ [] (
    IF EXIST "%~1" SET MD_MASTER_SVG=%~1
)

CALL :SET_OUTPUT_FILENAME "!MD_MASTER_SVG!"

magick -verbose -background none "!SRCDIR!\!MD_MASTER_SVG!" -resize "1024x1024" "!SRCDIR!\!MD_ICON_FILENAME!.png"

ECHO.
ECHO NOTE:
ECHO   You will need to use an online tool to upload the PNG file just created
ECHO   and convert it to an ICNS file for MacOS.
ECHO.
ECHO     Try: https://www.img2icns.com/
ECHO.
ECHO   PNG FILE CREATED: !SRCDIR!\!MD_ICON_FILENAME!.png
ECHO.

GOTO :END

:SET_OUTPUT_FILENAME
    SET MD_ICON_FILENAME=%~n1
GOTO :EOF

:END
EXIT /B 0
