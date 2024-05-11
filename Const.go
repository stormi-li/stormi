package stormi

const (
	rollbackwaiting = "rollbackwaiting"
	commitwaiting   = "commitwaiting"
	rollback        = "rollback"
	commit          = "commit"
	noresponse      = "noresponse"
	report          = "report"
	finished        = "finished"
	hi              = "hi"
	stop            = "stop"
	full            = "full"
)

const (
	reset     = "\x1b[0m"
	bold      = "\x1b[1m"
	dim       = "\x1b[2m"
	italic    = "\x1b[3m"
	underline = "\x1b[4m"
	blink     = "\x1b[5m"
	invert    = "\x1b[7m"
	strike    = "\x1b[9m"

	black   = "\x1b[30m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	white   = "\x1b[37m"

	// bgBlack   = "\x1b[40m"
	// bgRed     = "\x1b[41m"
	// bgGreen   = "\x1b[42m"
	// bgYellow  = "\x1b[43m"
	// bgBlue    = "\x1b[44m"
	// bgMagenta = "\x1b[45m"
	// bgCyan    = "\x1b[46m"
	// bgWhite   = "\x1b[47m"
)
