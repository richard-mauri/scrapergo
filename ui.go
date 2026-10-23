package main

import "clean-article-reader/internal/ui"

// Compatibility bridge for root tests; UI assets and rendering are owned by internal/ui.
var indexHTML = ui.Template()

func renderedIndexHTML() string { return ui.Render(appVersion,appBuild) }
