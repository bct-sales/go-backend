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
	ConfigKeySwagger                      = "server.swagger"
	ConfigKeyDatabase                     = "database"
	ConfigKeyLabelBarcodeWidth            = "labelGeneration.barcode.width"
	ConfigKeyLabelBarcodeHeight           = "labelGeneration.barcode.height"
	ConfigKeyLogFile                      = "log.file"
	ConfigKeyLogMaxSize                   = "log.maxSizeMegabytes"
	ConfigKeyLogMaxBackups                = "log.maxBackups"
	ConfigKeyLogMaxAge                    = "log.maxAgeDays"
	ConfigKeyLogCompression               = "log.compression"
	ConfigKeyEmailSenderToken             = "email.sender.token"
	ConfigKeyEmailSenderAddress           = "email.sender.address"
	ConfigKeyEmailReceiverAddress         = "email.receiver.address"
)
