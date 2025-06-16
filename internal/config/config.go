package config

import "github.com/webbben/caius/internal/llm"

// PROJECT - analyzing files, code, etc

// Max number of bytes to analyse for basic file analysis
// Setting this number lower will result in slightly faster file analyses,
// but also lower accuracy if you set it too low.
const MAX_BYTES_BASIC_ANALYSIS int = 1000

// DEBUG CONFIG

const SHOW_FUNCTION_METRICS bool = true
const SHOW_LLM_METRICS bool = true

// LLM MODELS

var BASIC_FILE_ANALYSIS_MODEL = llm.Models.CodeLlama
var DETECT_FILE_TYPE_MODEL = llm.Models.CodeLlama
