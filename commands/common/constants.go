package common

const (
	CLIFlagPort              = "port"
	CLIFlagDebug             = "debug"
	CLIFlagHTML              = "html"
	CLIFlagConfigurationPath = "config"

	ConfigKeyDebug                        = "server.debug"
	ConfigKeyPort                         = "server.port"
	ConfigKeyHTML                         = "server.html"
	ConfigKeyPruneExpiredSessionsInterval = "server.pruneExpiredSessionsInterval"
	ConfigKeyCookieDomain                 = "server.cookieDomain"
	ConfigKeyDatabase                     = "database"
	ConfigKeyLabelFontDirectory           = "labelGeneration.font.directory"
	ConfigKeyLabelFontFamily              = "labelGeneration.font.family"
	ConfigKeyLabelFontFilename            = "labelGeneration.font.filename"
	ConfigKeyLabelBarcodeWidth            = "labelGeneration.barcode.width"
	ConfigKeyLabelBarcodeHeight           = "labelGeneration.barcode.height"
	ConfigKeyLogFile                      = "log.file"
	ConfigKeyLogMaxSize                   = "log.maxSizeMegabytes"
	ConfigKeyLogMaxBackups                = "log.maxBackups"
	ConfigKeyLogMaxAge                    = "log.maxAgeDays"
	ConfigKeyLogCompression               = "log.compression"
)
