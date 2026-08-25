package web

import "embed"

// Files содержит HTML-шаблоны, CSS и JavaScript внутри исполняемого файла.
//
//go:embed templates/*.html static/*.css static/*.js
var Files embed.FS
