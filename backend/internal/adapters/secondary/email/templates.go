package email

import (
	_ "embed"
)

//go:embed templates/provider.html
var providerEmailTemplate string

//go:embed templates/student.html
var studentEmailTemplate string
