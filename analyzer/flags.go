package analyzer

const defaultAllowedLeadingWords = "create,creates,creating,initialize,initializes,init,configure,configures,setup,setups,start,starts,read,reads,write,writes,send,sends,generate,generates,decode,decodes,encode,encodes,marshal,marshals,unmarshal,unmarshals,apply,applies,process,processes,make,makes,build,builds,test,tests"

var (
	maxDistFlag                 = 5
	includeUnexportedFlag       = true
	includeExportedFlag         = false
	includeTypesFlag            = false
	includeGeneratedFlag        = false
	includeInterfaceMethodsFlag = false
	allowedLeadingWordsFlag     = defaultAllowedLeadingWords
	allowedPrefixesFlag         = ""
	skipPlainWordCamelFlag      = true
	maxCamelChunkInsertFlag     = 2
	maxCamelChunkReplaceFlag    = 2
)

const (
	minDocTokenLen   = 3
	maxChunkDiffSize = 6
)
